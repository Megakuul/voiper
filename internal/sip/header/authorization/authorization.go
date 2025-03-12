package authorization

import (
	"fmt"
	"strings"
)

type SCHEME string

const (
	SCHEME_DIGEST = "Digest"
)

// Authorization header as specified in the examples of 3261.22.2.
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
		return nil, fmt.Errorf("invalid authorization header: expected '<Scheme> <key>=\"<value>\", ...' got '%s'", str)
	}
	header.Scheme = SCHEME(blocks[0])

	params := strings.Split(blocks[1], ",")
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid authorization header param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", param)
		}
		// trims crap from the value (' "sipsucks\""' => 'sipsucks\"')
		header.Params[strings.TrimSpace(kv[0])] = strings.TrimSuffix(strings.TrimPrefix(
			strings.TrimSpace(kv[1]), "\""), "\"",
		)
	}

	return header, nil
}
