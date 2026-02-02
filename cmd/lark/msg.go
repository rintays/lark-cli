package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMsgCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"msg"},
		Short:   "Send chat messages",
		Long: `Messages are chat messages sent in conversations.

- Chats have chat_id; messages have message_id.
- Send uses receive_id + receive_id_type to target a chat or user.
- Reply/reactions/pin operate on an existing message.`,
	}
	annotateAuthServices(cmd, "im")
	cmd.AddCommand(newMsgSendCmd(state))
	cmd.AddCommand(newMsgReplyCmd(state))
	cmd.AddCommand(newMsgListCmd(state))
	cmd.AddCommand(newMsgSearchCmd(state))
	cmd.AddCommand(newMsgReactionsCmd(state))
	cmd.AddCommand(newMsgPinCmd(state))
	cmd.AddCommand(newMsgUnpinCmd(state))
	return cmd
}

func newMsgSendCmd(state *appState) *cobra.Command {
	var receiveID string
	var receiveIDType string
	var contentOpts messageContentOptions

	cmd := &cobra.Command{
		Use:   "send <receive-id>",
		Short: "Send a message to a chat or user",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			receiveID = strings.TrimSpace(args[0])
			if receiveID == "" {
				return argsUsageError(cmd, errors.New("receive-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			normalizedType, ok := normalizeReceiveIDType(receiveIDType)
			if !ok {
				return flagUsage(cmd, "receive-id-type must be one of chat_id, open_id, user_id, email")
			}
			receiveIDType = normalizedType
			msgType, content, err := resolveMessageContent(contentOpts)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			messageID, err := state.SDK.SendMessage(cmd.Context(), token, larksdk.MessageRequest{
				ReceiveID:     receiveID,
				ReceiveIDType: receiveIDType,
				MsgType:       msgType,
				Content:       content,
				Text:          strings.TrimSpace(contentOpts.Text),
				UUID:          strings.TrimSpace(contentOpts.UUID),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"message_id": messageID}
			return state.Printer.Print(payload, fmt.Sprintf("message_id: %s", messageID))
		},
	}

	cmd.Flags().StringVar(&receiveIDType, "receive-id-type", "chat_id", "receive ID type (chat_id, open_id, user_id, email)")
	addMessageContentFlags(cmd, &contentOpts)
	registerEnumCompletion(cmd, "receive-id-type", receiveIDTypeValues)
	return cmd
}
