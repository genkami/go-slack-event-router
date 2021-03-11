package urlverification

import (
	"github.com/slack-go/slack/slackevents"
)

type Handler interface {
	HandleURLVerification(*slackevents.ChallengeResponse) (string, error)
}

type HandlerFunc func(*slackevents.ChallengeResponse) (string, error)

func (f HandlerFunc) HandleURLVerification(challenge *slackevents.ChallengeResponse) (string, error) {
	return f(challenge)
}

var DefaultHandler Handler = HandlerFunc(func(challenge *slackevents.ChallengeResponse) (string, error) {
	return challenge.Challenge, nil
})
