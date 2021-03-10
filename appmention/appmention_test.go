package appmention_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/errors"
)

var _ = Describe("AppMention", func() {
	var (
		numHandlerCalled int
		innerHandler     = appmention.HandlerFunc(func(ev *slackevents.AppMentionEvent) error {
			numHandlerCalled++
			return nil
		})
	)
	BeforeEach(func() {
		numHandlerCalled = 0
	})

	Describe("InChannel", func() {
		Context("When the event's channels is the same as the predicate's", func() {
			It("calls the inner handler", func() {
				h := appmention.InChannel("XXX").Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Channel: "XXX",
				}
				err := h.HandleAppMentionEvent(e)
				Expect(err).ToNot(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("When the event's channels is different from the predicate's", func() {
			It("does not call the inner handler", func() {
				h := appmention.InChannel("XXX").Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Channel: "YYY",
				}
				err := h.HandleAppMentionEvent(e)
				Expect(err).To(Equal(errors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
