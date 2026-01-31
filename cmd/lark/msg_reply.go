package main

import (
	"context"
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
				return err
			}
			if len(args) > 0 {
				if messageID != "" && messageID != args[0] {
					return errors.New("message-id provided twice")
				}
				if err := cmd.Flags().Set("message-id", args[0]); err != nil {
					return err
				}
				return nil
			}
			if strings.TrimSpace(messageID) == "" {
				return errors.New("message-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			msgType, content, err := resolveMessageContent(contentOpts)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			replyID, err := state.SDK.ReplyMessage(context.Background(), token, larksdk.ReplyMessageRequest{
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

	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID to reply to (or provide as positional argument)")
	cmd.Flags().BoolVar(&replyInThread, "reply-in-thread", false, "reply in thread")
	addMessageContentFlags(cmd, &contentOpts)
	return cmd
}
