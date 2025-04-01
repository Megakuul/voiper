package sip

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/megakuul/voiper/internal/config"
	"github.com/megakuul/voiper/internal/sip/multiplexer/reliable"
	"github.com/megakuul/voiper/internal/sip/transaction/register"
)

// implements some parts of the SIP protocol
// but since anyone does whatever the fuck he wants in the pbx industry anyways
// we can also just call it "fully standardized"
type Client struct {
	rootCtx       context.Context
	rootCtxCancel context.CancelFunc

	rootWg sync.WaitGroup

	lock   sync.Mutex
	config *config.Config
}

type ClientOption func(*Client)

func NewClient(config *config.Config, opts ...ClientOption) *Client {
	rootCtx, rootCtxCancel := context.WithCancel(context.Background())
	client := &Client{
		rootCtx:       rootCtx,
		rootCtxCancel: rootCtxCancel,
		rootWg:        sync.WaitGroup{},
		lock:          sync.Mutex{},
		config:        config,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Register(ctx context.Context) error {
	mx := reliable.New(c.config.Server, reliable.WithLogger(slog.Default()))

	statusChan := make(chan *register.Status, 100)
	go func() {
		for {
			select {
			case status := <-statusChan:
				println(status.Message)
			}
		}
	}()

	output, err := register.Register(ctx, mx, statusChan, &register.Input{
		Secure:      false,
		LocalAddr:   []byte("10.1.10.237"),
		RemoteAddr:  []byte("10.1.10.252"),
		DisplayName: []byte(""),
		Username:    []byte("voiper"),
		Password:    []byte("1234"),
		CallID:      []byte(uuid.New().String()),
		FromTag:     []byte(uuid.New().String()),
		CSeq:        0,
		ExpiresIn:   1 * time.Hour,
	})
	if err != nil {
		return err
	}

	println("DONE")
	println(output.ExpiresIn)

	return nil
}

func (c *Client) Close() {
	c.rootCtxCancel()
	c.rootWg.Wait()
}
