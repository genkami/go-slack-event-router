// Package interactionrouter provides a way to dispatch interactive callbacks sent from Slack.
//
// For more details, see https://api.slack.com/interactivity/handling.
package interactionrouter

import (
	"encoding/json"
	"net/http"

	"github.com/pkg/errors"
	"github.com/slack-go/slack"

	routererrors "github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/internal/routerutils"
	"github.com/genkami/go-slack-event-router/signature"
)

// Handler processes interaction callbacks sent from Slack.
type Handler interface {
	HandleInteraction(*slack.InteractionCallback) error
}

type HandlerFunc func(*slack.InteractionCallback) error

func (f HandlerFunc) HandleInteraction(c *slack.InteractionCallback) error {
	return f(c)
}

// Predicate disthinguishes whether or not a certain handler should process coming events.
type Predicate interface {
	Wrap(Handler) Handler
}

type typePredicate struct {
	typeName slack.InteractionType
}

// Type is a predicate that is considered to be "true" if and only if the type of the InteractionCallback equals to the given one.
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

// BlockAction is a predicate that is considered to be "true" if and only if the InteractionCallback has a BlockAction identified by blockID and actionID.
func BlockAction(blockID, actionID string) Predicate {
	return &blockActionPredicate{blockID: blockID, actionID: actionID}
}

func (p *blockActionPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(callback *slack.InteractionCallback) error {
		if FindBlockAction(callback, p.blockID, p.actionID) == nil {
			return routererrors.NotInterested
		}
		return h.HandleInteraction(callback)
	})
}

type callbackIDPredicate struct {
	id string
}

// CallbackID is a predicate that is considered to be "true" if and only if the callback ID of the InteractionCallback equals to the given one.
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

// Channel is a predicate that is considered to be "true" if and only if the InteractionCallback is triggered in the given channel.
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

// Build decorates `h` with the given Predicates and returns a new Handler that calls the original handler `h` if and only if all the given Predicates are considered to be "true".
func Build(h Handler, preds ...Predicate) Handler {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	return h
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

// Router is an http.Handler that processes interaction callbacks from Slack.
//
// For more details, see https://api.slack.com/interactivity/handling.
type Router struct {
	signingSecret    string
	skipVerification bool
	handlers         map[slack.InteractionType][]Handler
	fallbackHandler  Handler
	verboseResponse  bool
	httpHandler      http.Handler
}

// New creates a new Router.
//
// At least one of WithSigningSecret() or InsecureSkipVerification() must be specified.
func New(opts ...Option) (*Router, error) {
	r := &Router{
		handlers: make(map[slack.InteractionType][]Handler),
	}
	for _, o := range opts {
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
			Secret:          r.signingSecret,
			VerboseResponse: r.verboseResponse,
			Handler:         r.httpHandler,
		}
	}
	return r, nil
}

// On registers a handler for a specific event type.
//
// Unlike `eventrouter.Router`, the Router does not have type-specific `OnXXX` methods because all types of
// interactions share the same struct in `slack-go/slack`.
//
// If more than one handlers are registered, the first ones take precedence.
//
// Handlers may return `routererrors.NotInterested` (or its equivalents in the sense of `errors.Is`). In such case the Router falls back to other handlers.
//
// Handlers also may return `routererrors.HttpError` (or its equivalents in the sense of `errors.Is`). In such case the Router responds with corresponding HTTP status codes.
//
// If any other errors are returned, the Router responds with Internal Server Error.
func (r *Router) On(typeName slack.InteractionType, h Handler, preds ...Predicate) {
	h = Build(h, preds...)
	handlers, ok := r.handlers[typeName]
	if !ok {
		handlers = make([]Handler, 0)
	}
	handlers = append(handlers, h)
	r.handlers[typeName] = handlers
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

// FindBlockAction finds a block action whose blockID and actionID equal to the given ones.
// If no such block action is found, it returns nil.
func FindBlockAction(callback *slack.InteractionCallback, blockID, actionID string) *slack.BlockAction {
	for _, ba := range callback.ActionCallback.BlockActions {
		if ba.BlockID == blockID && ba.ActionID == actionID {
			return ba
		}
	}
	return nil
}
