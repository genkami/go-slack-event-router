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

	Describe("BlockAction", func() {
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

		Context("when the interaction callback has the block_action specified by the predicate", func() {
			It("calls the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when one of the block_acsions that the interaction callback has is the one specified by the predicate", func() {
			It("calls the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
							{BlockID: "BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the interaction callback does not have any block_action", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the block_action in the interaction callback is not what the predicate expects", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the block_id in the block_action is the same as the predicate expected but the action_id isn't", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "BLOCK_ID", ActionID: "ANOTHER_ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})

		Context("when the action_id in the block_action is the same as the predicate expected but the block_id isn't", func() {
			It("does not call the inner handler", func() {
				h := ir.BlockAction("BLOCK_ID", "ACTION_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type: slack.InteractionTypeBlockActions,
					ActionCallback: slack.ActionCallbacks{
						BlockActions: []*slack.BlockAction{
							{BlockID: "ANOTHER_BLOCK_ID", ActionID: "ACTION_ID"},
						},
					},
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})

	Describe("CallbackID", func() {
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

		Context("when the callback_id in the interaction callback matches to the predicate's", func() {
			It("calls the inner handler", func() {
				h := ir.CallbackID("CALLBACK_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type:       slack.InteractionTypeBlockActions,
					CallbackID: "CALLBACK_ID",
				}
				err := h.HandleInteraction(callback)
				Expect(err).NotTo(HaveOccurred())
				Expect(numHandlerCalled).To(Equal(1))
			})
		})

		Context("when the callback_id in the interaction callback differs from the predicate's", func() {
			It("does not call the inner handler", func() {
				h := ir.CallbackID("CALLBACK_ID").Wrap(innerHandler)
				callback := &slack.InteractionCallback{
					Type:       slack.InteractionTypeBlockActions,
					CallbackID: "ANOTHER_CALLBACK_ID",
				}
				err := h.HandleInteraction(callback)
				Expect(err).To(Equal(routererrors.NotInterested))
				Expect(numHandlerCalled).To(Equal(0))
			})
		})
	})
})
