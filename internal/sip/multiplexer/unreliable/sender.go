package unreliable

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"strings"

	"github.com/megakuul/voiper/internal/sip/header/via"
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
	m.logger.Info(fmt.Sprintf("no udp client running: starting new socket to %s", m.remoteAddr))

	conn, err := net.Dial("udp", m.remoteAddr)
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
			}
		}
	}()
	m.operationWg.Add(1)
	go func() {
		defer m.operationWg.Done()
		errCounter := 0
		buffer := make([]byte, PACKET_SIZE)
		for {
			n, err := conn.Read(buffer)
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
			res, err := response.Parse(bytes.NewReader(buffer[:n]))
			if err != nil {
				m.logger.Warn(err.Error())
				continue
			}

			viaValues, ok := res.Headers["via"]
			if !ok || len(viaValues) < 1 {
				m.logger.Warn(fmt.Sprintf("response with status '%s' is missing the 'via' header; discarding...", string(res.Status)))
				continue
			}

			viaHeader, err := via.Parse(viaValues[0])
			if err != nil {
				m.logger.Warn(fmt.Sprintf("%v; discarding...", err))
				continue
			}

			branch, ok := viaHeader.Params["branch"]
			branchStr := string(branch)
			if !ok || !strings.HasPrefix(branchStr, via.IDIOT_SANDWICH_COOKIE) {
				m.logger.Warn("via header is missing a valid branch parameter; discarding...")
				continue
			}

			m.transactionsLock.RLock()
			trChan, ok := m.transactions[branchStr]
			m.transactionsLock.RUnlock()
			if ok {
				trChan <- res
			}
		}
	}()

	m.senderState = true

	return nil
}
