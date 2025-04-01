package register

import (
	"bytes"
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/megakuul/voiper/internal/sip/auth"
	"github.com/megakuul/voiper/internal/sip/header/authorization"
	"github.com/megakuul/voiper/internal/sip/header/callid"
	"github.com/megakuul/voiper/internal/sip/header/contact"
	"github.com/megakuul/voiper/internal/sip/header/contentlength"
	"github.com/megakuul/voiper/internal/sip/header/cseq"
	"github.com/megakuul/voiper/internal/sip/header/expires"
	"github.com/megakuul/voiper/internal/sip/header/from"
	"github.com/megakuul/voiper/internal/sip/header/to"
	"github.com/megakuul/voiper/internal/sip/header/via"
	"github.com/megakuul/voiper/internal/sip/header/wwwauthenticate"
	"github.com/megakuul/voiper/internal/sip/multiplexer"
	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/uri"
)

type Status struct {
	Code    int64
	Message string
}

type Input struct {
	Secure      bool
	LocalAddr   []byte
	RemoteAddr  []byte
	DisplayName []byte
	Username    []byte
	Password    []byte
	CallID      []byte
	FromTag     []byte
	CSeq        uint32
	ExpiresIn   time.Duration

	Authorized bool
	Additional map[string][][]byte
}

type Output struct {
	CSeq      uint32
	ExpiresIn time.Duration
}

// Performs a full registration transaction. If authentication is required the transaction automatically restarts with
// an authorization header derived from the provided password. The status of responses is emitted to the status channel.
func Register(ctx context.Context, m multiplexer.Multiplexer, status chan<- *Status, input *Input) (*Output, error) {
	branch := fmt.Sprint(via.IDIOT_SANDWICH_COOKIE + uuid.New().String())
	defer m.StopCall(branch)

	input.CSeq++

	registerUri := &uri.URI{
		Secure: input.Secure,
		Host:   input.RemoteAddr,
	}

	req := &request.Request{
		Method:  []byte("REGISTER"),
		URI:     uri.Serialize(registerUri),
		Version: []byte("SIP/2.0"),
		Headers: map[string][][]byte{
			"via": {via.Serialize(&via.Header{
				Version:  []byte("SIP/2.0"),
				Protocol: m.Protocol(),
				Host:     input.LocalAddr,
				Params: map[string][]byte{
					"branch": []byte(branch),
				},
			})},
			"to": {to.Serialize(&to.Header{
				DisplayName: []byte(input.DisplayName),
				Uri: &uri.URI{
					Secure: input.Secure,
					User:   input.Username,
					Host:   input.RemoteAddr,
				},
			})},
			"from": {from.Serialize(&from.Header{
				DisplayName: input.DisplayName,
				Uri: &uri.URI{
					Secure: input.Secure,
					User:   input.Username,
					Host:   input.RemoteAddr,
					Params: map[string][]byte{
						"tag": input.FromTag,
					},
				},
			})},
			"call-id": {callid.Serialize(&callid.Header{
				Identifier: input.CallID,
				Host:       input.RemoteAddr,
			})},
			"cseq": {cseq.Serialize(&cseq.Header{
				Sequence: input.CSeq,
				Method:   []byte("REGISTER"),
			})},
			"contact": {contact.Serialize(&contact.Header{
				DisplayName: input.DisplayName,
				Uri: &uri.URI{
					Secure: input.Secure,
					User:   input.Username,
					Host:   input.LocalAddr,
				},
			})},
			"expires": {expires.Serialize(&expires.Header{
				ExpiresIn: input.ExpiresIn,
			})},
			"content-length": {contentlength.Serialize(&contentlength.Header{
				Length: 0,
			})},
		},
		Body: bytes.NewReader(nil),
	}

	if input.Additional == nil {
		input.Additional = map[string][][]byte{}
	}
	for key, values := range input.Additional {
		req.Headers[key] = append(req.Headers[key], values...)
	}

	responses, err := m.StartCall(branch, req)
	if err != nil {
		return nil, err
	}

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("context exceeded")
		case res, ok := <-responses:
			if !ok {
				return nil, fmt.Errorf("transaction stopped")
			}

			code, err := strconv.ParseInt(string(res.Code), 10, 64)
			if err != nil {
				return nil, fmt.Errorf("expected valid response code: %v", err)
			}

			select {
			case status <- &Status{Code: code, Message: string(res.Status)}:
			default:
			}

			switch code {
			case 100:
				continue
			case 200:
				// for simplicity we just use the requested expiration. If the remote peer changes the expiration
				// it could be extracted from the contact-header parameter 'expires' according to rfc3261.10.2.8.
				return &Output{
					CSeq:      input.CSeq,
					ExpiresIn: input.ExpiresIn,
				}, nil
			case 301, 302, 305:
				// not supported because the underlying multiplexer is bound to exactly one remote peer.
				return nil, fmt.Errorf("redirection is not supported by this client")
			case 401:
				if input.Authorized {
					return nil, fmt.Errorf("invalid user credentials")
				}
				waValues, ok := res.Headers["www-authenticate"]
				if ok && len(waValues) > 0 {
					waHeader, err := wwwauthenticate.Parse(waValues[0])
					if err != nil {
						return nil, err
					}
					authHeader, err := auth.Authenticate(waHeader, &auth.Options{
						Method:   []byte("REGISTER"),
						URI:      uri.Serialize(registerUri),
						Username: input.Username,
						Password: input.Password,
					})
					if err != nil {
						return nil, err
					}
					input.Authorized = true
					input.Additional["authorization"] = [][]byte{authorization.Serialize(authHeader)}
					return Register(ctx, m, status, input)
				}
				return nil, fmt.Errorf("expected 'www-authenticate' header in response")
			case 407:
				if input.Authorized {
					return nil, fmt.Errorf("invalid user credentials")
				}
				paValues, ok := res.Headers["proxy-authenticate"]
				if ok && len(paValues) > 0 {
					paHeader, err := wwwauthenticate.Parse(paValues[0])
					if err != nil {
						return nil, err
					}
					authHeader, err := auth.Authenticate(paHeader, &auth.Options{
						Method:   []byte("REGISTER"),
						URI:      uri.Serialize(registerUri),
						Username: input.Username,
						Password: input.Password,
					})
					if err != nil {
						return nil, err
					}
					input.Authorized = true
					input.Additional["proxy-authorization"] = [][]byte{authorization.Serialize(authHeader)}
					return Register(ctx, m, status, input)
				}
				return nil, fmt.Errorf("expected 'proxy-authenticate' header in response")
			default:
				return nil, fmt.Errorf("transaction failed with status code '%d'", code)
			}
		}
	}
}
