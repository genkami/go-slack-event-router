package interactionrouter_test

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/pkg/errors"
	"github.com/slack-go/slack"

	routererrors "github.com/genkami/go-slack-event-router/errors"
	ir "github.com/genkami/go-slack-event-router/interactionrouter"
	"github.com/genkami/go-slack-event-router/signature"
)

var _ = Describe("InteractionRouter", func() {
	Describe("Type", func() {
		var (
			numHandlerCalled int
			innerHandler     = ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
		})

		Context("when the type of the interaction callback matches to the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.Type(slack.InteractionTypeBlockActions).Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the type of the interaction callback differs from the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.Type(slack.InteractionTypeBlockActions).Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeViewSubmission,
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})

	Describe("BlockAction", func() {
		var (
			numHandlerCalled int
			innerHandler     = ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
		})

		Context("when the interaction callback has the block_action specified by the predicate", func() {
			It("calls the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when one of the block_acsions that the interaction callback has is the one specified by the predicate", func() {
			It("calls the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
							{BlockID: "BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the interaction callback does not have any block_action", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the block_action in the interaction callback is not what the predicate expects", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the block_id in the block_action is the same as the predicate expected but the action_id isn't", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the action_id in the block_action is the same as the predicate expected but the block_id isn't", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})

	Describe("CallbackID", func() {
		var (
			numHandlerCalled int
			innerHandler     = ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
		})

		Context("when the callback_id in the interaction callback matches to the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.CallbackID("CALLBACK_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type:       slack.InteractionTypeBlockActions,
					CallbackID: "CALLBACK_ID",
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the callback_id in the interaction callback differs from the predicate's", func() {
			It("does not call the inner handler", func() {
				h := ir.CallbackID("CALLBACK_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type:       slack.InteractionTypeBlockActions,
					CallbackID: "ANOTHER_CALLBACK_ID",
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})

	Describe("New", func() {
		Context("when neither WithSigningToken nor InsecureSkipVerification is given", func() {
			It("returns an error", func() {
				_, err := ir.New()
				Expect(err).To(MatchError(MatchRegexp("WithSigningToken")))
			})
		})

		Context("when InsecureSkipVerification is given", func() {
			It("returns a new Router", func() {
				r, err := ir.New(ir.InsecureSkipVerification())
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			})
		})

		Context("when WithSigningToken is given", func() {
			It("returns a new Router", func() {
				r, err := ir.New(ir.WithSigningToken("THE_TOKEN"))
				Expect(err).NotTo(HaveOccurred())
				Expect(r).NotTo(BeNil())
			})
		})

		Context("when both WithSigningToken and InsecureSkipVerification are given", func() {
			It("returns an error", func() {
				_, err := ir.New(ir.InsecureSkipVerification(), ir.WithSigningToken("THE_TOKEN"))
				Expect(err).To(MatchError(MatchRegexp("WithSigningToken")))
			})
		})
	})

	Describe("WithSigningSecret", func() {
		var (
			r       *ir.Router
			token   = "THE_TOKEN"
			content = `
			{
				"type": "shortcut",
				"token": "XXXXXXXXXXXXX",
				"action_ts": "1581106241.371594",
				"team": {
				  "id": "TXXXXXXXX",
				  "domain": "shortcuts-test"
				},
				"user": {
				  "id": "UXXXXXXXXX",
				  "username": "aman",
				  "team_id": "TXXXXXXXX"
				},
				"callback_id": "shortcut_create_task",
				"trigger_id": "944799105734.773906753841.38b5894552bdd4a780554ee59d1f3638"
			}`
		)
		BeforeEach(func() {
			var err error
			r, err = ir.New(ir.WithSigningToken(token), ir.VerboseResponse())
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
			r       *ir.Router
			token   = "THE_TOKEN"
			content = `
			{
				"type": "shortcut",
				"token": "XXXXXXXXXXXXX",
				"action_ts": "1581106241.371594",
				"team": {
				  "id": "TXXXXXXXX",
				  "domain": "shortcuts-test"
				},
				"user": {
				  "id": "UXXXXXXXXX",
				  "username": "aman",
				  "team_id": "TXXXXXXXX"
				},
				"callback_id": "shortcut_create_task",
				"trigger_id": "944799105734.773906753841.38b5894552bdd4a780554ee59d1f3638"
			}`
		)
		BeforeEach(func() {
			var err error
			r, err = ir.New(ir.InsecureSkipVerification(), ir.VerboseResponse())
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

	Describe("On", func() {
		var (
			r       *ir.Router
			content = `
			{
				"type": "shortcut",
				"token": "XXXXXXXXXXXXX",
				"action_ts": "1581106241.371594",
				"team": {
				  "id": "TXXXXXXXX",
				  "domain": "shortcuts-test"
				},
				"user": {
				  "id": "UXXXXXXXXX",
				  "username": "aman",
				  "team_id": "TXXXXXXXX"
				},
				"callback_id": "shortcut_create_task",
				"trigger_id": "944799105734.773906753841.38b5894552bdd4a780554ee59d1f3638"
			}`
			numHandlerCalled = 0
			handler          = ir.HandlerFunc(func(e *slack.InteractionCallback) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
			var err error
			r, err = ir.New(ir.InsecureSkipVerification(), ir.VerboseResponse())
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when no handler is registered", func() {
			It("just responds with 200", func() {
				req, err := NewRequest(content)
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
				r.On(slack.InteractionTypeShortcut, handler)
				req, err := NewRequest(content)
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
				r.On("other_interaction_type", handler)
				req, err := NewRequest(content)
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
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					return fmt.Errorf("something wrong happened")
				}))
				req, err := NewRequest(content)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusInternalServerError))
			})
		})

		Context("when a handler returned NotInterested", func() {
			It("responds with 200", func() {
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					return routererrors.NotInterested
				}))
				req, err := NewRequest(content)
				Expect(err).NotTo(HaveOccurred())
				w := httptest.NewRecorder()
				r.ServeHTTP(w, req)
				resp := w.Result()
				Expect(resp.StatusCode).To(Equal(http.StatusOK))
			})
		})

		Context("when a handler returned an error that equals to NotInterested using errors.Is", func() {
			It("responds with 200", func() {
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					return errors.WithMessage(routererrors.NotInterested, "not interested")
				}))
				req, err := NewRequest(content)
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
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					return routererrors.HttpError(code)
				}))
				req, err := NewRequest(content)
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
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					return errors.WithMessage(routererrors.HttpError(code), "you ain't authorized")
				}))
				req, err := NewRequest(content)
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
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numFirstHandlerCalled++
					return firstError
				}))
				r.On(slack.InteractionTypeShortcut, ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numSecondHandlerCalled++
					return secondError
				}))
				r.SetFallback(ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numFallbackCalled++
					return fallbackError
				}))
			})

			Context("when a first handler returned nil", func() {
				It("does not fall back to other handlers", func() {
					firstError = nil
					secondError = nil
					fallbackError = nil
					req, err := NewRequest(content)
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
					req, err := NewRequest(content)
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
					req, err := NewRequest(content)
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
					req, err := NewRequest(content)
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
					req, err := NewRequest(content)
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
					req, err := NewRequest(content)
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
				r.SetFallback(ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numCalled++
					return nil
				}))
				req, err := NewRequest(content)
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
				r.SetFallback(ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numFirstHandlerCalled++
					return nil
				}))
				numLastHandlerCalled := 0
				r.SetFallback(ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
					numLastHandlerCalled++
					return nil
				}))
				req, err := NewRequest(content)
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

func NewRequest(payload string) (*http.Request, error) {
	body := buildRequestBody(payload)
	req, err := http.NewRequest(http.MethodPost, "http://example.com/path/to/callback", bytes.NewReader([]byte(body)))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	return req, nil
}

func NewSignedRequest(signingSecret string, payload string, ts *time.Time) (*http.Request, error) {
	var now time.Time
	if ts == nil {
		now = time.Now()
	} else {
		now = *ts
	}
	req, err := NewRequest(payload)
	if err != nil {
		return nil, err
	}
	body := buildRequestBody(payload)
	if err := signature.AddSignature(req.Header, []byte(signingSecret), []byte(body), now); err != nil {
		return nil, err
	}
	return req, nil
}

func buildRequestBody(payload string) []byte {
	form := url.Values{}
	form.Set("payload", payload)
	return []byte(form.Encode())
}
