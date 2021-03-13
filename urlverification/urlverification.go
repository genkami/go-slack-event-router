// Package urlverification provides handlers to process `url_verification` events.
//
// For more details, see https://api.slack.com/events/url_verification.
package urlverification

import (
	"github.com/slack-go/slack/slackevents"
)

// Handler processes `url_verification` events.
type Handler interface {
	HandleURLVerification(*slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)
}

type HandlerFunc func(*slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)

func (f HandlerFunc) HandleURLVerification(e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return f(e)
}

// DefaultHandler just echoes back the given challenge value.
var DefaultHandler Handler = HandlerFunc(func(e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return &slackevents.ChallengeResponse{
		Challenge: e.Challenge,
	}, nil
})
