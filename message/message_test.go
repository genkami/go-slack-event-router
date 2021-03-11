package message_test

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/message"
)

var _ = Describe("Message", func() {
	var (
		numHandlerCalled int
		innerHandler     = message.HandlerFunc(func(ev *slackevents.MessageEvent) error {
			numHandlerCalled++
			return nil
		})
	)
	BeforeEach(func() {
		numHandlerCalled = 0
	})

	Describe("TextRegexp", func() {
		Context("When the text matches to the pattern", func() {
			It("calls the inner handler", func() {
				h := message.TextRegexp(regexp.MustCompile(`\bapple\b`)).Wrap(innerHandler)
				e := &slackevents.MessageEvent{
					Text: "I ate an apple",
				}
				err := h.HandleMessageEvent(e)
				Expect(err).ToNot(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("When the text does not match to the pattern", func() {
			It("does not call the inner handler", func() {
				h := message.TextRegexp(regexp.MustCompile(`\bapple\b`)).Wrap(innerHandler)
				e := &slackevents.MessageEvent{
					Text: "I ate a banana",
				}
				err := h.HandleMessageEvent(e)
				Expect(err).To(Equal(errors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
