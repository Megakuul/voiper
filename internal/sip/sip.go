package sip

import (
	"fmt"
	"net"
	"sync"

	"github.com/google/uuid"
	"github.com/megakuul/voiper/internal/config"
	"github.com/megakuul/voiper/internal/sip/header/via"
	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/uri"
)

// implements some parts of the SIP protocol
// but since anyone does whatever the fuck he wants in the fucking pbx industry anyways
// we can also just call it "fully standardized"
type Client struct {
	lock   sync.Mutex
	config *config.Config
}

type ClientOption func(*Client)

func NewClient(config *config.Config, opts ...ClientOption) *Client {
	client := &Client{
		config: config,
	}

	for _, opt := range opts {
		opt(client)
	}

	return client
}

func (c *Client) Register() error {
	host := fmt.Sprintf("%s:%d", c.config.Server, c.config.Port)

	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}

	_, err = conn.Write([]byte(
		request.SerializeRequest(&request.Request{
			Method: request.REGISTER,
			URI: uri.URI{
				Secure: false,
				User:   c.config.Username,
				Host:   host,
			},
			Version: "SIP/2.0",
			Headers: map[string][]string{
				"via": {via.Serialize(&via.Header{
					Version:  "SIP/2.0",
					Protocol: via.PROTOCOL_TCP,
					Host:     host,
					Params:   map[string]string{"branch": via.IDIOT_SANDWICH_COOKIE + uuid.New().String()}, // branch is transaction unique
				})},
				"to":      {fmt.Sprintf("Voiper <sip:voiper@%s>", host)},
				"from":    {fmt.Sprintf("Voiper <sip:voiper@%s>", host)},
				"call-id": {fmt.Sprintf("%s@%s", uuid.New().String(), host)}, // call-id is dialog unique
				"cseq":    {fmt.Sprintf("1 REGISTER")},                       // cseq is just a dialog tracker that is used over the dialog and incremented on each request
			},
			Body: "", //empty.NewBody(),
		}),
	))
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
