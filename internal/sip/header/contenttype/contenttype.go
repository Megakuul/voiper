package contenttype

type CONTENT_TYPE string

const (
	CONTENT_SDP CONTENT_TYPE = "application/sdp"
)

// Content-Type header as specified in 3261.20.15.
type Header struct {
	Type CONTENT_TYPE
}

func Serialize(header *Header) string {
	return string(header.Type)
}

func Parse(str string) (*Header, error) {
	return &Header{
		Type: CONTENT_TYPE(str),
	}, nil
}
