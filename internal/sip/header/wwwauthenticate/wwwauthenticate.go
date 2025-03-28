package wwwauthenticate

import (
	"fmt"
	"strings"
)

type SCHEME string

const (
	SCHEME_DIGEST = "Digest"
)

// WWW-Authenticate header as specified in the examples of 3261.22.2.
// It only reads the first authentication scheme ('Digest realm="x", Digest realm="y"' => (realm 'y' is ignored)).
// Supply multiple schemes by adding multiple headers.
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
			return nil, fmt.Errorf("invalid www-authenticate header param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", param)
		}
		// trims crap from the value ('" sipsucks\""' => 'sipsucks\"')
		header.Params[strings.TrimSpace(kv[0])] = strings.TrimSpace(strings.TrimSuffix(
			strings.TrimPrefix(kv[1], "\""), "\"",
		))
	}

	return header, nil
}
