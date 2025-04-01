package callid

import (
	"bytes"
	"fmt"
)

// Call-ID header as specified in 3261.20.8.
type Header struct {
	Identifier []byte
	Host       []byte
}

func Serialize(header *Header) []byte {
	b := bytes.Buffer{}

	b.Write(header.Identifier)
	b.WriteString("@")
	b.Write(header.Host)

	return b.Bytes()
}

func Parse(input []byte) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{}

	blocks := bytes.Split(bytes.TrimSpace(input), []byte("@"))
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid call-id header: expected '<identifier>@<host>' got '%s'", string(input))
	}

	header.Identifier = blocks[0]
	header.Host = blocks[1]

	return header, nil
}
