package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/emiago/sipgo"
	"github.com/emiago/sipgo/sip"
	"github.com/icholy/digest"
	"github.com/megakuul/voiper/internal/config"
)

type Client struct {
	clientLock   sync.Mutex
	clientConfig *config.Config
	client       *sipgo.Client
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	client := &Client{
		client: nil,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) register(ctx context.Context) error {
	c.clientLock.Lock()
	defer c.clientLock.Unlock()
	if c.client != nil {
		return nil
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	agent, err := sipgo.NewUA()
	if err != nil {
		return err
	}
	client, err := sipgo.NewClient(agent)
	if err != nil {
		return err
	}

	req := sip.NewRequest(sip.REGISTER, sip.Uri{
		Scheme: "sip",
		Host:   "10.1.10.252",
	})

	// identity
	req.AppendHeader(&sip.FromHeader{
		DisplayName: "Voiper",
		Address: sip.Uri{
			Scheme:   "sip",
			User:     "voiper",
			Password: "password",
			Host:     "10.1.10.252",
			Port:     5060,
		},
	})

	// which user do i want to register
	req.AppendHeader(&sip.ToHeader{
		DisplayName: "Voiper",
		Address: sip.Uri{
			Scheme: "sip",
			User:   "voiper",
			Host:   "10.1.10.252",
			Port:   5060,
		},
	})

	// not required because we only do one way transactional requests
	// still include it because sip is a fucking legacy crapprotocol
	req.AppendHeader(&sip.ContactHeader{
		Address: sip.Uri{
			User: "voiper",
			Host: "10.1.10.237",
			Port: 5060,
		},
	})

	tx, err := client.TransactionRequest(ctx, req, sipgo.ClientRequestRegisterBuild)
	if err != nil {
		return err
	}
	defer tx.Terminate()

	for {
		select {
		case res, ok := <-tx.Responses():
			if !ok {
				return fmt.Errorf("transaction closed")
			}
			if res.StatusCode == 401 {
				// Get WwW-Authenticate
				wwwAuth := res.GetHeader("WWW-Authenticate")
				chal, err := digest.ParseChallenge(wwwAuth.Value())
				if err != nil {
					println("Fail to parse challenge", "error", err, "wwwauth", wwwAuth.Value)
					return err
				}

				// Reply with digest
				cred, _ := digest.Digest(chal, digest.Options{
					Method:   req.Method.String(),
					URI:      "10.1.10.252",
					Username: "voiper",
					Password: "password",
				})

				newReq := req.Clone()
				newReq.RemoveHeader("Via") // Must be regenerated by tranport layer
				newReq.AppendHeader(sip.NewHeader("Authorization", cred.String()))

				ctx := context.Background()
				tx, err := client.TransactionRequest(ctx, newReq, sipgo.ClientRequestAddVia)
				if err != nil {
					println("Fail to create transaction", "error", err)
					return err
				}
				defer tx.Terminate()

				res, err = getResponse(tx)
				if err != nil {
					println("Fail to get response", "error", err)
					return err
				}

			}
			println(res.StatusCode)
			println(res.Reason)
		case <-ctx.Done():
			return fmt.Errorf("context exceeded")
		}
	}
}

func getResponse(tx sip.ClientTransaction) (*sip.Response, error) {
	for {
		select {
		case <-tx.Done():
			return nil, fmt.Errorf("transaction died")
		case res := <-tx.Responses():
			if res.StatusCode == 100 {
				continue
			}
			if res.StatusCode != 200 {
				return nil, fmt.Errorf("ALARM i have status code ", res.StatusCode, " reason: ", res.Reason)
			}
			println("SUCCESSFULLY REGISTERED")
			return res, nil
		}
	}

}
