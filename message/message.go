package message

import (
	"regexp"

	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
)

type Handler interface {
	HandleMessageEvent(*slackevents.MessageEvent) error
}

type HandlerFunc func(*slackevents.MessageEvent) error

func (f HandlerFunc) HandleMessageEvent(e *slackevents.MessageEvent) error {
	return f(e)
}

type Predicate interface {
	Wrap(h Handler) Handler
}

type textRegexpPredicate struct {
	re *regexp.Regexp
}

func TextRegexp(re *regexp.Regexp) Predicate {
	return &textRegexpPredicate{re: re}
}

func (p *textRegexpPredicate) Wrap(h Handler) Handler {
	return HandlerFunc(func(e *slackevents.MessageEvent) error {
		idx := p.re.FindStringIndex(e.Text)
		if len(idx) == 0 {
			return errors.NotInterested
		}
		return h.HandleMessageEvent(e)
	})
}

func Build(h Handler, preds ...Predicate) Handler {
	for _, p := range preds {
		h = p.Wrap(h)
	}
	return h
}
