package contentlength

import (
	"bytes"
	"fmt"
	"strconv"
)

// Content-Type header as specified in 3261.20.14.
type Header struct {
	Length uint32
}

func Serialize(header *Header) []byte {
	return strconv.AppendInt(nil, int64(header.Length), 10)
}

func Parse(input []byte) (*Header, error) {
	header := &Header{}

	length, err := strconv.ParseUint(string(bytes.TrimSpace(input)), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid content-length header: expected uint32 length got '%s'", string(input))
	}
	header.Length = uint32(length)

	return header, nil
}
