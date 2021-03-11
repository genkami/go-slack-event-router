package eventrouter

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/appratelimited"
	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/message"
	"github.com/genkami/go-slack-event-router/reaction"
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

type Router struct {
	signingToken           string
	skipVerification       bool
	callbackHandlers       map[string][]Handler
	urlVerificationHandler urlverification.Handler
	appRateLimitedHandler  appratelimited.Handler
	fallbackHandler        Handler
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
	for _, p := range preds {
		h = p.Wrap(h)
	}
	r.On(slackevents.Message, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.MessageEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleMessageEvent(inner)
	}))
}

func (r *Router) OnAppMention(h appmention.Handler, preds ...appmention.Predicate) {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	r.On(slackevents.AppMention, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.AppMentionEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleAppMentionEvent(inner)
	}))
}

func (r *Router) OnReactionAdded(h reaction.AddedHandler, preds ...reaction.Predicate) {
	for _, p := range preds {
		h = p.WrapAdded(h)
	}
	r.On(slackevents.ReactionAdded, HandlerFunc(func(e *slackevents.EventsAPIEvent) error {
		inner, ok := e.InnerEvent.Data.(*slackevents.ReactionAddedEvent)
		if !ok {
			return routererrors.HttpError(http.StatusBadRequest)
		}
		return h.HandleReactionAddedEvent(inner)
	}))
}

func (r *Router) OnReactionRemoved(h reaction.RemovedHandler, preds ...reaction.Predicate) {
	for _, p := range preds {
		h = p.WrapRemoved(h)
	}
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
	if router.skipVerification {
		router.serveHTTP(w, req)
	} else {
		signature.Middleware(router.signingToken, http.HandlerFunc(router.serveHTTP)).ServeHTTP(w, req)
	}
}

func (router *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	eventsAPIEvent, err := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionNoVerifyToken())
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	switch eventsAPIEvent.Type {
	case slackevents.URLVerification:
		router.handleURLVerification(w, &eventsAPIEvent)
	case slackevents.CallbackEvent:
		router.handleCallbackEvent(w, &eventsAPIEvent)
	case slackevents.AppRateLimited:
		router.handleAppRateLimited(w, &eventsAPIEvent)
	default:
		w.WriteHeader(http.StatusBadRequest)
	}
}

func (r *Router) handleURLVerification(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	ev, ok := e.Data.(*slackevents.EventsAPIURLVerificationEvent)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
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
	var err error
	handlers, ok := r.callbackHandlers[e.InnerEvent.Type]
	if !ok {
		err = routererrors.NotInterested
	} else {
		for _, h := range handlers {
			err = h.HandleEventsAPIEvent(e)
			if err != routererrors.NotInterested {
				break
			}
		}
	}

	if err == routererrors.NotInterested {
		err = r.handleFallback(e)
	}

	if err != nil && err != routererrors.NotInterested {
		r.respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleAppRateLimited(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	ev, ok := e.Data.(*slackevents.EventsAPIAppRateLimited)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	err := r.appRateLimitedHandler.HandleAppRateLimited(ev)
	if err != nil {
		r.respondWithError(w, err)
		return
	}
	w.Write([]byte("OK"))
}

func (r *Router) handleFallback(e *slackevents.EventsAPIEvent) error {
	if r.fallbackHandler == nil {
		return routererrors.NotInterested
	}
	return r.fallbackHandler.HandleEventsAPIEvent(e)
}

func (r *Router) respondWithError(w http.ResponseWriter, err error) {
	var httpErr *routererrors.HttpError
	if errors.As(err, &httpErr) {
		w.WriteHeader(int(*httpErr))
	} else {
		w.WriteHeader(http.StatusInternalServerError)
	}
}
