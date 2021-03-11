package signature

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack"
)

func Middleware(signingSecret string, handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		tee := io.TeeReader(r.Body, &verifier)
		body, err := ioutil.ReadAll(tee)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		if err := verifier.Ensure(); err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		r.Body = ioutil.NopCloser(bytes.NewReader(body))
		handler.ServeHTTP(w, r)
	})
}
