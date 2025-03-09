package sip

import (
	"fmt"
	"net"
	"sync"

	"github.com/megakuul/voiper/internal/config"
)

// implements some parts of the SIP protocol
// but since anyone does whatever the fuck he wants in the fucking pbx industry anyways
// we can also just call it "fully standardized"
type Client struct {
	lock   sync.Mutex
	config *config.Config
}

type ClientOption func(*Client)

func NewClient(opts ...ClientOption) *Client {
	client := &Client{}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Register() error {
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", c.config.Server, c.config.Port))
	if err != nil {
		return err
	}

	conn.Write([]byte(
		"",
	))
}
