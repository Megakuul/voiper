package multiplexer

import (
	"context"

	"github.com/megakuul/voiper/internal/sip/request"
	"github.com/megakuul/voiper/internal/sip/response"
)

// Multiplexer provides a generic interface for the transaction transport multiplexer.
type Multiplexer interface {
	// Start a transaction identified by the specified transaction branch.
	// Returns a channel that emits all captured responses associated with the transaction.
	StartCall(string, *request.Request) (<-chan *response.Response, error)
	// Stops listening for responses of the specified transaction (and closes the response channel).
	StopCall(string)

	// Starts a transaction listener which captures incomming transaction requests with the specified method
	// and executes the associated callback function. The callback function receives the initial request and a channel
	// to emit all generated responses.
	StartListen(string, func(context.Context, *request.Request, chan *response.Response) error) error
	// Stops listening for further transactions. This does not stop running callbacks.
	StopListen(string)

	// Shutdown closes all running callbacks, connections, listeners, etc. of the multiplexer.
	// After shutdown trying to add new listeners or calls will fail and return an error.
	Shutdown()
}
