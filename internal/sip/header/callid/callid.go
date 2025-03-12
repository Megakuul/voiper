package callid

import (
	"fmt"
	"strings"
)

// Call-ID header as specified in 3261.20.8.
type Header struct {
	Identifier string
	Host       string
}

func Serialize(header *Header) string {
	b := strings.Builder{}

	b.WriteString(header.Identifier)
	b.WriteString("@")
	b.WriteString(header.Host)

	return b.String()
}

func Parse(str string) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{}

	blocks := strings.Split(str, "@")
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid call-id header: expected '<identifier>@<host>' got '%s'", str)
	}

	header.Identifier = blocks[0]
	header.Host = blocks[1]

	return header, nil
}
