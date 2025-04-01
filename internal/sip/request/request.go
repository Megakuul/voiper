package request

import (
	"bytes"
	"fmt"
	"io"
	"strings"
)

const (
	READ_BUFFER_SIZE = 1024
	MAX_HEADER_SIZE  = READ_BUFFER_SIZE * 8
)

type Request struct {
	Method  []byte
	URI     []byte
	Version []byte
	Headers map[string][][]byte
	Body    io.Reader
}

func Serialize(request *Request) io.Reader {
	b := bytes.Buffer{}

	b.Write(request.Method)
	b.WriteString(" ")
	b.Write(request.URI)
	b.WriteString(" ")
	b.Write(request.Version)
	b.WriteString("\r\n")

	for key, values := range request.Headers {
		for _, value := range values {
			b.WriteString(key)
			b.WriteString(": ")
			b.Write(value)
			b.WriteString("\r\n")
		}
	}

	b.WriteString("\r\n")

	return io.MultiReader(bytes.NewReader(b.Bytes()), request.Body)
}

// Peek checks if the io.Reader contains a valid request head.
// In any case it returns an io.Reader that reads exactly from where it left of before peeking.
func Peek(input io.Reader) (bool, io.Reader) {
	buffer := make([]byte, 100)
	n, err := input.Read(buffer)
	if err != nil {
		return false, input
	}
	lines := bytes.Split(buffer[:n], []byte("\r\n"))
	if len(lines) < 2 {
		return false, io.MultiReader(bytes.NewBuffer(buffer[:n]), input)
	}
	blocks := bytes.Split(lines[0], []byte(" "))
	if len(blocks) != 3 {
		return false, io.MultiReader(bytes.NewBuffer(buffer[:n]), input)
	}
	if !bytes.Equal(blocks[2], []byte("SIP/2.0")) {
		return false, io.MultiReader(bytes.NewBuffer(buffer[:n]), input)
	}
	return true, io.MultiReader(bytes.NewBuffer(buffer[:n]), input)
}

func Parse(input io.Reader) (*Request, error) {
	request := &Request{
		Headers: map[string][][]byte{},
	}
	builder := bytes.Buffer{}

	reads := 0
	for {
		reads++
		if reads*READ_BUFFER_SIZE > MAX_HEADER_SIZE {
			return nil, fmt.Errorf("invalid header: maximum header size ('%d') exceeded", MAX_HEADER_SIZE)
		}

		buffer := make([]byte, READ_BUFFER_SIZE)
		n, err := input.Read(buffer)
		if err != nil {
			return nil, err
		}

		for i := 0; i < n; i++ {
			if buffer[i] == '\r' && n > i+1 && buffer[i+1] == '\n' {
				i++ // skip the '\n'

				// if no header is buffered this means the header is done (double '\r\n')
				if builder.Len() == 0 {
					request.Body = io.MultiReader(bytes.NewReader(buffer[i+1:n]), input)
					return request, nil
				}

				line := bytes.Clone(builder.Bytes())
				builder.Reset()
				if len(request.Version) == 0 {
					blocks := bytes.SplitN(line, []byte(" "), 3)
					if len(blocks) != 3 {
						return nil, fmt.Errorf(
							"invalid request: expected '<METHOD> <uri> <version>\\r\\n' got '%.15s...'", string(line),
						)
					}
					request.Method, request.URI, request.Version = blocks[0], blocks[1], blocks[2]
				} else {
					blocks := bytes.SplitN(line, []byte(":"), 2)
					if len(blocks) != 2 {
						return nil, fmt.Errorf(
							"invalid request header: expected '<key>: <value>\\r\\n' got '%.15s...'", string(line),
						)
					}
					key := strings.ToLower(string(blocks[0]))
					request.Headers[key] = append(request.Headers[key], blocks[1])
				}
				continue
			}
			builder.WriteByte(buffer[i])
		}
	}
}
