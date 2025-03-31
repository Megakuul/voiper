// reliable provides a tcp client that opens a tcp stream to a sip endpoint and multiplexes transactions over it.
// Unlike the udp multiplexer, the tcp multiplexer is more complex: In udp there are two simple transaction directions:
// incoming (handled by the receiver) and outgoing (handled by the sender), tcp, on the other hand, can receive incoming and outgoing transactions
// over established sender or receiver connections. This means that both the receiver and the sender can receive and send transactions.
// However, the initial transaction that starts the receiver / sender is always receiving / sending.
// -> This basically allows reusing existing tcp connections for new transactions as defined in rfc5626.
// Confused? Me too. I love SIP...
package reliable

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"
	"net"
	"strings"
	"sync"

	"github.com/megakuul/voiper/internal/sip/header/contentlength"
	"github.com/megakuul/voiper/internal/sip/header/via"
	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/response"
)

const (
	QUEUE_SIZE       = 100 // size of internal request / response queues
	MAX_READ_ERRORS  = 3   // number of consecutive read errors before closing the socket
	MAX_WRITE_ERRORS = 3   // number of consecutive write errors before closing the socket
)

type Multiplexer struct {
	rootCtx       context.Context
	rootCtxCancel context.CancelFunc

	operationLock  sync.RWMutex   // lock used to synchronize operationState
	operationState bool           // determines if more operations can take action (essentially: !closed)
	operationWg    sync.WaitGroup // waitgroup that tracks all async operations of the multiplexer

	logger *slog.Logger

	remoteAddr string // udp addr of the server (for outgoing requests)
	localAddr  string // local udp listener address (for incomming requests)

	senderLock  sync.Mutex
	senderState bool // determines if a transaction sender is running (used for dynamic socket creation)

	receiverLock  sync.Mutex
	receiverState bool // determines if a transaction receiver is running (used for dynamic socket creation)

	requestChan chan *request.Request // queue for requests that should be sent to remoteAddr

	listenersLock sync.RWMutex
	listeners     map[string]func(context.Context, *request.Request, chan *response.Response) error // handler functions used to handle incomming transactions (key: method; value: callback)

	transactionsLock sync.RWMutex
	transactions     map[string]chan *response.Response // active transactions that were initiated by the local peer (key: branch; value: local response queue)
}

type Option func(*Multiplexer)

func New(remoteAddr string, opts ...Option) *Multiplexer {
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())
	multiplexer := &Multiplexer{
		rootCtx:          rootCtx,
		rootCtxCancel:    rootCtxCancel,
		operationLock:    sync.RWMutex{},
		operationState:   true,
		operationWg:      sync.WaitGroup{},
		logger:           slog.Default(),
		remoteAddr:       remoteAddr,
		localAddr:        "0.0.0.0:5060",
		senderLock:       sync.Mutex{},
		senderState:      false,
		receiverLock:     sync.Mutex{},
		receiverState:    false,
		requestChan:      make(chan *request.Request, QUEUE_SIZE),
		listenersLock:    sync.RWMutex{},
		listeners:        map[string]func(context.Context, *request.Request, chan *response.Response) error{},
		transactionsLock: sync.RWMutex{},
		transactions:     map[string]chan *response.Response{},
	}

	for _, opt := range opts {
		opt(multiplexer)
	}
	return multiplexer
}

// StartCall starts a sip transaction (identified via branch) by sending the specified request to the server.
// Returns a channel that provides all responses to the transaction.
func (m *Multiplexer) StartCall(branch string, req *request.Request) (<-chan *response.Response, error) {
	m.operationLock.RLock()
	defer m.operationLock.RUnlock()
	if !m.operationState {
		return nil, fmt.Errorf("multiplexer is already closed")
	}
	m.ensureSender()

	m.transactionsLock.Lock()
	defer m.transactionsLock.Unlock()
	if transaction, ok := m.transactions[branch]; ok {
		close(transaction)
	}
	m.transactions[branch] = make(chan *response.Response, QUEUE_SIZE)
	m.requestChan <- req
	return m.transactions[branch], nil
}

// StopCall stops a sip transaction (identified via branch).
// This closes the response channel of the transaction.
func (m *Multiplexer) StopCall(branch string) {
	m.transactionsLock.Lock()
	defer m.transactionsLock.Unlock()

	if transaction, ok := m.transactions[branch]; ok {
		close(transaction)
		delete(m.transactions, branch)
	}
}

