package appmention_test

import (
	"context"
	"regexp"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/slack-go/slack/slackevents"

	"github.com/genkami/go-slack-event-router/appmention"
	"github.com/genkami/go-slack-event-router/errors"
)

var _ = Describe("AppMention", func() {
	var (
		numHandlerCalled int
		innerHandler     = appmention.HandlerFunc(func(ctx context.Context, ev *slackevents.AppMentionEvent) error {
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
				h := appmention.Build(innerHandler)
				e := &slackevents.AppMentionEvent{Text: "hello world"}
				err := h.HandleAppMentionEvent(ctx, e)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when a single predicate is given", func() {
			Context("when the predicate matches to the given message", func() {
				It("calls the inner handler", func() {
					h := appmention.Build(innerHandler, appmention.TextRegexp(regexp.MustCompile(`hello`)))
					e := &slackevents.AppMentionEvent{Text: "hello world"}
					err := h.HandleAppMentionEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})

			Context("when the predicate does not match to the given message", func() {
				It("does not call the inner handler", func() {
					h := appmention.Build(innerHandler, appmention.TextRegexp(regexp.MustCompile(`BYE`)))
					e := &slackevents.AppMentionEvent{Text: "hello world"}
					err := h.HandleAppMentionEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})
		})

		Context("when more than one predicates are given", func() {
			Context("when none of the predicates matches to the given message", func() {
				It("does not call the inner handler", func() {
					h := appmention.Build(innerHandler,
						appmention.TextRegexp(regexp.MustCompile(`BYE`)),
						appmention.TextRegexp(regexp.MustCompile(`GOOD NIGHT`)),
					)
					e := &slackevents.AppMentionEvent{Text: "hello world"}
					err := h.HandleAppMentionEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when some of the predicates matche to the given message but others don't", func() {
				It("does not call the inner handler", func() {
					h := appmention.Build(innerHandler,
						appmention.TextRegexp(regexp.MustCompile(`hello`)),
						appmention.TextRegexp(regexp.MustCompile(`GOOD NIGHT`)),
					)
					e := &slackevents.AppMentionEvent{Text: "hello world"}
					err := h.HandleAppMentionEvent(ctx, e)
					Expect(err).To(Equal(errors.NotInterested))
					Expect(numHandlerCalled).To(Equal(0))
				})
			})

			Context("when all of the predicates matche to the given message", func() {
				It("calls the inner handler", func() {
					h := appmention.Build(innerHandler,
						appmention.TextRegexp(regexp.MustCompile(`hello`)),
						appmention.TextRegexp(regexp.MustCompile(`world`)),
					)
					e := &slackevents.AppMentionEvent{Text: "hello world"}
					err := h.HandleAppMentionEvent(ctx, e)
					Expect(err).NotTo(HaveOccurred())
					Expect(numHandlerCalled).To(Equal(1))
				})
			})
		})
	})

	Describe("Channel", func() {
		Context("When the event's channels is the same as the predicate's", func() {
			It("calls the inner handler", func() {
				h := appmention.Channel("XXX").Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Channel: "XXX",
				}
				err := h.HandleAppMentionEvent(ctx, e)
				Expect(err).ToNot(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("When the event's channels is different from the predicate's", func() {
			It("does not call the inner handler", func() {
				h := appmention.Channel("XXX").Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Channel: "YYY",
				}
				err := h.HandleAppMentionEvent(ctx, e)
				Expect(err).To(Equal(errors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})

	Describe("TextRegexp", func() {
		Context("When the text matches to the pattern", func() {
			It("calls the inner handler", func() {
				h := appmention.TextRegexp(regexp.MustCompile(`\bapple\b`)).Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Text: "I ate an apple",
				}
				err := h.HandleAppMentionEvent(ctx, e)
				Expect(err).ToNot(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("When the text does not match to the pattern", func() {
			It("does not call the inner handler", func() {
				h := appmention.TextRegexp(regexp.MustCompile(`\bapple\b`)).Wrap(innerHandler)
				e := &slackevents.AppMentionEvent{
					Text: "I ate a banana",
				}
				err := h.HandleAppMentionEvent(ctx, e)
				Expect(err).To(Equal(errors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
