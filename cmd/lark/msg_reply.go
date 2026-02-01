package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMsgReplyCmd(state *appState) *cobra.Command {
	var messageID string
	var replyInThread bool
	var contentOpts messageContentOptions

	cmd := &cobra.Command{
		Use:   "reply <message-id>",
		Short: "Reply to a message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) == 0 {
				return errors.New("message-id is required")
			}
			messageID = strings.TrimSpace(args[0])
			if messageID == "" {
				return errors.New("message-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			msgType, content, err := resolveMessageContent(contentOpts)
			if err != nil {
				return err
			}
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			replyID, err := state.SDK.ReplyMessage(cmd.Context(), token, larksdk.ReplyMessageRequest{
				MessageID:     messageID,
				MsgType:       msgType,
				Content:       content,
				Text:          strings.TrimSpace(contentOpts.Text),
				ReplyInThread: replyInThread,
				UUID:          strings.TrimSpace(contentOpts.UUID),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"message_id": replyID}
			return state.Printer.Print(payload, fmt.Sprintf("message_id: %s", replyID))
		},
	}

	cmd.Flags().BoolVar(&replyInThread, "reply-in-thread", false, "reply in thread")
	addMessageContentFlags(cmd, &contentOpts)
	return cmd
}
