package message_test

import (
	"context"
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
		innerHandler     = message.HandlerFunc(func(_ context.Context, _ *slackevents.MessageEvent) error {
			numHandlerCalled++
			return nil
		})
		ctx context.Context
	)
	BeforeEach(func() {
		numHandlerCalled = 0
		ctx = context.Background()
	})

	Describe("Build", func() {
		Context("when no predicate is given", func() {
			It("returns the original handler", func() {
				h := message.Build(innerHandler)
				e := &slackevents.MessageEvent{Text: "hello world"}
				err := h.HandleMessageEvent(ctx, e)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when a single predicate is given", func() {
			Context("when the predicate matches to the given message", func() {
				It("calls the inner handler", func() {
					h := message.Build(innerHandler, message.TextRegexp(regexp.MustCompile(`hello`)))
					e := &slackevents.MessageEvent{Text: "hello world"}
					err := h.HandleMessageEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("when the predicate does not match to the given message", func() {
				It("does not call the inner handler", func() {
					h := message.Build(innerHandler, message.TextRegexp(regexp.MustCompile(`BYE`)))
					e := &slackevents.MessageEvent{Text: "hello world"}
					err := h.HandleMessageEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Context("when more than one predicates are given", func() {
			Context("when none of the predicates matches to the given message", func() {
				It("does not call the inner handler", func() {
					h := message.Build(innerHandler,
						message.TextRegexp(regexp.MustCompile(`BYE`)),
						message.TextRegexp(regexp.MustCompile(`GOOD NIGHT`)),
					)
					e := &slackevents.MessageEvent{Text: "hello world"}
					err := h.HandleMessageEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when some of the predicates matche to the given message but others don't", func() {
				It("does not call the inner handler", func() {
					h := message.Build(innerHandler,
						message.TextRegexp(regexp.MustCompile(`hello`)),
						message.TextRegexp(regexp.MustCompile(`GOOD NIGHT`)),
					)
					e := &slackevents.MessageEvent{Text: "hello world"}
					err := h.HandleMessageEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when all of the predicates matche to the given message", func() {
				It("calls the inner handler", func() {
					h := message.Build(innerHandler,
						message.TextRegexp(regexp.MustCompile(`hello`)),
						message.TextRegexp(regexp.MustCompile(`world`)),
					)
					e := &slackevents.MessageEvent{Text: "hello world"}
					err := h.HandleMessageEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})
		})
	})

	Describe("TextRegexp", func() {
		Context("When the text matches to the pattern", func() {
			It("calls the inner handler", func() {
				h := message.TextRegexp(regexp.MustCompile(`\bapple\b`)).Wrap(innerHandler)
				e := &slackevents.MessageEvent{
					Text: "I ate an apple",
				}
				err := h.HandleMessageEvent(ctx, e)
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
				err := h.HandleMessageEvent(ctx, e)
				Expect(err).To(Equal(errors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
