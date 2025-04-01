package via

import (
	"bytes"
	"fmt"
)

// magic cookie specified in 3261.20.42.
// I also need that stuff they smoked when creating this.
const IDIOT_SANDWICH_COOKIE = "z9hG4bK"

type PROTOCOL string

const (
	PROTOCOL_UDP  PROTOCOL = "UDP"
	PROTOCOL_TCP  PROTOCOL = "TCP"
	PROTOCOL_TLS  PROTOCOL = "TLS"
	PROTOCOL_SCTP PROTOCOL = "SCTP"
)

// Via header according to RFC 3261.20.42.
// Host includes the optional 'Port' (e.g. Host: "server.local:5060").
// Between 'Host' and 'Port', there is no whitespace trimming. If someone puts a whitespace there,
// I will personally revoke their license to ever touch a keyboard again.
type Header struct {
	Version  []byte
	Protocol PROTOCOL
	Host     []byte
	Params   map[string][]byte
}

func Serialize(header *Header) []byte {
	b := bytes.Buffer{}
	b.Write(header.Version)
	b.WriteString("/")
	b.WriteString(string(header.Protocol))
	b.WriteString(" ")
	b.Write(header.Host)

	for key, value := range header.Params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.Write(value)
	}

	return b.Bytes()
}

func Parse(input []byte) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string][]byte{},
	}

	block := bytes.SplitN(bytes.TrimSpace(input), []byte("/"), 3)
	if len(block) != 3 {
		return nil, fmt.Errorf("invalid via header: expected 'SIP/<VERSION>/<PROTOCOL> ...' got '%s'", string(input))
	}
	header.Version = append(bytes.TrimSpace(block[0]), byte('/'))
	header.Version = append(header.Version, bytes.TrimSpace(block[1])...)

	data := bytes.SplitN(bytes.TrimSpace(block[2]), []byte(" "), 2)
	if len(data) != 2 {
		return nil, fmt.Errorf("invalid via header: expected '.../<PROTOCOL> <host>:<port>;<params...>' got '.../%s'", string(block[2]))
	}
	header.Protocol = PROTOCOL(data[0])

	rhs := bytes.Split(bytes.TrimSpace(data[1]), []byte(";"))
	for i, params := range rhs {
		if i == 0 {
			if len(params) == 0 {
				return nil, fmt.Errorf(
					"invalid via header rhs: expected '... <host>[:<port>];<params...>' got '... %s;<params...>'", string(params),
				)
			}
			header.Host = bytes.TrimSpace(params)
			continue
		}

		kv := bytes.SplitN(params, []byte("="), 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf(
				"invalid via header param: expected ';<key>=<value>' got ';%s'", string(params),
			)
		}

		header.Params[string(kv[0])] = kv[1]
	}

	return header, nil
}
