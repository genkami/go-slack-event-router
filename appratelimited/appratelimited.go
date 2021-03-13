// package appratelimited provides handler to process `app_rate_limited` events.
//
// For more details, see https://api.slack.com/events/app_rate_limited.
package appratelimited

import "github.com/slack-go/slack/slackevents"

// Handler processes `app_rate_limited` events.
type Handler interface {
	HandleAppRateLimited(*slackevents.EventsAPIAppRateLimited) error
}

type HandlerFunc func(*slackevents.EventsAPIAppRateLimited) error

func (f HandlerFunc) HandleAppRateLimited(e *slackevents.EventsAPIAppRateLimited) error {
	return f(e)
}

// DefaultHandler is the default handler that the Router uses.
// It just ignores `app_rate_limited` events.
var DefaultHandler Handler = HandlerFunc(func(e *slackevents.EventsAPIAppRateLimited) error {
	return nil
})
