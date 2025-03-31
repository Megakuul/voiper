package sip

import (
	"context"
	"fmt"
	"net"
	"sync"

	"github.com/megakuul/voiper/internal/config"
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
	host := fmt.Sprintf("%s:%d", c.config.Server, c.config.Port)

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}

	if err != nil {
		return err
	}

	for i := 0; i < 10; i++ {
		buffer := make([]byte, 1024)
		n, err := conn.Read(buffer)
		if err != nil {
			return err
		}
		println(string(buffer[:n]))
	}

	return nil
}

func (c *Client) Close() {
	c.rootCtxCancel()
	c.rootWg.Wait()
}
