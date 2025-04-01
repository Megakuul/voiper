package contact

import (
	"bytes"
	"fmt"

	"github.com/megakuul/voiper/internal/sip/uri"
)

// Contact header according to RFC 3261.20.10.
// The header has three different formats "display-name", "name-addr" and "addr-spec"
// (d-n: '"Karlo" <sip:kater.karlo@entnet.com>' n-a: '<sip:kater,karlo@entnet.com>' a-s: 'sip:kater.karlo@entnet.com')
// I don't give a shit about those formats, however parsing of all formats works expected.
// Serialization always adds '<>' and is therefore never in addr-spec form (this avoids issues with special chars in the uri).
// Only one contact entry is supported, for multiple
type Header struct {
	DisplayName []byte
	Uri         *uri.URI
	Params      map[string][]byte
}

func Serialize(header *Header) []byte {
	b := bytes.Buffer{}
	if len(header.DisplayName) != 0 {
		b.WriteString("\"")
		b.Write(header.DisplayName)
		b.WriteString("\" ")
	}
	b.WriteString(" <")
	b.Write(uri.Serialize(header.Uri))
	b.WriteString(">")

	for key, value := range header.Params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.Write(value)
	}

	return b.Bytes()
}

func Parse(input []byte) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string][]byte{},
	}

	uriStr := []byte{}
	paramStr := []byte{}

	blocks := bytes.SplitN(bytes.TrimSpace(input), []byte("<"), 2)
	if len(blocks) == 2 {
		// uri has '<>' which means there could be a DisplayName
		header.DisplayName = bytes.TrimSuffix(bytes.TrimPrefix(
			bytes.TrimSpace(blocks[0]), []byte("\"")), []byte("\""),
		)
		data := bytes.SplitN(blocks[1], []byte(">"), 2)
		if len(data) != 2 {
			return nil, fmt.Errorf("invalid to header: expected closing '>' bracket got '... <%s'", string(blocks[1]))
		}
		uriStr = data[0]
		paramStr = bytes.TrimPrefix(bytes.TrimSpace(data[1]), []byte(";"))
	} else {
		// uri has no '<>' which means there is no DisplayName
		data := bytes.SplitN(input, []byte(";"), 2)
		if len(data) != 2 {
			uriStr = input
		} else {
			uriStr = bytes.TrimSpace(data[0])
			paramStr = bytes.TrimSpace(data[1]) // uri params are treated as header params if there are no '<>'
		}
	}

	var err error
	header.Uri, err = uri.Parse(uriStr)
	if err != nil {
		return nil, fmt.Errorf("invalid to header uri: %w", err)
	}

	params := bytes.Split(paramStr, []byte(";"))
	for _, param := range params {
		kv := bytes.SplitN(param, []byte("="), 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid to header param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", string(param))
		}
		key := bytes.TrimSpace(kv[0])
		if bytes.Contains(key, []byte(" ")) {
			// space in the key means there is a new scheme (which is ignored)
			break
		}
		// trims crap from the value (' "sipsucks\""' => 'sipsucks\"')
		header.Params[string(bytes.TrimSpace(kv[0]))] = bytes.TrimSuffix(bytes.TrimPrefix(
			bytes.TrimSpace(kv[1]), []byte("\"")), []byte("\""),
		)
	}

	return header, nil
}
