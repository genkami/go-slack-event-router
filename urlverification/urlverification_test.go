package urlverification_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/urlverification"
)

var _ = Describe("URLVerification", func() {
	Describe("DefaultHandler", func() {
		It("returns the given challenge", func() {
			ctx := context.Background()
			e := &slackevents.EventsAPIURLVerificationEvent{Challenge: "hello"}
			resp, err := urlverification.DefaultHandler.HandleURLVerification(ctx, e)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(&slackevents.ChallengeResponse{Challenge: "hello"}))
		})
	})
})
