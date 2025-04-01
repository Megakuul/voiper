package cseq

import (
	"bytes"
	"fmt"
	"strconv"
)

// CSeq header as specified in 3261.20.16.
type Header struct {
	Sequence uint32
	Method   []byte
}

func Serialize(header *Header) []byte {
	b := bytes.Buffer{}

	b.WriteString(strconv.Itoa(int(header.Sequence)))
	b.WriteString(" ")
	b.Write(header.Method)

	return b.Bytes()
}

func Parse(input []byte) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{}

	blocks := bytes.Split(bytes.TrimSpace(input), []byte(" "))
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid cseq header: expected '<sequence> <METHOD>' got '%s'", string(input))
	}

	sequence, err := strconv.ParseUint(string(blocks[0]), 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid cseq header: expected uint32 sequence got '%s'", string(blocks[0]))
	}
	header.Sequence = uint32(sequence)
	header.Method = blocks[1]

	return header, nil
}
