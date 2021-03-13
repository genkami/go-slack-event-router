// Package reaction provides handlers to process `reaction_*` events.
//
// For more details, see the following pages:
//   * https://api.slack.com/events/reaction_added
//   * https://api.slack.com/events/reaction_removed
package reaction

import (
	"regexp"

	"github.com/genkami/go-slack-event-router/errors"
	"github.com/slack-go/slack/slackevents"
)

// AddedHandler processes `reaction_added` events.
type AddedHandler interface {
	HandleReactionAddedEvent(*slackevents.ReactionAddedEvent) error
}

type AddedHandlerFunc func(*slackevents.ReactionAddedEvent) error

func (f AddedHandlerFunc) HandleReactionAddedEvent(e *slackevents.ReactionAddedEvent) error {
	return f(e)
}

// RemovedHandler processes `reaction_removed` events.
type RemovedHandler interface {
	HandleReactionRemovedEvent(*slackevents.ReactionRemovedEvent) error
}

type RemovedHandlerFunc func(*slackevents.ReactionRemovedEvent) error

func (f RemovedHandlerFunc) HandleReactionRemovedEvent(e *slackevents.ReactionRemovedEvent) error {
	return f(e)
}

// Predicate disthinguishes whether or not a certain handler should process coming events.
// This can be used with both `AddedHandler` and `RemovedHandler`.
type Predicate interface {
	WrapAdded(AddedHandler) AddedHandler
	WrapRemoved(RemovedHandler) RemovedHandler
}

type namePredicate struct {
	reaction string
}

// Name is a predicate that is considered to be "true" if and only if a reaction name equals to the given one.
func Name(reaction string) Predicate {
	return &namePredicate{reaction: reaction}
}

func (p *namePredicate) WrapAdded(h AddedHandler) AddedHandler {
	return AddedHandlerFunc(func(e *slackevents.ReactionAddedEvent) error {
		if p.reaction != e.Reaction {
			return errors.NotInterested
		}
		return h.HandleReactionAddedEvent(e)
	})
}

func (p *namePredicate) WrapRemoved(h RemovedHandler) RemovedHandler {
	return RemovedHandlerFunc(func(e *slackevents.ReactionRemovedEvent) error {
		if p.reaction != e.Reaction {
			return errors.NotInterested
		}
		return h.HandleReactionRemovedEvent(e)
	})
}

type inChannelPredicate struct {
	channel string
}

// InChannel is a predicate that is considered to be "true" if and only if an event happened in the given channel.
func InChannel(channel string) Predicate {
	return &inChannelPredicate{channel: channel}
}

func (p *inChannelPredicate) WrapAdded(h AddedHandler) AddedHandler {
	return AddedHandlerFunc(func(e *slackevents.ReactionAddedEvent) error {
		if p.channel != e.Item.Channel {
			return errors.NotInterested
		}
		return h.HandleReactionAddedEvent(e)
	})
}

func (p *inChannelPredicate) WrapRemoved(h RemovedHandler) RemovedHandler {
	return RemovedHandlerFunc(func(e *slackevents.ReactionRemovedEvent) error {
		if p.channel != e.Item.Channel {
			return errors.NotInterested
		}
		return h.HandleReactionRemovedEvent(e)
	})
}

type messageTextRegexpPredicate struct {
	re *regexp.Regexp
}

// MessageTextRegexp is a predicate that is considered to be "true" if and only if a text of a reacted message matches to the given regexp.
func MessageTextRegexp(re *regexp.Regexp) Predicate {
	return &messageTextRegexpPredicate{re: re}
}

func (p *messageTextRegexpPredicate) match(item *slackevents.Item) error {
	if item.Message == nil {
		return errors.NotInterested
	}
	idx := p.re.FindStringIndex(item.Message.Text)
	if len(idx) == 0 {
		return errors.NotInterested
	}
	return nil
}

func (p *messageTextRegexpPredicate) WrapAdded(h AddedHandler) AddedHandler {
	return AddedHandlerFunc(func(e *slackevents.ReactionAddedEvent) error {
		if err := p.match(&e.Item); err != nil {
			return err
		}
		return h.HandleReactionAddedEvent(e)
	})
}

func (p *messageTextRegexpPredicate) WrapRemoved(h RemovedHandler) RemovedHandler {
	return RemovedHandlerFunc(func(e *slackevents.ReactionRemovedEvent) error {
		if err := p.match(&e.Item); err != nil {
			return err
		}
		return h.HandleReactionRemovedEvent(e)
	})
}

// BuildAdded decorates `AddedHandler` `h` with the given Predicates and returns a new Handler that calls the original handler `h` if and only if all the given Predicates are considered to be "true".
func BuildAdded(h AddedHandler, preds ...Predicate) AddedHandler {
	for _, p := range preds {
		h = p.WrapAdded(h)
	}
	return h
}

// BuildRemoved decorates `RemovedHandler` `h` with the given Predicates and returns a new Handler that calls the original handler `h` if and only if all the given Predicates are considered to be "true".
func BuildRemoved(h RemovedHandler, preds ...Predicate) RemovedHandler {
	for _, p := range preds {
		h = p.WrapRemoved(h)
	}
	return h
}
