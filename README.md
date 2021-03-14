# go-slack-event-router

![ci status](https://github.com/genkami/go-slack-event-router/workflows/Test/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/genkami/go-slack-event-router.svg)](https://pkg.go.dev/github.com/genkami/go-slack-event-router)

This is not a chatbot framework, but rather a simple event dispatcher that enhances the functionality of [slack-go/slack](https://github.com/slack-go/slack).

## Example

To handle Events API:

```go
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
	r.OnReactionAdded(reaction.AddedHandlerFunc(handleIssue), reaction.Name("issue"), reaction.Channel("ABCXYZ"))

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
```

To handle user interaction:

```go
import (
	"context"
	"net/http"
	"os"

	"github.com/genkami/go-slack-event-router/interactionrouter"
	"github.com/slack-go/slack"
)

func ExampleRouter() {
	signingSecret := os.Getenv("SLACK_SIGNING_SECRET")
	r, _ := interactionrouter.New(interactionrouter.WithSigningSecret(signingSecret)) // omitted error handling

	// Call handlePostNiceGif whenever the router receives `block_actions` event with a block `post_nice_gif` with an action `gif_keyword`.
	r.On(slack.InteractionTypeBlockActions, interactionrouter.HandlerFunc(handlePostNiceGif),
		interactionrouter.BlockAction("post_nice_gif", "gif_keyword"))

	http.Handle("/slack/actions", r)

	// ...
}

func handlePostNiceGif(ctx context.Context, callback *slack.InteractionCallback) error {
	// Do whatever you want...
	return nil
}
```

## License

Distributed under the Apache License Version 2.0. See LICENSE for more information.
