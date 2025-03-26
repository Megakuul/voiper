package multiplexer

import "github.com/megakuul/voiper/internal/sip/request"

// Multiplexer provides a generic interface for the transaction transport multiplexer.
type Multiplexer interface {
	// Start a transaction identified by the specified transaction id.
	// Returns a channel that emits all captured responses associated with the transaction.
	Start(string, *request.Request) (<-chan *response.Response, error)
	// Stops the specified transaction and cleans up resources.
	Stop(string) error
}
