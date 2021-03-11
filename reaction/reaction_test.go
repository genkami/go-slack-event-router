package reaction_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/reaction"
)

var _ = Describe("AppMention", func() {
	var (
		numHandlerCalled  int
		innerAddedHandler = reaction.AddedHandlerFunc(func(_ *slackevents.ReactionAddedEvent) error {
			numHandlerCalled++
			return nil
		})
		innerRemovedHandler = reaction.RemovedHandlerFunc(func(ev *slackevents.ReactionRemovedEvent) error {
			numHandlerCalled++
			return nil
		})
	)
	BeforeEach(func() {
		numHandlerCalled = 0
	})

	Describe("Name", func() {
		Describe("WrapAdded", func() {
			Context("When the reaction's name is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.Name("smile").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's name is different from the predicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.Name("smile").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "sob",
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Describe("WrapRemoved", func() {
			Context("When the reaction's name is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.Name("smile").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's name is different from the predicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.Name("smile").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "sob",
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})
})
