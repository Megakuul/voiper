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
	Method  string
	URI     string
	Version string
	Headers []string
	Body    io.Reader
}

func Serialize(request *Request) string {
	b := strings.Builder{}

	b.WriteString(request.Method)
	b.WriteString(" ")
	b.WriteString(request.URI)
	b.WriteString(" ")
	b.WriteString(request.Version)
	b.WriteString("\r\n")

	for _, header := range request.Headers {
		b.WriteString(header)
		b.WriteString("\r\n")
	}

	b.WriteString("\r\n")

	body, _ := io.ReadAll(request.Body)
	b.WriteString(string(body))

	return b.String()
}

func Parse(input io.Reader) (*Request, error) {
	request := &Request{}
	builder := strings.Builder{}

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

				line := builder.String()
				builder.Reset()
				if request.Version == "" {
					blocks := strings.SplitN(line, " ", 3)
					if len(blocks) != 3 {
						return nil, fmt.Errorf("invalid header request-line: expected '<METHOD> <uri> <version>'")
					}
					request.Method, request.URI, request.Version = blocks[0], blocks[1], blocks[2]
				} else {
					request.Headers = append(request.Headers, line)
				}
				continue
			}
			builder.WriteByte(buffer[i])
		}
	}
}
