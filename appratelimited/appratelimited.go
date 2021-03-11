package appratelimited

import "github.com/slack-go/slack/slackevents"

type Handler interface {
	HandleAppRateLimited(*slackevents.EventsAPIAppRateLimited) error
}

type HandlerFunc func(*slackevents.EventsAPIAppRateLimited) error

func (f HandlerFunc) HandleAppRateLimited(e *slackevents.EventsAPIAppRateLimited) error {
	return f(e)
}

var DefaultHandler Handler = HandlerFunc(func(e *slackevents.EventsAPIAppRateLimited) error {
	return nil
})
