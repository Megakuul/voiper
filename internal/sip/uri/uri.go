package uri

import (
	"fmt"
	"strings"
)

// SIP uri according to RFC 3261.19.1.
// Host includes the optional 'Port' (e.g. Host: "server.local:5060").
// Ignores 'password' segment because there is no logical reason to ever use this.
// (they even say it should not be used.. IN THEIR OWN SPEC WTF?!)
type URI struct {
	Secure bool
	User   string
	Host   string
	params map[string]string
}

// Serializes a sip uri (RFC 3261.19.1).
func Serialize(uri *URI) string {
	b := strings.Builder{}
	if uri.Secure {
		b.WriteString("sips:")
	} else {
		b.WriteString("sip:")
	}

	b.WriteString(uri.User)
	b.WriteString("@")
	b.WriteString(uri.Host)

	for key, value := range uri.params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(value)
	}

	return b.String()
}

// Parses a sip uri (RFC 3291.19.1).
func Parse(str string) (*URI, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like SplitN().

	uri := &URI{
		params: map[string]string{},
	}

	blocks := strings.SplitN(str, "@", 2)
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid uri: expected '@' got '%s'", str)
	}

	lhs := strings.Split(blocks[0], ":")
	if len(lhs) < 2 {
		return nil, fmt.Errorf("invalid uri lhs: expected '<scheme>:<user>@' got '%s@'", blocks[0])
	}
	if lhs[0] == "sip" {
		uri.Secure = false
	} else if lhs[0] == "sips" {
		uri.Secure = true
	} else {
		return nil, fmt.Errorf("invalid uri scheme: expected 'sip:|sips:' got '%s:'", lhs[0])
	}
	uri.Host = lhs[len(lhs)-1]

	rhs := strings.Split(blocks[1], ";")
	for i, params := range rhs {
		if i == 0 {
			if params == "" {
				return nil, fmt.Errorf(
					"invalid uri rhs: expected '@<host>[:<port>]' got '@%s'", params,
				)
			}
			uri.Host = params
			continue
		}

		kv := strings.SplitN(params, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf(
				"invalid uri param: expected ';<key>=<value>' got ';%s'", params,
			)
		}

		uri.params[kv[0]] = kv[1]
	}

	return uri, nil
}
