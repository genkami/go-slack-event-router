package interactionrouter_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack"

	routererrors "github.com/genkami/go-slack-event-router/errors"
	ir "github.com/genkami/go-slack-event-router/interactionrouter"
)

var _ = Describe("InteractionRouter", func() {
	Describe("Type", func() {
		var (
			numHandlerCalled int
			innerHandler     = ir.HandlerFunc(func(_ *slack.InteractionCallback) error {
				numHandlerCalled++
				return nil
			})
		)
		BeforeEach(func() {
			numHandlerCalled = 0
		})

		Context("when the type of the interaction callback matches to the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.Type(slack.InteractionTypeBlockActions).Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the type of the interaction callback differs from the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.Type(slack.InteractionTypeBlockActions).Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeViewSubmission,
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
