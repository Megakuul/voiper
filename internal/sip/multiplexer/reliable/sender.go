package reliable

import (
	"context"
	"fmt"
	"io"
	"net"

	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/response"
)

// ensureSender starts a transaction sender if not already running. Don't get confused by the term "sender" here.
// It is just a "sender" from the SIP transaction perspective. The underlying udp connection does both read and write.
// It creates a net.Conn that sends application generated requests and receives a series of responses.
func (m *Multiplexer) ensureSender() error {
	m.senderLock.Lock()
	defer m.senderLock.Unlock()
	if m.senderState {
		return nil
	}
	m.logger.Info(fmt.Sprintf("no tcp client running: starting new socket to %s", m.remoteAddr))

	responseChan := make(chan *response.Response, QUEUE_SIZE)
	conn, err := net.Dial("tcp", m.remoteAddr)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithCancel(m.rootCtx)
	m.operationWg.Add(1)
	go func() {
		defer m.operationWg.Done()
		errCounter := 0
		for {
			select {
			case <-ctx.Done():
				if err := conn.Close(); err != nil {
					m.logger.Error(fmt.Sprintf("failed to close socket: %v", err))
				}
				m.senderLock.Lock()
				m.senderState = false
				m.senderLock.Unlock()
				return
			case req := <-m.requestChan:
				_, err := io.Copy(conn, request.Serialize(req))
				if err != nil {
					errCounter++
					if errCounter >= MAX_WRITE_ERRORS {
						m.logger.Error(fmt.Sprintf(
							"write failure (%d/%d): shutting down socket...", errCounter, MAX_WRITE_ERRORS,
						))
						cancel()
						return
					}
					m.logger.Warn(fmt.Sprintf("write failure (%d/%d): %v", errCounter, MAX_WRITE_ERRORS, err))
					continue
				}
				errCounter = 0
			case res := <-responseChan:
				_, err := io.Copy(conn, response.Serialize(res))
				if err != nil {
					errCounter++
					if errCounter >= MAX_WRITE_ERRORS {
						m.logger.Error(fmt.Sprintf(
							"write failure (%d/%d): shutting down socket...", errCounter, MAX_WRITE_ERRORS,
						))
						cancel()
						return
					}
					m.logger.Warn(fmt.Sprintf("write failure (%d/%d): %v", errCounter, MAX_WRITE_ERRORS, err))
					continue
				}
				errCounter = 0
			}
		}
	}()
	m.operationWg.Add(1)
	go func() {
		defer m.operationWg.Done()
		for {
			m.multiplex(ctx, conn, responseChan)
		}
	}()

	return nil
}