// StartListen starts listening for incomming transaction requests with the specified method.
// Executes the provided callback for each new incomming transaction.
// The callback must immediately exit if the context is cancelled.
func (m *Multiplexer) StartListen(method string, callback func(context.Context, *request.Request, chan *response.Response) error) error {
	m.operationLock.RLock()
	defer m.operationLock.RUnlock()
	if !m.operationState {
		return fmt.Errorf("multiplexer is already closed")
	}
	m.ensureReceiver()

	m.listenersLock.Lock()
	defer m.listenersLock.Unlock()

	m.listeners[method] = callback
	return nil
}

// StopListen stops a listener, this does not stop transaction callbacks,
// instead it just stops further callbacks with the specified method to be initiated.
func (m *Multiplexer) StopListen(method string) {
	m.listenersLock.Lock()
	defer m.listenersLock.Unlock()

	delete(m.listeners, method)
}

// Shutdown closes the multiplexer, this includes stopping all operating
// connections and listener callbacks. After shutdown, no new calls or listeners can be started.
func (m *Multiplexer) Shutdown() {
	m.operationLock.Lock()
	defer m.operationLock.Unlock()
	if !m.operationState {
		return
	}
	m.operationState = false

	m.rootCtxCancel()

	m.operationWg.Wait()

	m.transactionsLock.Lock()
	for _, responses := range m.transactions {
		close(responses)
	}
	m.transactionsLock.Unlock()
}

// multiplex reads from a connection and handles the incomming data appropriate.
// Incomming requests / responses are handled in a blocking manner to avoid corruption of the stream.
// Returns an error on critical failures like the corruption of the stream.
func (m *Multiplexer) multiplex(ctx context.Context, conn net.Conn, responseChan chan *response.Response) error {
	ok, reader := request.Peek(conn)
	if ok {
		req, err := request.Parse(reader)
		if err != nil {
			return err
		}

		clValues, ok := req.Headers["content-length"]
		if !ok || len(clValues) < 1 {
			return fmt.Errorf("request with method '%s' is missing the 'content-length' header; aborting...", string(req.Method))
		}

		clHeader, err := contentlength.Parse(string(clValues[0]))
		if err != nil {
			return fmt.Errorf("%v; aborting...", err)
		}

		// copy the body into memory to avoid blocking the streamreader
		// this is especially important because the request callback may be dependent on a user action
		// (e.g. the user must take up the phone, as long as he does not, the whole stream would be blocked)
		buffer := make([]byte, clHeader.Length)
		n, err := req.Body.Read(buffer)
		if err != nil {
			return fmt.Errorf("read failure: %v", err)
		}
		req.Body = bytes.NewBuffer(buffer[:n])

		m.listenersLock.RLock()
		callback, ok := m.listeners[string(req.Method)]
		m.listenersLock.RUnlock()
		if ok {
			m.operationWg.Add(1)
			go func() {
				defer m.operationWg.Done()
				err = callback(ctx, req, responseChan)
				if err != nil {
					m.logger.Warn(err.Error())
				}
			}()
		}
		return nil
	}

	ok, reader = response.Peek(reader)
	if ok {
		res, err := response.Parse(reader)
		if err != nil {
			return err
		}

		viaValues, ok := res.Headers["via"]
		if !ok || len(viaValues) < 1 {
			return fmt.Errorf("response with status '%s' is missing the 'via' header; aborting...", string(res.Status))
		}

		viaHeader, err := via.Parse(string(viaValues[0]))
		if err != nil {
			return fmt.Errorf("%v; aborting...", err)
		}

		branch, ok := viaHeader.Params["branch"]
		if !ok || !strings.HasPrefix(branch, via.IDIOT_SANDWICH_COOKIE) {
			return fmt.Errorf("via header is missing a valid branch parameter; discarding...")
		}

		clValues, ok := res.Headers["content-length"]
		if !ok || len(clValues) < 1 {
			return fmt.Errorf("response with status '%s' is missing the 'content-length' header; aborting...", string(res.Status))
		}

		clHeader, err := contentlength.Parse(string(clValues[0]))
		if err != nil {
			return fmt.Errorf("%v; aborting...", err)
		}

		buffer := make([]byte, clHeader.Length)
		n, err := res.Body.Read(buffer)
		if err != nil {
			return fmt.Errorf("read failure: %v", err)
		}
		res.Body = bytes.NewBuffer(buffer[:n])

		m.transactionsLock.RLock()
		trChan, ok := m.transactions[branch]
		m.transactionsLock.RUnlock()
		if ok {
			trChan <- res
		}
	}

	return fmt.Errorf("encountered unknown stream data: expected request or response head")
}
