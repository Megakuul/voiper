package phone

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/megakuul/voiper/pkg/sip"
)

func (m *Manager) Suspend() {
	m.mu.Lock()
	calls := make([]*call, 0, len(m.calls))
	for id, c := range m.calls {
		calls = append(calls, c)
		delete(m.calls, id)
	}
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(m.ctx, time.Second)
	defer cancel()
	var signaling sync.WaitGroup
	for _, c := range calls {
		signaling.Add(1)
		go func(c *call) {
			defer signaling.Done()
			if c.owner == nil || c.owner.client == nil {
				return
			}
			m.mu.Lock()
			ringing := c.state.State == "ringing"
			id := c.state.ID
			m.mu.Unlock()
			if ringing {
				c.owner.client.Reject(ctx, id)
			} else {
				c.owner.client.Hangup(ctx, id)
			}
			// Sleep ends local ownership even if the peer could not acknowledge cleanup.
			c.owner.client.AbortCall(id)
		}(c)
	}
	signaling.Wait()
	for _, c := range calls {
		m.finish(c, "ended for system sleep")
	}
	m.notify()
}
func (m *Manager) registerAccount(a *account) error {
	err := a.client.Register(a.ctx)
	m.mu.Lock()
	a.registrationError = err
	if err != nil && a.ctx.Err() == nil && m.accounts[a.state.Name] == a {
		a.state.State = "failed"
		a.state.Error = err.Error()
	}
	m.mu.Unlock()
	if err == nil {
		select {
		case a.publicationWake <- struct{}{}:
		default:
		}
	}
	m.notify()
	return err
}

// RetryRegistration reuses the account and its dialogs; enabling it again would end calls.
func (m *Manager) RetryRegistration(name string) error {
	if !m.recovery.TryLock() {
		return errors.New("registration recovery is already in progress")
	}
	defer m.recovery.Unlock()
	m.mu.Lock()
	a := m.accounts[name]
	m.mu.Unlock()
	if a == nil {
		return errors.New("account is not enabled")
	}
	select {
	case <-a.registrationDone:
	default:
		return errors.New("registration is already in progress")
	}
	m.mu.Lock()
	initialError := a.registrationError
	m.mu.Unlock()
	if initialError != nil {
		return m.registerAccount(a)
	}
	ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
	defer cancel()
	return a.client.RefreshRegistration(ctx)
}

func (m *Manager) RecoverRegistrations() {
	m.recovery.Lock()
	defer m.recovery.Unlock()
	m.mu.Lock()
	accounts := make([]*account, 0, len(m.accounts))
	for _, a := range m.accounts {
		accounts = append(accounts, a)
	}
	m.mu.Unlock()
	for _, a := range accounts {
		if m.ctx.Err() != nil {
			return
		}
		if a.registrationDone != nil {
			select {
			case <-a.registrationDone:
			case <-a.ctx.Done():
				continue
			}
		}
		m.mu.Lock()
		initialError := a.registrationError
		m.mu.Unlock()
		if initialError != nil {
			if sip.IsTransientRegistrationError(initialError) {
				m.registerAccount(a)
			}
			continue
		}
		ctx, cancel := context.WithTimeout(a.ctx, 15*time.Second)
		err := a.client.RefreshRegistration(ctx)
		cancel()
		if err != nil {
			m.report(err)
		} else {
			m.report(a.client.RefreshSubscriptions(a.ctx))
			select {
			case a.publicationWake <- struct{}{}:
			default:
			}
		}
	}
}

func (m *Manager) RetryAudio(id string) error {
	m.mu.Lock()
	c := m.calls[id]
	if c == nil || c.media == nil {
		m.mu.Unlock()
		return errors.New("call media is not connected")
	}
	stream := c.media
	m.mu.Unlock()
	return stream.RetryAudio()
}

func (m *Manager) RestartMediaPath(id string) error {
	m.mu.Lock()
	c := m.calls[id]
	if c == nil || c.state.State != "connected" || !c.mediaConnected || c.media == nil {
		m.mu.Unlock()
		return errors.New("call media is not connected")
	}
	stream := c.media
	m.mu.Unlock()
	ctx, cancel := context.WithTimeout(m.ctx, 15*time.Second)
	defer cancel()
	var offer []byte
	err := c.owner.client.ReinviteWithOffer(ctx, id, func() ([]byte, error) {
		next, err := stream.RestartOffer(ctx, advertisedIP(c.owner.config))
		if err == nil {
			offer = next
		}
		return next, err
	})
	if err != nil && len(offer) > 0 {
		stream.CancelRestart(offer)
	}
	m.notify()
	return err
}
