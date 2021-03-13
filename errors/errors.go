// Package errors provides error values and types that are intended to be used to implement handlers.
package errors

import (
	"errors"
	"net/http"
)

// NotInterested indicates that the handler does not interested in the incoming events or actions.
// When this error is returned from handlers, the processing event is falled back to another handler.
var NotInterested = errors.New("not interested")

// HttpError represents errors that can be represented as http status codes.
// When the router receives this error, the router responds with the corresponding status code.
type HttpError int

func (e HttpError) Error() string {
	return http.StatusText(int(e))
}

var _ error = HttpError(0)
