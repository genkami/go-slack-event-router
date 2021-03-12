package eventrouter

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/pkg/errors"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/appratelimited"
	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/message"
	"github.com/genkami/go-slack-event-router/reaction"
	"github.com/genkami/go-slack-event-router/routerutils"
	"github.com/genkami/go-slack-event-router/signature"
	"github.com/genkami/go-slack-event-router/urlverification"
)

type Handler interface {
	HandleEventsAPIEvent(*slackevents.EventsAPIEvent) error
}

type HandlerFunc func(*slackevents.EventsAPIEvent) error

func (f HandlerFunc) HandleEventsAPIEvent(e *slackevents.EventsAPIEvent) error {
	return f(e)
}

type Option interface {
	apply(*Router)
}

type optionFunc func(*Router)

func (f optionFunc) apply(r *Router) {
	f(r)
}

func InsecureSkipVerification() Option {
	return optionFunc(func(r *Router) {
		r.skipVerification = true
	})
}

func WithSigningToken(token string) Option {
	return optionFunc(func(r *Router) {
		r.signingToken = token
	})
}

func VerboseResponse() Option {
	return optionFunc(func(r *Router) {
		r.verboseResponse = true
	})
}

type Router struct {
	signingToken           string
	skipVerification       bool
	verboseResponse        bool
	callbackHandlers       map[string][]Handler
	urlVerificationHandler urlverification.Handler
	appRateLimitedHandler  appratelimited.Handler
	fallbackHandler        Handler
	httpHandler            http.Handler
}

func New(options ...Option) (*Router, error) {
	r := &Router{
		callbackHandlers:       make(map[string][]Handler),
		urlVerificationHandler: urlverification.DefaultHandler,
		appRateLimitedHandler:  appratelimited.DefaultHandler,
	}
	for _, o := range options {
		o.apply(r)
	}
	if r.signingToken == "" && !r.skipVerification {
		return nil, errors.New("WithSigningToken must be set, or you can ignore this by setting InsecureSkipVerification")
	}
	if r.signingToken != "" && r.skipVerification {
		return nil, errors.New("both WithSigningToken and InsecureSkipVerification are given")
	}

	r.httpHandler = http.HandlerFunc(r.serveHTTP)
	if !r.skipVerification {
		r.httpHandler = &signature.Middleware{
			Secret:          r.signingToken,
			VerboseResponse: r.verboseResponse,
			Handler:         r.httpHandler,
		}
	}
	return r, nil
}

func (r *Router) On(eventType string, h Handler) {
	handlers, ok := r.callbackHandlers[eventType]
	if !ok {
		handlers = make([]Handler, 0)
	}
	handlers = append(handlers, h)
	r.callbackHandlers[eventType] = handlers
}

func (r *Router) OnMessage(h message.Handler, preds ...message.Predicate) {
	h = message.Build(h, preds...)
	r.On(slackevents.Message, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.MessageEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleMessageEvent(inner)
	}))
}

func (r *Router) OnAppMention(h appmention.Handler, preds ...appmention.Predicate) {
	h = appmention.Build(h, preds...)
	r.On(slackevents.AppMention, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleAppMentionEvent(inner)
	}))
}

func (r *Router) OnReactionAdded(h reaction.AddedHandler, preds ...reaction.Predicate) {
	h = reaction.BuildAdded(h, preds...)
	r.On(slackevents.ReactionAdded, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.ReactionAddedEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleReactionAddedEvent(inner)
	}))
}

func (r *Router) OnReactionRemoved(h reaction.RemovedHandler, preds ...reaction.Predicate) {
	h = reaction.BuildRemoved(h, preds...)
	r.On(slackevents.ReactionRemoved, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.ReactionRemovedEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleReactionRemovedEvent(inner)
	}))
}

func (r *Router) SetURLVerificationHandler(h urlverification.Handler) {
	r.urlVerificationHandler = h
}

func (r *Router) SetAppRateLimitedHandler(h appratelimited.Handler) {
	r.appRateLimitedHandler = h
}

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

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		router.handleURLVerification(w, &eventsAPIEvent)
	case slackevents.CallbackEvent:
		router.handleCallbackEvent(w, &eventsAPIEvent)
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
		router.handleAppRateLimited(w, &appRateLimited)
	default:
		router.respondWithError(
			w,
			errors.WithMessagef(routererrors.HttpError(http.StatusBadRequest),
				"unknown event type: %s", eventsAPIEvent.Type))
	}
}

func (r *Router) handleURLVerification(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	ev, ok := e.Data.(*slackevents.EventsAPIURLVerificationEvent)
	if !ok {
		r.respondWithError(w, fmt.Errorf("expected EventsAPIURLVerificationEvent but got %T", e.Data))
		return
	}
	resp, err := r.urlVerificationHandler.HandleURLVerification(ev)
	if err != nil {
		r.respondWithError(w, err)
		return
	}
	w.Header().Add("Content-Type", "application/json")
	enc := json.NewEncoder(w)
	_ = enc.Encode(resp)
}

func (r *Router) handleCallbackEvent(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	var err error = routererrors.NotInterested
	handlers, ok := r.callbackHandlers[e.InnerEvent.Type]
	if ok {
		for _, h := range handlers {
			err = h.HandleEventsAPIEvent(e)
			if !errors.Is(err, routererrors.NotInterested) {
				break
			}
		}
	}

	if errors.Is(err, routererrors.NotInterested) {
		err = r.handleFallback(e)
	}

	if err != nil && !errors.Is(err, routererrors.NotInterested) {
		r.respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleAppRateLimited(w http.ResponseWriter, e *slackevents.EventsAPIAppRateLimited) {
	err := r.appRateLimitedHandler.HandleAppRateLimited(e)
	if err != nil {
		r.respondWithError(w, err)
		return
	}
	_, _ = w.Write([]byte("OK"))
}

func (r *Router) handleFallback(e *slackevents.EventsAPIEvent) error {
	if r.fallbackHandler == nil {
		return routererrors.NotInterested
	}
	return r.fallbackHandler.HandleEventsAPIEvent(e)
}

func (r *Router) respondWithError(w http.ResponseWriter, err error) {
	routerutils.RespondWithError(w, err, r.verboseResponse)
}
