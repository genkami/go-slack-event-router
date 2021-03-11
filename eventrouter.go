package eventrouter

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/reaction"
	"github.com/genkami/go-slack-event-router/urlverification"
)

type FallbackHandler interface {
	HandleEventsAPIEvent(e *slackevents.EventsAPIEvent) error
}

type Router struct {
	signingToken            string
	appMentionHandlers      []appmention.Handler
	reactionAddedHandlers   []reaction.AddedHandler
	reactionRemovedHandlers []reaction.RemovedHandler
	urlVerificationHandler  urlverification.Handler
	fallbackHandler         FallbackHandler
}

func New(signingToken string) *Router {
	return &Router{
		signingToken:            signingToken,
		appMentionHandlers:      make([]appmention.Handler, 0),
		reactionAddedHandlers:   make([]reaction.AddedHandler, 0),
		reactionRemovedHandlers: make([]reaction.RemovedHandler, 0),
		urlVerificationHandler:  urlverification.DefaultHandler,
		fallbackHandler:         nil,
	}
}

func (r *Router) OnAppMention(h appmention.Handler, preds ...appmention.Predicate) {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	r.appMentionHandlers = append(r.appMentionHandlers, h)
}

func (r *Router) OnReactionAdded(h reaction.AddedHandler, preds ...reaction.Predicate) {
	for _, p := range preds {
		h = p.WrapAdded(h)
	}
	r.reactionAddedHandlers = append(r.reactionAddedHandlers, h)
}

func (r *Router) OnReactionRemoved(h reaction.RemovedHandler, preds ...reaction.Predicate) {
	for _, p := range preds {
		h = p.WrapRemoved(h)
	}
	r.reactionRemovedHandlers = append(r.reactionRemovedHandlers, h)
}

func (r *Router) SetURLVerificationHandler(h urlverification.Handler) {
	r.urlVerificationHandler = h
}

func (r *Router) SetFallback(h FallbackHandler) {
	r.fallbackHandler = h
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	// TODO: check signature
	router.serveHTTP(w, req)
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
		return
	case slackevents.CallbackEvent:
		router.handleCallbackEvent(w, &eventsAPIEvent)
		return
	case slackevents.AppRateLimited:
		router.handleAppRateLimited(w, &eventsAPIEvent)
		return
	}
}

func (r *Router) handleURLVerification(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	_, ok := e.Data.(*slackevents.EventsAPIURLVerificationEvent)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO: implement
	w.Write([]byte("OK"))
}

func (r *Router) handleCallbackEvent(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	var err error
	switch inner := e.InnerEvent.Data.(type) {
	case *slackevents.AppMentionEvent:
		err = r.handleAppMentionEvent(inner)
	case *slackevents.ReactionAddedEvent:
		err = r.handleReactionAddedEvent(inner)
	case *slackevents.ReactionRemovedEvent:
		err = r.handleReactionRemovedEvent(inner)
	default:
		// TODO: implemtn all event handlers
		err = errors.NotInterested
	}
	if err == errors.NotInterested {
		err = r.handleFallback(e)
	}
	if err != nil {
		// TODO: handle errors.HttpError
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleAppRateLimited(w http.ResponseWriter, e *slackevents.EventsAPIEvent) {
	// TODO: implement
	w.Write([]byte("OK"))
}

func (r *Router) handleAppMentionEvent(e *slackevents.AppMentionEvent) error {
	// TODO: implement
	return nil
}

func (r *Router) handleReactionAddedEvent(e *slackevents.ReactionAddedEvent) error {
	// TODO: implement
	return nil
}

func (r *Router) handleReactionRemovedEvent(e *slackevents.ReactionRemovedEvent) error {
	// TODO: implement
	return nil
}

func (r *Router) handleFallback(e *slackevents.EventsAPIEvent) error {
	// TODO: implement
	return nil
}
