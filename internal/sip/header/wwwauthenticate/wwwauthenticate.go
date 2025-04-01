package wwwauthenticate

import (
	"bytes"
	"fmt"
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
	Params map[string][]byte
}

func Serialize(header *Header) []byte {
	b := bytes.Buffer{}

	b.WriteString(string(header.Scheme))
	b.WriteString(" ")

	for key, value := range header.Params {
		b.WriteString(key)
		b.WriteString("=\"")
		b.Write(value)
		b.WriteString("\",")
	}

	return bytes.TrimPrefix(b.Bytes(), []byte(","))
}

func Parse(input []byte) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string][]byte{},
	}

	blocks := bytes.SplitN(bytes.TrimSpace(input), []byte(" "), 2)
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid www-authenticate header: expected '<Scheme> <key>=\"<value>\", ...' got '%s'", string(input))
	}
	header.Scheme = SCHEME(blocks[0])

	params := bytes.Split(blocks[1], []byte(","))
	for _, param := range params {
		kv := bytes.SplitN(param, []byte("="), 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid www-authenticate header param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", string(param))
		}
		// trims crap from the value ('" sipsucks\""' => 'sipsucks\"')
		header.Params[string(bytes.TrimSpace(kv[0]))] = bytes.TrimSpace(bytes.TrimSuffix(
			bytes.TrimPrefix(kv[1], []byte("\"")), []byte("\""),
		))
	}

	return header, nil
}
