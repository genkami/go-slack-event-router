// Package urlverification provides handlers to process `url_verification` events.
//
// For more details, see https://api.slack.com/events/url_verification.
package urlverification

import (
	"context"

	"github.com/slack-go/slack/slackevents"
)

// Handler processes `url_verification` events.
type Handler interface {
	HandleURLVerification(context.Context, *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)
}

type HandlerFunc func(context.Context, *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)

func (f HandlerFunc) HandleURLVerification(ctx context.Context, e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return f(ctx, e)
}

// DefaultHandler just echoes back the given challenge value.
var DefaultHandler Handler = HandlerFunc(func(_ context.Context, e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return &slackevents.ChallengeResponse{
		Challenge: e.Challenge,
	}, nil
})
