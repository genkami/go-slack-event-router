package interactionrouter

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/slack-go/slack"

	routererrors "github.com/genkami/go-slack-event-router/errors"
)

type Handler interface {
	HandleInteraction(*slack.InteractionCallback) error
}

type HandlerFunc func(*slack.InteractionCallback) error

func (f HandlerFunc) HandleInteraction(c *slack.InteractionCallback) error {
	return f(c)
}

type Predicate interface {
	Wrap(Handler) Handler
}

type typePredicate struct {
	typeName slack.InteractionType
}

func Type(typeName slack.InteractionType) Predicate {
	return &typePredicate{typeName: typeName}
}

func (p *typePredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(callback *slack.InteractionCallback) error {
		if callback.Type != p.typeName {
			return routererrors.NotInterested
		}
		return h.HandleInteraction(callback)
	})
}

type blockActionPredicate struct {
	blockID  string
	actionID string
}

func BlockAction(blockID, actionID string) Predicate {
	return &blockActionPredicate{blockID: blockID, actionID: actionID}
}

func (p *blockActionPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(callback *slack.InteractionCallback) error {
		for _, ba := range callback.ActionCallback.BlockActions {
			if ba.BlockID == p.blockID && ba.ActionID == p.actionID {
				return nil
			}
		}
		return routererrors.NotInterested
	})
}

func Build(h Handler, preds ...Predicate) Handler {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	return h
}

type Option interface {
	apply(*Router)
}

type optionFunc func(*Router)

func (f optionFunc) apply(r *Router) {
	f(r)
}

type Router struct {
	// TODO: check signature
	handlers        map[slack.InteractionType][]Handler
	fallbackHandler Handler
}

func (r *Router) On(typeName slack.InteractionType, h Handler, preds ...Predicate) {
	h = Build(h, preds...)
	handlers, ok := r.handlers[typeName]
	if !ok {
		handlers = make([]Handler, 0)
	}
	handlers = append(handlers, h)
	r.handlers[typeName] = handlers
}

func (router *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	callback := slack.InteractionCallback{}
	if err := json.Unmarshal([]byte(req.FormValue("payload")), &callback); err != nil {
		router.respondWithError(w, err)
		return
	}

	router.handleInteractionCallback(w, &callback)
}

func (r *Router) handleInteractionCallback(w http.ResponseWriter, callback *slack.InteractionCallback) {
	var err error = routererrors.NotInterested
	handlers, ok := r.handlers[callback.Type]
	if ok {
		for _, h := range handlers {
			err = h.HandleInteraction(callback)
			if !errors.Is(err, routererrors.NotInterested) {
				break
			}
		}
	}

	if errors.Is(err, routererrors.NotInterested) {
		err = r.handleFallback(callback)
	}

	if err != nil && !errors.Is(err, routererrors.NotInterested) {
		r.respondWithError(w, err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (r *Router) handleFallback(callback *slack.InteractionCallback) error {
	if r.fallbackHandler == nil {
		return routererrors.NotInterested
	}
	return r.fallbackHandler.HandleInteraction(callback)
}

func (r *Router) respondWithError(w http.ResponseWriter, err error) {
	// TODO: verbose output
	// TODO: move this to utils
	w.WriteHeader(http.StatusInternalServerError)
}
