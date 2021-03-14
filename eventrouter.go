// Package eventrouter provides a way to dispatch events from Slack.
package eventrouter

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/appratelimited"
	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/internal/routerutils"
	"github.com/genkami/go-slack-event-router/message"
	"github.com/genkami/go-slack-event-router/reaction"
	"github.com/genkami/go-slack-event-router/signature"
	"github.com/genkami/go-slack-event-router/urlverification"
)

// Handler is a handler that processes events from Slack.
// Usually you don't need to use this directly. Instead, you might want to use event-specific handler types like `appmention.Handler`.
//
// Handlers may return `routererrors.NotInterested` (or its equivalents in the sense of `errors.Is`). In such case the Router falls back to other handlers.
//
// Handlers also may return `routererrors.HttpError` (or its equivalents in the sense of `errors.Is`). In such case the Router responds with corresponding HTTP status codes.
//
// If any other errors are returned, the Router responds with Internal Server Error.
type Handler interface {
	HandleEventsAPIEvent(context.Context, *slackevents.EventsAPIEvent) error
}

type HandlerFunc func(context.Context, *slackevents.EventsAPIEvent) error

func (f HandlerFunc) HandleEventsAPIEvent(ctx context.Context, e *slackevents.EventsAPIEvent) error {
	return f(ctx, e)
}

// Option configures the Router.
type Option interface {
	apply(*Router)
}

type optionFunc func(*Router)

func (f optionFunc) apply(r *Router) {
	f(r)
}

// InsecureSkipVerification skips verifying request signatures.
// This is useful to test your handlers, but do not use this in production environments.
func InsecureSkipVerification() Option {
	return optionFunc(func(r *Router) {
		r.skipVerification = true
	})
}

// WithSigningSecret sets a signing token to verify requests from Slack.
//
// For more details, see https://api.slack.com/authentication/verifying-requests-from-slack.
func WithSigningSecret(token string) Option {
	return optionFunc(func(r *Router) {
		r.signingSecret = token
	})
}

// If VerboseResponse is set, the Router shows error details when it fails to process requests.
func VerboseResponse() Option {
	return optionFunc(func(r *Router) {
		r.verboseResponse = true
	})
}

// Router is an http.Handler that processes events from Slack via Events API.
//
// For more details, see https://api.slack.com/apis/connections/events-api.
type Router struct {
	signingSecret          string
	skipVerification       bool
	verboseResponse        bool
	callbackHandlers       map[string][]Handler
	urlVerificationHandler urlverification.Handler
	appRateLimitedHandler  appratelimited.Handler
	fallbackHandler        Handler
	httpHandler            http.Handler
}

// New creates a new Router.
//
// At least one of WithSigningSecret() or InsecureSkipVerification() must be specified.
func New(options ...Option) (*Router, error) {
	r := &Router{
		callbackHandlers:       make(map[string][]Handler),
		urlVerificationHandler: urlverification.DefaultHandler,
		appRateLimitedHandler:  appratelimited.DefaultHandler,
	}
	for _, o := range options {
		o.apply(r)
	}
	if r.signingSecret == "" && !r.skipVerification {
		return nil, errors.New("WithSigningSecret must be set, or you can ignore this by setting InsecureSkipVerification")
	}
	if r.signingSecret != "" && r.skipVerification {
		return nil, errors.New("both WithSigningSecret and InsecureSkipVerification are given")
	}

	r.httpHandler = http.HandlerFunc(r.serveHTTP)
	if !r.skipVerification {
		r.httpHandler = &signature.Middleware{
			SigningSecret:   r.signingSecret,
			VerboseResponse: r.verboseResponse,
			Handler:         r.httpHandler,
		}
	}
	return r, nil
}

// On registers a handler for a specific event type.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Handlers may return `routererrors.NotInterested` (or its equivalents in the sense of `errors.Is`). In such case the Router falls back to other handlers.
//
// Handlers also may return `routererrors.HttpError` (or its equivalents in the sense of `errors.Is`). In such case the Router responds with corresponding HTTP status codes.
//
// If any other errors are returned, the Router responds with Internal Server Error.
//
// This can be useful if you have a general-purpose event handlers that can process arbitrary types of events,
// but, in the most cases it would be better option to use event-specfic `OnEVENT_NAME` methods instead.
func (r *Router) On(eventType string, h Handler) {
	handlers, ok := r.callbackHandlers[eventType]
	if !ok {
		handlers = make([]Handler, 0)
	}
	handlers = append(handlers, h)
	r.callbackHandlers[eventType] = handlers
}

// OnMessage registers a handler that processes `message` events.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Predicates are used to distinguish whether a coming event should be processed by the given handler or not.
// The handler `h` will be called only when all of given Predicates are true.
func (r *Router) OnMessage(h message.Handler, preds ...message.Predicate) {
	h = message.Build(h, preds...)
	r.On(slackevents.Message, HandlerFunc(func(ctx context.Context, e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.MessageEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleMessageEvent(ctx, inner)
	}))
}

// OnAppMention registers a handler that processes `app_mention` events.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Predicates are used to distinguish whether a coming event should be processed by the given handler or not.
// The handler `h` will be called only when all of given Predicates are true.
func (r *Router) OnAppMention(h appmention.Handler, preds ...appmention.Predicate) {
	h = appmention.Build(h, preds...)
	r.On(slackevents.AppMention, HandlerFunc(func(ctx context.Context, e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleAppMentionEvent(ctx, inner)
	}))
}

