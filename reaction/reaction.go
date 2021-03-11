package reaction

import (
	"github.com/genkami/go-slack-event-router/errors"
	"github.com/slack-go/slack/slackevents"
)

type AddedHandler interface {
	HandleReactionAddedEvent(*slackevents.ReactionAddedEvent) error
}

type AddedHandlerFunc func(*slackevents.ReactionAddedEvent) error

func (f AddedHandlerFunc) HandleReactionAddedEvent(e *slackevents.ReactionAddedEvent) error {
	return f(e)
}

type RemovedHandler interface {
	HandleReactionRemovedEvent(*slackevents.ReactionRemovedEvent) error
}

type RemovedHandlerFunc func(*slackevents.ReactionRemovedEvent) error

func (f RemovedHandlerFunc) HandleReactionRemovedEvent(e *slackevents.ReactionRemovedEvent) error {
	return f(e)
}

type Predicate interface {
	WrapAdded(AddedHandler) AddedHandler
	WrapRemoved(RemovedHandler) RemovedHandler
}

type namePredicate struct {
	reaction string
}

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
