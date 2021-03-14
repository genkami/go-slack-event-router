package eventrouter_test

import (
	"context"
	"net/http"
	"os"
	"regexp"

	eventrouter "github.com/genkami/go-slack-event-router"
	"github.com/genkami/go-slack-event-router/message"
	"github.com/genkami/go-slack-event-router/reaction"
	"github.com/slack-go/slack/slackevents"
)

func ExampleRouter() {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	r, _ := eventrouter.New(eventrouter.WithSigningSecret(signingSecret)) // omitted error handling

	// Call handleDeploy whenever the router receives `message` events and the text of the message matches to /deploy/.
	r.OnMessage(message.HandlerFunc(handleDeploy), message.TextRegexp(regexp.MustCompile(`deploy`)))

	// Call handleIssue whenever the router receives `reaction_added` events with reaction `:issue:` and the event happens in the channel ABCXYZ.
	r.OnReactionAdded(reaction.AddedHandlerFunc(handleIssue), reaction.Name("issue"), reaction.InChannel("ABCXYZ"))

	http.Handle("/slack/events", r)

	// ...
}

func handleDeploy(ctx context.Context, e *slackevents.MessageEvent) error {
	// Do whatever you want...
	return nil
}

func handleIssue(ctx context.Context, e *slackevents.ReactionAddedEvent) error {
	// Do whatever you want...
	return nil
}
