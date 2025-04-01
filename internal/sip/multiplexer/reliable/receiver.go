package reliable

import (
	"fmt"
	"io"
	"net"

	"github.com/megakuul/voiper/internal/sip/response"
	"golang.org/x/sync/errgroup"
)

// ensureReceiver starts a transaction receiver if not already running. Don't get confused by the term "receiver" here.
// It is just a "receiver" from the SIP transaction perspective. The underlying udp connection does both read and write.
// It creates a net.Listener that listens for SIP requests which are then handled and answered by a series of responses.
func (m *Multiplexer) ensureReceiver() error {
	m.receiverLock.Lock()
	defer m.receiverLock.Unlock()
	if m.receiverState {
		return nil
	}
	m.logger.Info(fmt.Sprintf("no udp listener running: starting new socket on %s", m.localAddr))

	listener, err := net.Listen("tcp", m.localAddr)
	if err != nil {
		return err
	}

	m.operationWg.Add(1)
	go func() {
		defer m.operationWg.Done()
		<-m.rootCtx.Done()
		if err := listener.Close(); err != nil {
			m.logger.Error(fmt.Sprintf("failed to close listener: %v", err))
		}
		m.receiverLock.Lock()
		m.receiverState = false
		m.receiverLock.Unlock()
	}()

	m.operationWg.Add(1)
	go func() {
		defer m.operationWg.Done()
		for {
			m.accept(listener)
		}
	}()

	m.receiverState = true

	return nil
}

// accept wraps the logic to accept and handle an incomming tcp connection.
func (m *Multiplexer) accept(listener net.Listener) {
	conn, err := listener.Accept()
	if err != nil {
		m.logger.Error(fmt.Sprintf("accept failure: %v", err))
		return
	}
	defer conn.Close()

	responseChan := make(chan *response.Response, QUEUE_SIZE)
	g, ctx := errgroup.WithContext(m.rootCtx)
	g.Go(func() error {
		for {
			select {
			case res := <-responseChan:
				_, err := io.Copy(conn, response.Serialize(res))
				if err != nil {
					return fmt.Errorf("write failure: %w", err)
				}
			case <-ctx.Done():
				return nil
			}
		}
	})

	g.Go(func() error {
		return m.multiplex(ctx, conn, responseChan)
	})

	err = g.Wait()
	if err != nil {
		m.logger.Warn(err.Error())
	}
}
