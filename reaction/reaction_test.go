package reaction_test

import (
	"context"
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
		innerAddedHandler = reaction.AddedHandlerFunc(func(_ context.Context, _ *slackevents.ReactionAddedEvent) error {
			numHandlerCalled++
			return nil
		})
		innerRemovedHandler = reaction.RemovedHandlerFunc(func(_ context.Context, _ *slackevents.ReactionRemovedEvent) error {
			numHandlerCalled++
			return nil
		})
		ctx context.Context
	)
	BeforeEach(func() {
		numHandlerCalled = 0
		ctx = context.Background()
	})

	Describe("BuildAdded", func() {
		Context("when no predicate is given", func() {
			It("returns the original handler", func() {
				h := reaction.BuildAdded(innerAddedHandler)
				e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
				err := h.HandleReactionAddedEvent(ctx, e)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when a single predicate is given", func() {
			Context("when the predicate matches to the given message", func() {
				It("calls the inner handler", func() {
					h := reaction.BuildAdded(innerAddedHandler, reaction.Name("smile"))
					e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("when the predicate does not match to the given message", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildAdded(innerAddedHandler, reaction.Name("sob"))
					e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Context("when more than one predicates are given", func() {
			Context("when none of the predicates matches to the given message", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildAdded(innerAddedHandler,
						reaction.Name("sob"),
						reaction.Name("cry"),
					)
					e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when some of the predicates matche to the given message but others don't", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildAdded(innerAddedHandler,
						reaction.Name("smile"),
						reaction.Name("sob"),
					)
					e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when all of the predicates matche to the given message", func() {
				It("calls the inner handler", func() {
					h := reaction.BuildAdded(innerAddedHandler,
						reaction.Name("smile"),
						reaction.Name("smile"),
					)
					e := &slackevents.ReactionAddedEvent{Reaction: "smile"}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})
		})
	})

	Describe("BuildRemoved", func() {
		Context("when no predicate is given", func() {
			It("returns the original handler", func() {
				h := reaction.BuildRemoved(innerRemovedHandler)
				e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
				err := h.HandleReactionRemovedEvent(ctx, e)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when a single predicate is given", func() {
			Context("when the predicate matches to the given message", func() {
				It("calls the inner handler", func() {
					h := reaction.BuildRemoved(innerRemovedHandler, reaction.Name("smile"))
					e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("when the predicate does not match to the given message", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildRemoved(innerRemovedHandler, reaction.Name("sob"))
					e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Context("when more than one predicates are given", func() {
			Context("when none of the predicates matches to the given message", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildRemoved(innerRemovedHandler,
						reaction.Name("sob"),
						reaction.Name("cry"),
					)
					e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when some of the predicates matche to the given message but others don't", func() {
				It("does not call the inner handler", func() {
					h := reaction.BuildRemoved(innerRemovedHandler,
						reaction.Name("smile"),
						reaction.Name("sob"),
					)
					e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when all of the predicates matche to the given message", func() {
				It("calls the inner handler", func() {
					h := reaction.BuildRemoved(innerRemovedHandler,
						reaction.Name("smile"),
						reaction.Name("smile"),
					)
					e := &slackevents.ReactionRemovedEvent{Reaction: "smile"}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})
		})
	})

	Describe("Name", func() {
		Describe("WrapAdded", func() {
			Context("When the reaction's name is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.Name("smile").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
					}
					err := h.HandleReactionAddedEvent(ctx, e)
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
					err := h.HandleReactionAddedEvent(ctx, e)
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
					err := h.HandleReactionRemovedEvent(ctx, e)
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
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})

	Describe("Channel", func() {
		Describe("WrapAdded", func() {
			Context("When the reaction's channel is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.Channel("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "XXX",
						},
					}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's channel is different from thepredicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.Channel("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "YYY",
						},
					}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Describe("WrapRemoved", func() {
			Context("When the reaction's channel is the same as the predicate's", func() {
				It("calls the inner handler", func() {
					h := reaction.Channel("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "XXX",
						},
					}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the reaction's channel is different from thepredicate's", func() {
				It("does not call the inner handler", func() {
					h := reaction.Channel("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						Item: slackevents.Item{
							Channel: "YYY",
						},
					}
					err := h.HandleReactionRemovedEvent(ctx, e)
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
					err := h.HandleReactionAddedEvent(ctx, e)
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
					err := h.HandleReactionAddedEvent(ctx, e)
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
					err := h.HandleReactionAddedEvent(ctx, e)
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
					err := h.HandleReactionRemovedEvent(ctx, e)
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
					err := h.HandleReactionRemovedEvent(ctx, e)
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
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})

	Describe("ItemUser", func() {
		Describe("WrapAdded", func() {
			Context("When the author of the reacted item is the given one", func() {
				It("calls the inner handler", func() {
					h := reaction.ItemUser("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						ItemUser: "XXX",
					}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the author of the reacted message is different from the given one", func() {
				It("does not call the inner handler", func() {
					h := reaction.ItemUser("XXX").WrapAdded(innerAddedHandler)
					e := &slackevents.ReactionAddedEvent{
						Reaction: "smile",
						ItemUser: "YYY",
					}
					err := h.HandleReactionAddedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Describe("WrapRemoved", func() {
			Context("When the author of the reacted message is the given one", func() {
				It("calls the inner handler", func() {
					h := reaction.ItemUser("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						ItemUser: "XXX",
					}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).ToNot(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("When the author of the reacted message is different from the given one", func() {
				It("does not call the inner handler", func() {
					h := reaction.ItemUser("XXX").WrapRemoved(innerRemovedHandler)
					e := &slackevents.ReactionRemovedEvent{
						Reaction: "smile",
						ItemUser: "YYY",
					}
					err := h.HandleReactionRemovedEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})
	})
})
