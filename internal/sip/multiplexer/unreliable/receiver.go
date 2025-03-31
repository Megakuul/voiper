package unreliable

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"

	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/response"
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

	conn, err := net.ListenPacket("udp", m.localAddr)
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
			case <-m.rootCtx.Done():
				if err := conn.Close(); err != nil {
					m.logger.Error(fmt.Sprintf("failed to close socket: %v", err))
				}
				m.receiverLock.Lock()
				m.receiverState = false
				m.receiverLock.Unlock()
				return
			case res := <-m.responseChan:
				addr, err := net.ResolveUDPAddr("udp", m.remoteAddr)
				if err != nil {
					m.logger.Warn(fmt.Sprintf("failed to resolve remote addr: %v", err))
					continue
				}
				resBuffer, _ := io.ReadAll(response.Serialize(res))
				_, err = conn.WriteTo(resBuffer, addr)
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
		errCounter := 0
		buffer := make([]byte, PACKET_SIZE)
		for {
			n, _, err := conn.ReadFrom(buffer)
			if err != nil {
				errCounter++
				if errCounter >= MAX_READ_ERRORS {
					m.logger.Error(fmt.Sprintf(
						"read failure (%d/%d): shutting down socket...", errCounter, MAX_READ_ERRORS,
					))
					cancel()
					return
				}
				m.logger.Warn(fmt.Sprintf("read failure (%d/%d): %v", errCounter, MAX_READ_ERRORS, err))
				continue
			}
			errCounter = 0
			req, err := request.Parse(bytes.NewReader(buffer[:n]))
			if err != nil {
				m.logger.Warn(err.Error())
				continue
			}

			m.listenersLock.RLock()
			callback, ok := m.listeners[string(req.Method)]
			m.listenersLock.RUnlock()
			if ok {
				m.operationWg.Add(1)
				go func() {
					defer m.operationWg.Done()
					callback(ctx, req, m.responseChan)
				}()
			}
		}
	}()

	return nil
}
