package contentlength

import (
	"fmt"
	"strconv"
)

// Content-Type header as specified in 3261.20.14.
type Header struct {
	Length uint32
}

func Serialize(header *Header) string {
	return strconv.Itoa(int(header.Length))
}

func Parse(str string) (*Header, error) {
	header := &Header{}

	length, err := strconv.ParseUint(str, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid content-length header: expected uint32 length got '%s'", str)
	}
	header.Length = uint32(length)

	return header, nil
}
