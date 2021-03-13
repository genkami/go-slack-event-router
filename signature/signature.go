// Package signature provides helpers to validate request signature.
//
// Fore more details, see https://api.slack.com/authentication/verifying-requests-from-slack.
package signature

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

// Middleware is an `http.Handler` middleware that automatically verifies request signatures.
type Middleware struct {
	// Secret is a signing secret.
	//
	// You can find this value by following this instruction: https://api.slack.com/authentication/verifying-requests-from-slack#signing_secrets_admin_page
	Secret string

	// If set to true, the middleware puts error details to the response body when it fails verification.
	VerboseResponse bool

	// Handler is an internal handler to perform actual request processing.
	Handler http.Handler
}

func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	verifier, err := slack.NewSecretsVerifier(r.Header, m.Secret)
	if err != nil {
		if errors.Is(err, slack.ErrExpiredTimestamp) {
			w.WriteHeader(http.StatusUnauthorized)
		} else {
			w.WriteHeader(http.StatusInternalServerError)
		}
		if m.VerboseResponse {
			fmt.Fprintf(w, "failed to initialize verifier: %s", err.Error())
		}
		return
	}
	tee := io.TeeReader(r.Body, &verifier)
	body, err := ioutil.ReadAll(tee)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		if m.VerboseResponse {
			fmt.Fprintf(w, "failed to read response: %s", err.Error())
		}
		return
	}
	if err := verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if m.VerboseResponse {
			fmt.Fprintf(w, "verification failed: %s", err.Error())
		}
		return
	}
	r.Body = ioutil.NopCloser(bytes.NewReader(body))
	m.Handler.ServeHTTP(w, r)
}