// OnReactionAdded registers a handler that processes `reaction_added` events.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Predicates are used to distinguish whether a coming event should be processed by the given handler or not.
// The handler `h` will be called only when all of given Predicates are true.
func (r *Router) OnReactionAdded(h reaction.AddedHandler, preds ...reaction.Predicate) {
	h = reaction.BuildAdded(h, preds...)
	r.On(slackevents.ReactionAdded, HandlerFunc(func(ctx context.Context, e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.ReactionAddedEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleReactionAddedEvent(ctx, inner)
	}))
}

// OnReactionRemoved registers a handler that processes `reaction_removed` events.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Predicates are used to distinguish whether a coming event should be processed by the given handler or not.
// The handler `h` will be called only when all of given Predicates are true.
func (r *Router) OnReactionRemoved(h reaction.RemovedHandler, preds ...reaction.Predicate) {
	h = reaction.BuildRemoved(h, preds...)
	r.On(slackevents.ReactionRemoved, HandlerFunc(func(ctx context.Context, e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.ReactionRemovedEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleReactionRemovedEvent(ctx, inner)
	}))
}

// SetURLVerificationHandler sets a handler to process `url_verification` events.
//
// If more than one handlers are registered, the last one will be used.
//
// If no handler is set explicitly, the Rotuer uses the default handler.
//
// For more details see https://api.slack.com/events/url_verification.
func (r *Router) SetURLVerificationHandler(h urlverification.Handler) {
	r.urlVerificationHandler = h
}

// SetAppRateLimitedHandler sets a handler to process `app_rate_limited` events.
//
// If more than one handlers are registered, the last one will be used.
//
// If no handler is set explicitly, the Rotuer uses the default handler that simply ignores events of this type.
//
// For more details see https://api.slack.com/docs/rate-limits#rate-limits__events-api.
func (r *Router) SetAppRateLimitedHandler(h appratelimited.Handler) {
	r.appRateLimitedHandler = h
}

// SetFallback sets a fallback handler that is called when none of the registered handlers matches to a coming event.
//
// If more than one handlers are registered, the last one will be used.
func (r *Router) SetFallback(h Handler) {
	r.fallbackHandler = h
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.httpHandler.ServeHTTP(w, req)
}

func (router *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		router.respondWithError(w, err)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		router.respondWithError(
			w,
			errors.WithMessage(routererrors.HttpError(http.StatusBadRequest), err.Error()))
		return
	}

	ctx := req.Context()
	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		router.handleURLVerification(ctx, w, &eventsAPIEvent)
	case slackevents.CallbackEvent:
		router.handleCallbackEvent(ctx, w, &eventsAPIEvent)
	case slackevents.AppRateLimited:
		// Surprisingly, ParseEvent can't deal with EventsAPIAppRateLimitedEvent correctly.
		// So we should re-parse the entire body for now.
		appRateLimited := slackevents.EventsAPIAppRateLimited{}
		err := json.Unmarshal(body, &appRateLimited)
		if err != nil {
			router.respondWithError(
				w,
				errors.WithMessage(err, "failed to parse app_rate_limited event"))
		}
		router.handleAppRateLimited(ctx, w, &appRateLimited)
	default:
		router.respondWithError(
			w,
			errors.WithMessagef(routererrors.HttpError(http.StatusBadRequest),
				"unknown event type: %s", eventsAPIEvent.Type))
	}
}

func (r *Router) handleURLVerification(ctx context.Context, w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	ev, ok := e.Data.(*slackevents.EventsAPIURLVerificationEvent)
	if !ok {
		r.respondWithError(w, fmt.Errorf("expected EventsAPIURLVerificationEvent but got %T", e.Data))
		return
	}
	resp, err := r.urlVerificationHandler.HandleURLVerification(ctx, ev)
	if err != nil {
		r.respondWithError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(resp)
}

func (r *Router) handleCallbackEvent(ctx context.Context, w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	var err error = routererrors.NotInterested
	handlers, ok := r.callbackHandlers[e.InnerEvent.Type]
	if ok {
		for _, h := range handlers {
			err = h.HandleEventsAPIEvent(ctx, e)
			if !errors.Is(err, routererrors.NotInterested) {
				break
			}
		}
	}

	if errors.Is(err, routererrors.NotInterested) {
		err = r.handleFallback(ctx, e)
	}

	if err != nil && !errors.Is(err, routererrors.NotInterested) {
		r.respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleAppRateLimited(ctx context.Context, w http.ResponseWriter, e *slackevents.EventsAPIAppRateLimited) {
	err := r.appRateLimitedHandler.HandleAppRateLimited(ctx, e)
	if err != nil {
		r.respondWithError(w, err)
		return
	}
	_, _ = w.Write([]byte("OK"))
}

func (r *Router) handleFallback(ctx context.Context, e *slackevents.EventsAPIEvent) error {
	if r.fallbackHandler == nil {
		return routererrors.NotInterested
	}
	return r.fallbackHandler.HandleEventsAPIEvent(ctx, e)
}

func (r *Router) respondWithError(w http.ResponseWriter, err error) {
	routerutils.RespondWithError(w, err, r.verboseResponse)
}
