package reaction_test

import (
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/errors"
	"github.com/genkami/go-slack-event-router/reaction"
)

var _ = Describe("Reaction", func() {
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

	Describe("InChannel", func() {
		Describe("WrapAdded", func() {
			Context("When the reaction's channel is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.InChannel("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "XXX",
						},
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's channel is different from thepredicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.InChannel("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "YYY",
						},
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Describe("WrapRemoved", func() {
			Context("When the reaction's channel is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.InChannel("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "XXX",
						},
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's channel is different from thepredicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.InChannel("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "YYY",
						},
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})

	Describe("MessageTextRegexp", func() {
		Describe("WrapAdded", func() {
			Context("When the text of the reacted message matches to the pattern", func() {
				It("calls the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Message: &slackevents.ItemMessage{
								Text: "I ate an apple",
							},
						},
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the text of the reacted message does not match to the pattern", func() {
				It("does not call the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Message: &slackevents.ItemMessage{
								Text: "I ate a banana",
							},
						},
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("When the message is nil", func() {
				It("does not call the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
					}
					err := h.HandleReactionAddedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Describe("WrapRemoved", func() {
			Context("When the text of the reacted message matches to the pattern", func() {
				It("calls the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Message: &slackevents.ItemMessage{
								Text: "I ate an apple",
							},
						},
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the text of the reacted message does not match to the pattern", func() {
				It("does not call the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Message: &slackevents.ItemMessage{
								Text: "I ate a banana",
							},
						},
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("When the message is nil", func() {
				It("does not call the inner handler", func() {
					h := reaction.MessageTextRegexp(regexp.MustCompile(`apple`)).WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
					}
					err := h.HandleReactionRemovedEvent(e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})
})
