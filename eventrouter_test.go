package eventrouter_test

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	eventrouter "github.com/genkami/go-slack-event-router"
)

var _ = Describe("EventRouter", func() {
	Describe("New", func() {
		Context("when neither WithSigningToken nor InsecureSkipVerification is given", func() {
			It("returns an error", func() {
				_, err := eventrouter.New()
				Expect(err).To(MatchError(MatchRegexp("WithSigningToken")))
			})
		})

		Context("when InsecureSkipVerification is given", func() {
			It("returns a new Router", func() {
				r, err := eventrouter.New(eventrouter.InsecureSkipVerification())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			})
		})

		Context("when WithSigningToken is given", func() {
			It("returns a new Router", func() {
				r, err := eventrouter.New(eventrouter.WithSigningToken("THE_TOKEN"))
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			})
		})

		Context("when both WithSigningToken and InsecureSkipVerification are given", func() {
			It("returns an error", func() {
				_, err := eventrouter.New(eventrouter.InsecureSkipVerification(), eventrouter.WithSigningToken("THE_TOKEN"))
				Expect(err).To(MatchError(MatchRegexp("WithSigningToken")))
			})
		})
	})

	Describe("WithSigningSecret", func() {
		var (
			r     *eventrouter.Router
			token = "THE_TOKEN"
		)
		BeforeEach(func() {
			var err error
			r, err = eventrouter.New(eventrouter.WithSigningToken(token))
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the signature is valid", func() {
			It("responds with 200", func() {
				req, err := NewRequest(token, `{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
				"type": "url_verification"
			}`, nil)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				body, err := ioutil.ReadAll(resp.Body)
				fmt.Printf("body: %s\n", body)
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})
})

func NewRequest(signingSecret string, body string, ts *time.Time) (*http.Request, error) {
	hash := hmac.New(sha256.New, []byte(signingSecret))
	var now time.Time
	if ts == nil {
		now = time.Now()
	} else {
		now = *ts
	}
	timestamp := strconv.FormatInt(now.Unix(), 10)
	if _, err := hash.Write([]byte(fmt.Sprintf("v0:%s:", timestamp))); err != nil {
		return nil, err
	}
	if _, err := hash.Write([]byte(body)); err != nil {
		return nil, err
	}
	signature := hash.Sum(nil)
	req, err := http.NewRequest(http.MethodPost, "http://example.com/path/to/callback", bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Slack-Request-Timestamp", timestamp)
	req.Header.Set("X-Slack-Signature", "v0="+hex.EncodeToString(signature))
	return req, nil
}
