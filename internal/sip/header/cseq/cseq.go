package cseq

import (
	"fmt"
	"strconv"
	"strings"
)

// CSeq header as specified in 3261.20.16.
type Header struct {
	Sequence uint32
	Method   string
}

func Serialize(header *Header) string {
	b := strings.Builder{}

	b.WriteString(strconv.Itoa(int(header.Sequence)))
	b.WriteString(" ")
	b.WriteString(header.Method)

	return b.String()
}

func Parse(str string) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{}

	blocks := strings.Split(str, " ")
	if len(blocks) != 2 {
		return nil, fmt.Errorf("invalid cseq header: expected '<sequence> <METHOD>' got '%s'", str)
	}

	sequence, err := strconv.ParseUint(blocks[0], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid cseq header: expected uint32 sequence got '%s'", blocks[0])
	}
	header.Sequence = uint32(sequence)
	header.Method = blocks[1]

	return header, nil
}
