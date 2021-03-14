// Package message provides handlers to process `message` events.
//
// For more details, see https://api.slack.com/events/message.
package message

import (
	"context"
	"regexp"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
)

// Handler processes `message` events.
type Handler interface {
	HandleMessageEvent(context.Context, *slackevents.MessageEvent) error
}

type HandlerFunc func(context.Context, *slackevents.MessageEvent) error

func (f HandlerFunc) HandleMessageEvent(ctx context.Context, e *slackevents.MessageEvent) error {
	return f(ctx, e)
}

// Predicate disthinguishes whether or not a certain handler should process coming events.
type Predicate interface {
	Wrap(h Handler) Handler
}

type textRegexpPredicate struct {
	re *regexp.Regexp
}

// TextRegexp is a predicate that is considered to be "true" if and only if a text of a message matches to the given regexp.
func TextRegexp(re *regexp.Regexp) Predicate {
	return &textRegexpPredicate{re: re}
}

func (p *textRegexpPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, e *slackevents.MessageEvent) error {
		idx := p.re.FindStringIndex(e.Text)
		if len(idx) == 0 {
			return errors.NotInterested
		}
		return h.HandleMessageEvent(ctx, e)
	})
}

type channelPredicate struct {
	id string
}

// Channel is a predicate that is considered to be "true" if and only if a message is posted to the given channel.
func Channel(id string) Predicate {
	return &channelPredicate{id: id}
}

func (p *channelPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, e *slackevents.MessageEvent) error {
		if e.Channel != p.id {
			return errors.NotInterested
		}
		return h.HandleMessageEvent(ctx, e)
	})
}

type subTypePredicate struct {
	subType string
}

// SubType is a predicate that is considered to be "true" is and only if a subtype of a message is the given one.
//
// The full list of all subtypes is described here: https://api.slack.com/events/message
func SubType(subType string) Predicate {
	return &subTypePredicate{subType: subType}
}

func (p *subTypePredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, e *slackevents.MessageEvent) error {
		if e.SubType != p.subType {
			return errors.NotInterested
		}
		return h.HandleMessageEvent(ctx, e)
	})
}

// Build decorates `h` with the given Predicates and returns a new Handler that calls the original handler `h` if and only if all the given Predicates are considered to be "true".
func Build(h Handler, preds ...Predicate) Handler {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	return h
}
