package multiplexer

import (
	"bytes"
	"fmt"
	"io"
	"log/slog"
	"net"
	"strconv"
	"unsafe"
)

type Processor struct {
	packetSize int
}

func (p *Processor) ProcessUnreliable(reader io.Reader) error {
	header := &Header{}
	packet := make([]byte, p.packetSize)

	for {
		n, err := reader.Read(packet)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				slog.Warn(fmt.Sprintf("network read error: %v", err))
				continue
			}
			return err
		}

		err = ParseHeader(packet[:n], header)
		if err != nil {
			slog.Warn(fmt.Sprintf("packet error: %v", err))
			continue
		}

		// heuristic to avoid parsing packets that are definitely to long.
		if header.ContentLength > int64(n) {
			slog.Warn("packet error: specified content-length exceeds the udp packet buffer")
			continue
		}

	}
}

type Header struct {
	Method        []byte
	URI           []byte
	Version       []byte
	Headers       [][]byte
	ContentLength int64
}

const CL_HDR = "content-length:"
const CL_HDR_LEN = len(CL_HDR)

// ParseHeader does basic parsing on the header that is relevant for the transport layer.
// It performs zero allocations after initialization... not because it's required, but because it's cool.
func ParseHeader(input []byte, header *Header) error {
	var err error

	if header.Headers == nil {
		header.Headers = make([][]byte, 256)
	}
	header.Headers = header.Headers[:0]

	blocks := bytes.SplitN(input, []byte(" "), 3)
	if len(blocks) != 3 {
		return fmt.Errorf("invalid header request-line: expected '<METHOD> <uri> <version>'")
	}
	header.Method, header.URI = blocks[0], blocks[1]

	fields := bytes.Split(blocks[0], []byte("\r\n"))
	if len(fields) < 2 {
		return fmt.Errorf("invalid header request-line: expected '<METHOD> <uri> <version>\\r\\n'")
	}
	header.Version = fields[0]

	for _, field := range fields[1:] {
		if len(field) == 0 {
			break
		}
		if len(field) > CL_HDR_LEN && bytes.EqualFold(field[:CL_HDR_LEN], []byte(CL_HDR)) {
			data := bytes.TrimSpace(field[CL_HDR_LEN:])
			header.ContentLength, err = strconv.ParseInt(
				unsafe.String(&data[0], len(data)), 10, 64,
			)
			if err != nil {
				return fmt.Errorf("invalid header content-length: expected 'content-length: <int64>'")
			}
		}
		header.Headers = append(header.Headers, field)
	}

	return nil
}

func (p *Processor) ProcessReliable(reader io.Reader) {

}
