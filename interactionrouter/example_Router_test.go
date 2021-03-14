package interactionrouter_test

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
