package eventrouter_test

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/slack-go/slack/slackevents"

	eventrouter "github.com/genkami/go-slack-event-router"
	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/internal/testutils"
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
				req.Header.Set(testutils.HeaderSignature, "v0="+hex.EncodeToString([]byte("INVALID_SIGNATURE")))
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
				req.Header.Set(testutils.HeaderSignature, "v0="+hex.EncodeToString([]byte("INVALID_SIGNATURE")))
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
			Expect(resp.StatusCode).To(Equal(http.StatusOK))
		})
	})

	Describe("On", func() {
		var (
			r       *eventrouter.Router
			content = `
			{
				"token": "XXYYZZ",
				"team_id": "TXXXXXXXX",
				"api_app_id": "AXXXXXXXXX",
				"event": {
					"type": "message",
					"channel": "C2147483705",
					"user": "U2147483697",
					"text": "Hello world",
					"ts": "1355517523.000005"
				},
				"type": "event_callback",
				"event_context": "EC12345",
				"event_id": "Ev08MFMKH6",
				"event_time": 1234567890
			}`
			numHandlerCalled = 0
			handler          = eventrouter.HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
			var err error
			r, err = eventrouter.New(eventrouter.InsecureSkipVerification(), eventrouter.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no handler is registered", func() {
			It("just responds with 200", func() {
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when a matching handler is registered", func() {
			It("calls the handler and responds with 200", func() {
				r.On(slackevents.Message, handler)
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when a matching handler is registered to a different type of events", func() {
			It("does not call the handler and responds with 200", func() {
				r.On("other_type", handler)
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when a handler returned an error", func() {
			It("responds with InternalServerError", func() {
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					return fmt.Errorf("something wrong happened")
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when a handler returned NotInterested", func() {
			It("responds with 200", func() {
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					return routererrors.NotInterested
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when a handler returned an error that equals to NotInterested using errors.Is", func() {
			It("responds with 200", func() {
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					return errors.WithMessage(routererrors.NotInterested, "not interested")
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when a handler returned an HttpError", func() {
			It("responds with a corresponding status code", func() {
				code := http.StatusUnauthorized
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					return routererrors.HttpError(code)
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(code))
			})
		})

		Context("when a handler returned an error that equals to HttpError using errors.As", func() {
			It("responds with a corresponding status code", func() {
				code := http.StatusUnauthorized
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					return errors.WithMessage(routererrors.HttpError(code), "you ain't authorized")
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(code))
			})
		})

		Describe("Fallback", func() {
			var (
				numFirstHandlerCalled  int
				numSecondHandlerCalled int
				numFallbackCalled      int
				firstError             error
				secondError            error
				fallbackError          error
			)
			BeforeEach(func() {
				numFirstHandlerCalled = 0
				numSecondHandlerCalled = 0
				numFallbackCalled = 0
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numFirstHandlerCalled++
					return firstError
				}))
				r.On(slackevents.Message, eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numSecondHandlerCalled++
					return secondError
				}))
				r.SetFallback(eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numFallbackCalled++
					return fallbackError
				}))
			})

			Context("when a first handler returned nil", func() {
				It("does not fall back to other handlers", func() {
					firstError = nil
					secondError = nil
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(0))
					Expect(numFallbackCalled).To(Equal(0))
				})
			})

			Context("when a first handler returned an error", func() {
				It("responds with InternalServerError and does not fall back to other handlers", func() {
					firstError = errors.New("error in the first handler")
					secondError = nil
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(0))
					Expect(numFallbackCalled).To(Equal(0))
				})
			})

			Context("when a first handler returned NotInterested", func() {
				It("falls back to another handler", func() {
					firstError = routererrors.NotInterested
					secondError = nil
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(1))
					Expect(numFallbackCalled).To(Equal(0))
				})
			})

			Context("when a first handler returned an error that equals to NotInterested using errors.Is", func() {
				It("falls back to another handler", func() {
					firstError = errors.WithMessage(routererrors.NotInterested, "not interested")
					secondError = nil
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(1))
					Expect(numFallbackCalled).To(Equal(0))
				})
			})

			Context("when the last handler returned NotInterested", func() {
				It("falls back to fallback handler", func() {
					firstError = routererrors.NotInterested
					secondError = routererrors.NotInterested
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(1))
					Expect(numFallbackCalled).To(Equal(1))
				})
			})

			Context("when the last handler returned an error that equals to NotInterested using errors.Is", func() {
				It("falls back to fallback handler", func() {
					firstError = routererrors.NotInterested
					secondError = errors.WithMessage(routererrors.NotInterested, "not interested")
					fallbackError = nil
					req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
					Expect(err).NotTo(HaveOccurred())
					w := httptest.NewRecorder()
					r.ServeHTTP(w, req)
					resp := w.Result()
					Expect(resp.StatusCode).To(Equal(http.StatusOK))
					Expect(numFirstHandlerCalled).To(Equal(1))
					Expect(numSecondHandlerCalled).To(Equal(1))
					Expect(numFallbackCalled).To(Equal(1))
				})
			})
		})

		Context("when no handler except for fallback is registered", func() {
			It("calls fallback handler", func() {
				numCalled := 0
				r.SetFallback(eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numCalled++
					return nil
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(numCalled).To(Equal(1))
			})
		})

		Context("when more than one fallback handlers are registered", func() {
			It("uses the last one", func() {
				numFirstHandlerCalled := 0
				r.SetFallback(eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numFirstHandlerCalled++
					return nil
				}))
				numLastHandlerCalled := 0
				r.SetFallback(eventrouter.HandlerFunc(func(_ *slackevents.EventsAPIEvent) error {
					numLastHandlerCalled++
					return nil
				}))
				req, err := http.NewRequest(http.MethodPost, "http:/example.com/path", bytes.NewReader([]byte(content)))
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
				Expect(numFirstHandlerCalled).To(Equal(0))
				Expect(numLastHandlerCalled).To(Equal(1))
			})
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
	if err := testutils.AddSignature(req.Header, []byte(signingSecret), []byte(body), now); err != nil {
		return nil, err
	}
	return req, nil
}
