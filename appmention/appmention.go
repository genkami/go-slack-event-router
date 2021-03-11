package appmention

import (
	"regexp"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
)

type Handler interface {
	HandleAppMentionEvent(*slackevents.AppMentionEvent) error
}

type HandlerFunc func(*slackevents.AppMentionEvent) error

func (f HandlerFunc) HandleAppMentionEvent(e *slackevents.AppMentionEvent) error {
	return f(e)
}

type Predicate interface {
	Wrap(Handler) Handler
}

type inChannelPredicate struct {
	channel string
}

func InChannel(channel string) Predicate {
	return &inChannelPredicate{channel: channel}
}

func (p *inChannelPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(e *slackevents.AppMentionEvent) error {
		if e.Channel != p.channel {
			return errors.NotInterested
		}
		return h.HandleAppMentionEvent(e)
	})
}

type nameRegexpPredicate struct {
	re *regexp.Regexp
}

func NameRegexp(re *regexp.Regexp) Predicate {
	return &nameRegexpPredicate{re: re}
}

func (p *nameRegexpPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(e *slackevents.AppMentionEvent) error {
		idx := p.re.FindStringIndex(e.Text)
		if len(idx) == 0 {
			return errors.NotInterested
		}
		return h.HandleAppMentionEvent(e)
	})
}
