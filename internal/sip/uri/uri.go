package uri

import (
	"bytes"
	"fmt"
)

// SIP uri according to RFC 3261.19.1.
// Host includes the optional 'Port' (e.g. Host: "server.local:5060").
// Ignores 'password' segment because there is no logical reason to ever use this.
// (they even say it should not be used.. IN THEIR OWN SPEC WTF?!)
type URI struct {
	Secure bool
	User   []byte
	Host   []byte
	Params map[string][]byte
}

// Serializes a sip uri (RFC 3261.19.1).
func Serialize(uri *URI) []byte {
	b := bytes.Buffer{}
	if uri.Secure {
		b.WriteString("sips:")
	} else {
		b.WriteString("sip:")
	}

	if len(uri.User) != 0 {
		b.Write(uri.User)
		b.WriteString("@")
	}
	b.Write(uri.Host)

	for key, value := range uri.Params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.Write(value)
	}

	return b.Bytes()
}

// Parses a sip uri (RFC 3291.19.1).
func Parse(input []byte) (*URI, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like SplitN().

	uri := &URI{
		Params: map[string][]byte{},
	}

	blocks := bytes.SplitN(input, []byte("@"), 2)
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid uri: expected '@' got '%s'", string(input))
	}

	lhs := bytes.Split(blocks[0], []byte(":"))
	if len(lhs) < 2 {
		return nil, fmt.Errorf("invalid uri lhs: expected '<scheme>:<user>@' got '%s@'", string(blocks[0]))
	}
	if bytes.Equal(lhs[0], []byte("sip")) {
		uri.Secure = false
	} else if bytes.Equal(lhs[0], []byte("sips")) {
		uri.Secure = true
	} else {
		return nil, fmt.Errorf("invalid uri scheme: expected 'sip:|sips:' got '%s:'", string(lhs[0]))
	}
	uri.Host = lhs[len(lhs)-1]

	rhs := bytes.Split(blocks[1], []byte(";"))
	for i, param := range rhs {
		if i == 0 {
			if len(param) == 0 {
				return nil, fmt.Errorf(
					"invalid uri rhs: expected '@<host>[:<port>]' got '@%s'", string(param),
				)
			}
			uri.Host = param
			continue
		}

		kv := bytes.SplitN(param, []byte("="), 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid uri param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", param)
		}
		// trims crap from the value (' "sipsucks\""' => 'sipsucks\"')
		uri.Params[string(bytes.TrimSpace(kv[0]))] = bytes.TrimSuffix(bytes.TrimPrefix(
			bytes.TrimSpace(kv[1]), []byte("\"")), []byte("\""),
		)
	}

	return uri, nil
}
