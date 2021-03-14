// Package appmention provides handlers to process `app_mention` events.
//
// For more details, see https://api.slack.com/events/app_mention.
package appmention

import (
	"context"
	"regexp"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
)

// Handler processes `app_mention` events.
type Handler interface {
	HandleAppMentionEvent(context.Context, *slackevents.AppMentionEvent) error
}

type HandlerFunc func(context.Context, *slackevents.AppMentionEvent) error

func (f HandlerFunc) HandleAppMentionEvent(ctx context.Context, e *slackevents.AppMentionEvent) error {
	return f(ctx, e)
}

// Predicate disthinguishes whether or not a certain handler should process coming events.
type Predicate interface {
	Wrap(Handler) Handler
}

type inChannelPredicate struct {
	channel string
}

// InChannel is a predicate that is considered to be "true" if and only if an event happened in the given channel.
func InChannel(channel string) Predicate {
	return &inChannelPredicate{channel: channel}
}

func (p *inChannelPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, e *slackevents.AppMentionEvent) error {
		if e.Channel != p.channel {
			return errors.NotInterested
		}
		return h.HandleAppMentionEvent(ctx, e)
	})
}

type textRegexpPredicate struct {
	re *regexp.Regexp
}

// TextRegexp is a predicate that is considered to be "true" if and only if a text of a message matches to the given regexp.
func TextRegexp(re *regexp.Regexp) Predicate {
	return &textRegexpPredicate{re: re}
}

func (p *textRegexpPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(ctx context.Context, e *slackevents.AppMentionEvent) error {
		idx := p.re.FindStringIndex(e.Text)
		if len(idx) == 0 {
			return errors.NotInterested
		}
		return h.HandleAppMentionEvent(ctx, e)
	})
}

// Build decorates `h` with the given Predicates and returns a new Handler that calls the original handler `h` if and only if all the given Predicates are considered to be "true".
func Build(h Handler, preds ...Predicate) Handler {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	return h
}
