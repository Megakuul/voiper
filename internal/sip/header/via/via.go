package via

import (
	"fmt"
	"strings"
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
	Version  string
	Protocol PROTOCOL
	Host     string
	Params   map[string]string
}

// Serializes the Via header into a string (key is not included).
func Serialize(header *Header) string {
	b := strings.Builder{}
	b.WriteString(header.Version)
	b.WriteString("/")
	b.WriteString(string(header.Protocol))
	b.WriteString(" ")
	b.WriteString(header.Host)

	for key, value := range header.Params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(value)
	}

	return b.String()
}

// Parses the Via header string
func Parse(str string) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string]string{},
	}

	block := strings.SplitN(str, "/", 3)
	if len(block) != 3 {
		return nil, fmt.Errorf("invalid via header: expected 'SIP/<VERSION>/<PROTOCOL> ...' got '%s'", str)
	}
	header.Version = strings.TrimSpace(block[0]) + "/" + strings.TrimSpace(block[1])

	data := strings.SplitN(strings.TrimSpace(block[2]), " ", 2)
	if len(data) != 2 {
		return nil, fmt.Errorf("invalid via header: expected '.../<PROTOCOL> <host>:<port>;<params...>' got '.../%s'", block[2])
	}
	header.Protocol = PROTOCOL(data[0])

	rhs := strings.Split(strings.TrimSpace(data[1]), ";")
	for i, params := range rhs {
		if i == 0 {
			if params == "" {
				return nil, fmt.Errorf(
					"invalid via header rhs: expected '... <host>[:<port>];<params...>' got '... %s;<params...>'", params,
				)
			}
			header.Host = strings.TrimSpace(params)
			continue
		}

		kv := strings.SplitN(params, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf(
				"invalid via header param: expected ';<key>=<value>' got ';%s'", params,
			)
		}

		header.Params[kv[0]] = kv[1]
	}

	return header, nil
}
