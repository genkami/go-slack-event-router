package urlverification

import (
	"github.com/slack-go/slack/slackevents"
)

type Handler interface {
	HandleURLVerification(*slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)
}

type HandlerFunc func(*slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error)

func (f HandlerFunc) HandleURLVerification(e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return f(e)
}

var DefaultHandler Handler = HandlerFunc(func(e *slackevents.EventsAPIURLVerificationEvent) (*slackevents.ChallengeResponse, error) {
	return &slackevents.ChallengeResponse{
		Challenge: e.Challenge,
	}, nil
})
