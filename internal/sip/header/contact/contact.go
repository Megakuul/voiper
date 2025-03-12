package contact

import (
	"fmt"
	"strings"

	"github.com/megakuul/voiper/internal/sip/uri"
)

// Contact header according to RFC 3261.20.10.
// The header has three different formats "display-name", "name-addr" and "addr-spec"
// (d-n: '"Karlo" <sip:kater.karlo@entnet.com>' n-a: '<sip:kater,karlo@entnet.com>' a-s: 'sip:kater.karlo@entnet.com')
// I don't give a shit about those formats, however parsing of all formats works expected.
// Serialization always adds '<>' and is therefore never in addr-spec form (this avoids issues with special chars in the uri).
// Only one contact entry is supported, for multiple
type Header struct {
	DisplayName string
	Uri         *uri.URI
	Params      map[string]string
}

func Serialize(header *Header) string {
	b := strings.Builder{}
	if header.DisplayName != "" {
		b.WriteString("\"")
		b.WriteString(header.DisplayName)
		b.WriteString("\" ")
	}
	b.WriteString(" <")
	b.WriteString(uri.Serialize(header.Uri))
	b.WriteString(">")

	for key, value := range header.Params {
		b.WriteString(";")
		b.WriteString(key)
		b.WriteString("=")
		b.WriteString(value)
	}

	return b.String()
}

func Parse(str string) (*Header, error) {
	// performs many unnecessary string reallocs, if this is bottlenecking
	// it should be rewriten without the simple but slow string functions like TrimSpace() && SplitN().

	header := &Header{
		Params: map[string]string{},
	}

	uriStr := ""
	paramStr := ""

	blocks := strings.SplitN(str, "<", 2)
	if len(blocks) == 2 {
		// uri has '<>' which means there could be a DisplayName
		header.DisplayName = strings.TrimSuffix(strings.TrimPrefix(
			strings.TrimSpace(blocks[0]), "\""), "\"",
		)
		data := strings.SplitN(blocks[1], ">", 2)
		if len(data) != 2 {
			return nil, fmt.Errorf("invalid to header: expected closing '>' bracket got '... <%s'", blocks[1])
		}
		uriStr = data[0]
		paramStr = strings.TrimPrefix(strings.TrimSpace(data[1]), ";")
	} else {
		// uri has no '<>' which means there is no DisplayName
		data := strings.SplitN(str, ";", 2)
		if len(data) != 2 {
			uriStr = str
		} else {
			uriStr = strings.TrimSpace(data[0])
			paramStr = strings.TrimSpace(data[1]) // uri params are treated as header params if there are no '<>'
		}
	}

	var err error
	header.Uri, err = uri.Parse(uriStr)
	if err != nil {
		return nil, fmt.Errorf("invalid to header uri: %w", err)
	}

	params := strings.Split(paramStr, ";")
	for _, param := range params {
		kv := strings.SplitN(param, "=", 2)
		if len(kv) != 2 {
			return nil, fmt.Errorf("invalid to header param: expected '... <key>=\"<value>\", ...' got '... %s, ...'", param)
		}
		key := strings.TrimSpace(kv[0])
		if strings.Contains(key, " ") {
			// space in the key means there is a new scheme (which is ignored)
			break
		}
		// trims crap from the value (' "sipsucks\""' => 'sipsucks\"')
		header.Params[strings.TrimSpace(kv[0])] = strings.TrimSuffix(strings.TrimPrefix(
			strings.TrimSpace(kv[1]), "\""), "\"",
		)
	}

	return header, nil
}
