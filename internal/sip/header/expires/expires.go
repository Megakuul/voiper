package expires

import (
	"bytes"
	"fmt"
	"strconv"
	"time"
)

// Expires header as specified in 3261.20.19.
type Header struct {
	ExpiresIn time.Duration
}

func Serialize(header *Header) []byte {
	return strconv.AppendInt(nil, int64(header.ExpiresIn.Seconds()), 10)
}

func Parse(input []byte) (*Header, error) {
	header := &Header{}

	length, err := strconv.ParseUint(string(bytes.TrimSpace(input)), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid expires header: expected uint32 duration got '%s'", string(input))
	}
	header.ExpiresIn = time.Duration(uint32(length)) * time.Second

	return header, nil
}
