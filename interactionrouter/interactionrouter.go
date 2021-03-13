package interactionrouter

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"

	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/routerutils"
	"github.com/genkami/go-slack-event-router/signature"
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
				return h.HandleInteraction(callback)
			}
		}
		return routererrors.NotInterested
	})
}

type callbackIDPredicate struct {
	id string
}

func CallbackID(id string) Predicate {
	return &callbackIDPredicate{id: id}
}

func (p *callbackIDPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(callback *slack.InteractionCallback) error {
		if callback.CallbackID != p.id {
			return routererrors.NotInterested
		}
		return h.HandleInteraction(callback)
	})
}

type channelPredicate struct {
	id string
}

func Channel(id string) Predicate {
	return &channelPredicate{id: id}
}

func (p *channelPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(callback *slack.InteractionCallback) error {
		if callback.Channel.ID != p.id {
			return routererrors.NotInterested
		}
		return h.HandleInteraction(callback)
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
	signingToken     string
	skipVerification bool
	handlers         map[slack.InteractionType][]Handler
	fallbackHandler  Handler
	verboseResponse  bool
	httpHandler      http.Handler
}

func New(opts ...Option) (*Router, error) {
	r := &Router{
		handlers: make(map[slack.InteractionType][]Handler),
	}
	for _, o := range opts {
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

func (r *Router) On(typeName slack.InteractionType, h Handler, preds ...Predicate) {
	h = Build(h, preds...)
	handlers, ok := r.handlers[typeName]
	if !ok {
		handlers = make([]Handler, 0)
	}
	handlers = append(handlers, h)
	r.handlers[typeName] = handlers
}

func (router *Router) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	router.httpHandler.ServeHTTP(w, req)
}

func (router *Router) serveHTTP(w http.ResponseWriter, req *http.Request) {
	callback := slack.InteractionCallback{}
	if req.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
		router.respondWithError(w,
			errors.WithMessage(routererrors.HttpError(http.StatusBadRequest), "unexpected Content-Type"))
		return
	}
	payload := req.FormValue("payload")
	if payload == "" {
		router.respondWithError(w,
			errors.WithMessage(routererrors.HttpError(http.StatusBadRequest), "missing payload"))
		return
	}
	if err := json.Unmarshal([]byte(payload), &callback); err != nil {
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
	routerutils.RespondWithError(w, err, r.verboseResponse)
}

func FindBlockAction(callback *slack.InteractionCallback, blockID, actionID string) *slack.BlockAction {
	for _, ba := range callback.ActionCallback.BlockActions {
		if ba.BlockID == blockID && ba.ActionID == actionID {
			return ba
		}
	}
	return nil
}
