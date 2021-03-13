// TODO: move this to internal/routerutils
package routerutils

import (
	"errors"
	"net/http"

	routererrors "github.com/genkami/go-slack-event-router/errors"
)

func RespondWithError(w http.ResponseWriter, err error, verboseResponse bool) {
	var httpErr routererrors.HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(int(httpErr))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
	if verboseResponse {
		_, _ = w.Write([]byte(err.Error()))
	}
}
