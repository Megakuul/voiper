package response

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

type Response struct {
	Version string
	Code    string
	Status  string
	Headers []string
	Body    io.Reader
}

func Serialize(response *Response) string {
	b := strings.Builder{}

	b.WriteString(response.Version)
	b.WriteString(" ")
	b.WriteString(response.Code)
	b.WriteString(" ")
	b.WriteString(response.Status)
	b.WriteString("\r\n")

	for _, header := range response.Headers {
		b.WriteString(header)
		b.WriteString("\r\n")
	}

	b.WriteString("\r\n")

	body, _ := io.ReadAll(response.Body)
	b.WriteString(string(body))

	return b.String()
}

func Parse(input io.Reader) (*Response, error) {
	response := &Response{}
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
					response.Body = io.MultiReader(bytes.NewReader(buffer[i+1:n]), input)
					return response, nil
				}

				line := builder.String()
				builder.Reset()
				if response.Version == "" {
					blocks := strings.SplitN(line, " ", 3)
					if len(blocks) != 3 {
						return nil, fmt.Errorf("invalid header response-line: expected '<version> <code> <status>'")
					}
					response.Version, response.Code, response.Status = blocks[0], blocks[1], blocks[2]
				} else {
					response.Headers = append(response.Headers, line)
				}
				continue
			}
			builder.WriteByte(buffer[i])
		}
	}
}
