package wwwauthenticate

import (
	"fmt"
	"strings"
)

type SCHEME string

const (
	SCHEME_DIGEST = "Digest"
	SCHEME_BEARER = "Bearer"
)

// WWW-Authenticate header as specified in the HTTP RFC 7235.4.1.
// It only reads the first authentication scheme ('Digest x="y", Bearer y="x"' only).
// Supply multiple schemes by adding multiple headers.
// No idea if SIP also strictly follows 7235 for this header, but I assume.
type Header struct {
	Scheme SCHEME
	Params map[string]string
}

func Serialize(header *Header) string {
	b := strings.Builder{}

	b.WriteString(string(header.Scheme))
	b.WriteString(" ")

	for key, value := range header.Params {
		b.WriteString(key)
		b.WriteString("=\"")
		b.WriteString(value)
		b.WriteString("\",")
	}

	return strings.TrimPrefix(b.String(), ",")
}

func Parse(str string) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string]string{},
	}

	blocks := strings.SplitN(str, " ", 2)
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid www-authenticate header: expected '<Scheme> <key>=\"<value>\", ...' got '%s'", str)
	}
	header.Scheme = SCHEME(blocks[0])

	params := strings.Split(blocks[1], ",")
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid www-authenticate param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", param)
		}
		key := strings.TrimSpace(kv[0])
		if strings.Contains(key, " ") {
			// space in the key means there is a new scheme (which is ignored)
			break
		}
		// trims crap from the value ('" sipsucks\""' => 'sipsucks\"')
		header.Params[key] = strings.TrimSpace(strings.TrimSuffix(
			strings.TrimPrefix(kv[1], "\""), "\"",
		))
	}

	return header, nil
}
