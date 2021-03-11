package signature

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/slack-go/slack"
)

const (
	HeaderTimestamp = "X-Slack-Request-Timestamp"
	HeaderSignature = "X-Slack-Signature"
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

func AddSignature(h http.Header, key, body []byte, timestamp time.Time) error {
	hash := hmac.New(sha256.New, key)
	strTime := strconv.FormatInt(timestamp.Unix(), 10)
	if _, err := hash.Write([]byte(fmt.Sprintf("v0:%s:", strTime))); err != nil {
		return err
	}
	if _, err := hash.Write(body); err != nil {
		return err
	}
	signature := hex.EncodeToString(hash.Sum(nil))

	h.Set(HeaderTimestamp, strTime)
	h.Set(HeaderSignature, "v0="+signature)
	return nil
}
