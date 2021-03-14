// package appratelimited provides handler to process `app_rate_limited` events.
//
// For more details, see https://api.slack.com/events/app_rate_limited.
package appratelimited

import (
	"context"

	"github.com/slack-go/slack/slackevents"
)

// Handler processes `app_rate_limited` events.
type Handler interface {
	HandleAppRateLimited(context.Context, *slackevents.EventsAPIAppRateLimited) error
}

type HandlerFunc func(context.Context, *slackevents.EventsAPIAppRateLimited) error

func (f HandlerFunc) HandleAppRateLimited(ctx context.Context, e *slackevents.EventsAPIAppRateLimited) error {
	return f(ctx, e)
}

// DefaultHandler is the default handler that the Router uses.
// It just ignores `app_rate_limited` events.
var DefaultHandler Handler = HandlerFunc(func(_ context.Context, _ *slackevents.EventsAPIAppRateLimited) error {
	return nil
})
