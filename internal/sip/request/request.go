package request

import (
	"io"
	"strings"

	"github.com/megakuul/voiper/internal/sip/header/contentlength"
	"github.com/megakuul/voiper/internal/sip/uri"
)

type METHOD string

const (
	REGISTER  METHOD = "REGISTER"
	OPTIONS   METHOD = "OPTIONS"
	INVITE    METHOD = "INVITE"
	ACK       METHOD = "ACK"
	BYE       METHOD = "BYE"
	PRACK     METHOD = "PRACK"
	SUBSCRIBE METHOD = "SUBSCRIBE"
	NOTIFY    METHOD = "NOTIFY"
	PUBLISH   METHOD = "PUBLISH"
	INFO      METHOD = "INFO"
	UPDATE    METHOD = "UPDATE"
	MESSAGE   METHOD = "MESSAGE"
	REFER     METHOD = "REFER"
)

var methods map[string]METHOD = map[string]METHOD{
	"REGISTER":  REGISTER,
	"OPTIONS":   OPTIONS,
	"INVITE":    INVITE,
	"ACK":       ACK,
	"BYE":       BYE,
	"PRACK":     PRACK,
	"SUBSCRIBE": SUBSCRIBE,
	"NOTIFY":    NOTIFY,
	"PUBLISH":   PUBLISH,
	"INFO":      INFO,
	"UPDATE":    UPDATE,
	"MESSAGE":   MESSAGE,
	"REFER":     REFER,
}

type STATUS int

const (
	TRYING                  STATUS = 100
	RINGING                 STATUS = 180
	CALL_IS_BEING_FORWARDED STATUS = 181
	QUEUED                  STATUS = 182
	SESSION_PROGRESS        STATUS = 183

	OK       STATUS = 200
	ACCEPTED STATUS = 202

	MULTIPLE_CHOICES    STATUS = 300
	MOVED_PERMANENTLY   STATUS = 301
	MOVED_TEMPORARILY   STATUS = 302
	USE_PROXY           STATUS = 305
	ALTERNATIVE_SERVICE STATUS = 380 // "hey guys, I was thinking... lets add unnecessary complexity to the protocol"

	BAD_REQUEST                   STATUS = 400
	UNAUTHORIZED                  STATUS = 401
	FORBIDDEN                     STATUS = 403
	NOT_FOUND                     STATUS = 404
	PROXY_AUTHENTICATION_REQUIRED STATUS = 407
	REQUEST_TIMEOUT               STATUS = 408
	GONE                          STATUS = 410
	UNSUPPORTED_MEDIA_TYPE        STATUS = 415
	TEMPORARILY_UNAVAILABLE       STATUS = 480
	BUSY_HERE                     STATUS = 486
	REQUEST_TERMINATED            STATUS = 487

	INTERNAL_SERVER_ERROR STATUS = 500
	NOT_IMPLEMENTED       STATUS = 501
	BAD_GATEWAY           STATUS = 502
	SERVER_UNAVAILABLE    STATUS = 503
	SERVER_TIMEOUT        STATUS = 504

	BUSY_EVERYWHERE         STATUS = 600
	DECLINE                 STATUS = 603
	DOES_NOT_EXIST_ANYWHERE STATUS = 604
	NOT_ACCEPTABLE          STATUS = 606
)

type CONTENT_TYPE int

const (
	APPLICATION_SDP CONTENT_TYPE = iota
	APPLICATION_PIDF_XML
	MESSAGE_SIPFRAG
)

var contents map[string]CONTENT_TYPE = map[string]CONTENT_TYPE{
	"application/sdp":      APPLICATION_SDP,
	"application/pidf+xml": APPLICATION_PIDF_XML,
	"message/sipfrag":      MESSAGE_SIPFRAG,
}

type Request struct {
	Method  METHOD
	URI     uri.URI
	Version string
	Headers map[string][]string
	Body    string
}

func SerializeRequest(request *Request) string {
	b := strings.Builder{}

	b.WriteString(string(request.Method))
	b.WriteString(" ")
	b.WriteString(uri.Serialize(&request.URI))
	b.WriteString(" ")
	b.WriteString(request.Version)
	b.WriteString("\r\n")

	request.Headers["content-type"] = []string{
		contentlength.Serialize(&contentlength.Header{
			Length: uint32(len(request.Body)),
		}),
	}

	for key, values := range request.Headers {
		for _, value := range values {
			b.WriteString(key)
			b.WriteString(": ")
			b.WriteString(value)
			b.WriteString("\r\n")
		}
	}

	b.WriteString("\r\n")

	b.WriteString(request.Body)

	return b.String()
}

type Header struct {
	Method  METHOD
	URI     uri.URI
	Version string
	Headers map[string][]string
}

func ParseHeader(reader io.ReadCloser) (*Header, error) {
	// header := &Header{}

	// reader.Read()
	return nil, nil
}
