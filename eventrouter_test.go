package eventrouter_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	eventrouter "github.com/genkami/go-slack-event-router"
	"github.com/genkami/go-slack-event-router/signature"
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
			r       *eventrouter.Router
			token   = "THE_TOKEN"
			content = `
			{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
				"type": "url_verification"
			}`
		)
		BeforeEach(func() {
			var err error
			r, err = eventrouter.New(eventrouter.WithSigningToken(token), eventrouter.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the signature is valid", func() {
			It("responds with 200", func() {
				req, err := NewSignedRequest(token, content, nil)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when the signature is invalid", func() {
			It("responds with Unauthorized", func() {
				req, err := NewSignedRequest(token, content, nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set(signature.HeaderSignature, "v0="+hex.EncodeToString([]byte("INVALID_SIGNATURE")))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})

		Context("when the timestamp is too old", func() {
			It("responds with Unauthorized", func() {
				ts := time.Now().Add(-1 * time.Hour)
				req, err := NewSignedRequest(token, content, &ts)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusUnauthorized))
			})
		})
	})

	Describe("InsecureSkipVerification", func() {
		var (
			r       *eventrouter.Router
			token   = "THE_TOKEN"
			content = `
			{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"challenge": "3eZbrw1aBm2rZgRNFdxV2595E9CY3gmdALWMmHkvFXO7tYXAYM8P",
				"type": "url_verification"
			}`
		)
		BeforeEach(func() {
			var err error
			r, err = eventrouter.New(eventrouter.InsecureSkipVerification(), eventrouter.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the signature is valid", func() {
			It("responds with 200", func() {
				req, err := NewSignedRequest(token, content, nil)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when the signature is invalid", func() {
			It("responds with 200", func() {
				req, err := NewSignedRequest(token, content, nil)
				Expect(err).NotTo(HaveOccurred())
				req.Header.Set(signature.HeaderSignature, "v0="+hex.EncodeToString([]byte("INVALID_SIGNATURE")))
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when the timestamp is too old", func() {
			It("responds with 200", func() {
				ts := time.Now().Add(-1 * time.Hour)
				req, err := NewSignedRequest(token, content, &ts)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})
	})

	Describe("URL Verification", func() {
		var (
			r *eventrouter.Router
		)
		BeforeEach(func() {
			var err error
			r, err = eventrouter.New(eventrouter.InsecureSkipVerification(), eventrouter.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		It("returns the given challenge in JSON", func() {
			req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(`
			{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"challenge": "THE_SECRET_CHALLENGE_VALUE",
				"type": "url_verification"
			}
			`)))
			Expect(err).NotTo(HaveOccurred())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			resp := w.Result()
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
			dec := json.NewDecoder(resp.Body)
			body := slackevents.ChallengeResponse{}
			err = dec.Decode(&body)
			Expect(err).NotTo(HaveOccurred())
			Expect(body.Challenge).To(Equal("THE_SECRET_CHALLENGE_VALUE"))
		})
	})

	Describe("App Rate Limited", func() {
		var (
			r *eventrouter.Router
		)
		BeforeEach(func() {
			var err error
			r, err = eventrouter.New(eventrouter.InsecureSkipVerification(), eventrouter.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		It("responds with 200", func() {
			req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(`
			{
				"token": "Jhj5dZrVaK7ZwHHjRyZWjbDl",
				"type": "app_rate_limited",
				"team_id": "T123456",
				"minute_rate_limited": 1518467820,
				"api_app_id": "A123456"
			}
			`)))
			Expect(err).NotTo(HaveOccurred())
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			resp := w.Result()
			body, _ := ioutil.ReadAll(resp.Body)
			fmt.Printf("body: %s\n", body)
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})
})

func NewSignedRequest(signingSecret string, body string, ts *time.Time) (*http.Request, error) {
	var now time.Time
	if ts == nil {
		now = time.Now()
	} else {
		now = *ts
	}
	req, err := http.NewRequest(http.MethodPost, "http://example.com/path/to/callback", bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	if err := signature.AddSignature(req.Header, []byte(signingSecret), []byte(body), now); err != nil {
		return nil, err
	}
	return req, nil
}
