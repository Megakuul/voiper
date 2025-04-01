package contenttype

import "bytes"

type CONTENT_TYPE string

const (
	CONTENT_SDP CONTENT_TYPE = "application/sdp"
)

// Content-Type header as specified in 3261.20.15.
type Header struct {
	Type CONTENT_TYPE
}

func Serialize(header *Header) []byte {
	return []byte(header.Type)
}

func Parse(input []byte) (*Header, error) {
	return &Header{
		Type: CONTENT_TYPE(bytes.TrimSpace(input)),
	}, nil
}
