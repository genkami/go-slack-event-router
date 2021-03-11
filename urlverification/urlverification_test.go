package urlverification_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/urlverification"
)

var _ = Describe("URLVerification", func() {
	Describe("DefaultHandler", func() {
		It("returns the given challenge", func() {
			challenge := &slackevents.ChallengeResponse{Challenge: "hello"}
			resp, err := urlverification.DefaultHandler.HandleURLVerification(challenge)
			Expect(err).ToNot(HaveOccurred())
			Expect(resp).To(Equal(challenge.Challenge))
		})
	})
})
