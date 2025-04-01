// unreliable provides a udp client that opens a udp connection to a sip endpoint and multiplexes transactions over it.
// The multiplexer starts the underlying listener and client lazily when StartListen() / StartCall() is called.
// Transactions started by StartCall() are tracked locally, responses are read from the udp client and provided in the channel acquired from StartCall().
// Transactions started by the remote peer flow via the listener and are not tracked locally instead they are handled by the callback defined in StartListen().
package unreliable

import (
	"context"
	"fmt"
	"log/slog"
	"sync"

	"github.com/megakuul/voiper/internal/sip/header/via"
	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/response"
)

const (
	QUEUE_SIZE       = 100  // size of internal request / response queues
	PACKET_SIZE      = 1300 // udp packet size (according to 3261.18.1.1)
	MAX_READ_ERRORS  = 3    // number of consecutive read errors before closing the socket
	MAX_WRITE_ERRORS = 3    // number of consecutive write errors before closing the socket
)

type ListenerFunc func(context.Context, *request.Request, chan *response.Response) error

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

	requestChan  chan *request.Request   // queue for requests that should be sent to remoteAddr
	responseChan chan *response.Response // queue for responses that should be sent to remoteAddr

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
		responseChan:     make(chan *response.Response, QUEUE_SIZE),
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

func WithLogger(logger *slog.Logger) Option {
	return func(m *Multiplexer) {
		m.logger = logger
	}
}

func (m *Multiplexer) Protocol() via.PROTOCOL {
	return via.PROTOCOL_UDP
}

// StartCall starts a sip transaction (identified via branch) by sending the specified request to the server.
// Returns a channel that provides all responses to the transaction.
func (m *Multiplexer) StartCall(branch string, req *request.Request) (<-chan *response.Response, error) {
	m.operationLock.RLock()
	defer m.operationLock.RUnlock()
	if !m.operationState {
		return nil, fmt.Errorf("multiplexer is already closed")
	}
	err := m.ensureSender()
	if err != nil {
		return nil, err
	}

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
	err := m.ensureReceiver()
	if err != nil {
		return err
	}

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
