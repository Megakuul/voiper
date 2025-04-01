package auth

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/megakuul/voiper/internal/sip/header/authorization"
	"github.com/megakuul/voiper/internal/sip/header/wwwauthenticate"
)

type Options struct {
	Method   []byte
	URI      []byte
	Username []byte
	Password []byte
}

// Authenticate checks the first authenticate scheme and converts it into an authorization header
// based on the provided authenticate header and the input options.
func Authenticate(wa *wwwauthenticate.Header, opts *Options) (*authorization.Header, error) {
	switch wa.Scheme {
	case authorization.SCHEME_DIGEST:
		algorithm := "md5"
		if algo := wa.Params["algorithm"]; algo != nil {
			algorithm = string(algo)
		}
		realm := wa.Params["realm"]
		if realm == nil {
			return nil, fmt.Errorf("required param 'realm' is missing in authentication header")
		}
		nonce := wa.Params["nonce"]
		if nonce == nil {
			return nil, fmt.Errorf("required param 'nonce' is missing in authentication header")
		}

		response, err := authDigest(algorithm, realm, nonce, opts.Method, opts.URI, opts.Username, opts.Password)
		if err != nil {
			return nil, err
		}
		return &authorization.Header{
			Scheme: authorization.SCHEME_DIGEST,
			Params: map[string][]byte{
				"algorithm": []byte(algorithm),
				"username":  opts.Username,
				"realm":     realm,
				"nonce":     nonce,
				"uri":       opts.URI,
				"response":  response,
			},
		}, nil
	default:
		return nil, fmt.Errorf("unsupported authentication scheme: %s", wa.Scheme)
	}
}

// authDigest creates a digest hash from the provided inputs. Follows rfc2617 but qop is not supported.
func authDigest(algorithm string, realm, nonce, method, uri, username, password []byte) ([]byte, error) {
	switch strings.ToLower(algorithm) {
	case "md5":
		hasher := md5.New()

		hasher.Write(username)
		hasher.Write([]byte(":"))
		hasher.Write(realm)
		hasher.Write([]byte(":"))
		hasher.Write(password)

		ha1 := hasher.Sum(nil)
		hasher.Reset()

		hasher.Write(method)
		hasher.Write([]byte(":"))
		hasher.Write(uri)

		ha2 := hasher.Sum(nil)
		hasher.Reset()

		hasher.Write(hex.AppendEncode(nil, ha1[:]))
		hasher.Write([]byte(":"))
		hasher.Write(nonce)
		hasher.Write([]byte(":"))
		hasher.Write(hex.AppendEncode(nil, ha2[:]))

		return hex.AppendEncode(nil, hasher.Sum(nil)), nil
	default:
		return nil, fmt.Errorf("unsupported algorithm '%s' in digest authentication header", algorithm)
	}
}
