package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newMsgCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "messages",
		Aliases: []string{"msg"},
		Short:   "Send chat messages",
	}
	cmd.AddCommand(newMsgSendCmd(state))
	return cmd
}

func newMsgSendCmd(state *appState) *cobra.Command {
	var chatID string
	var receiveID string
	var receiveIDType string
	var text string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send a text message to a chat",
		RunE: func(cmd *cobra.Command, args []string) error {
			if receiveID == "" {
				receiveID = chatID
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			messageID, err := state.SDK.SendMessage(context.Background(), token, larksdk.MessageRequest{
				ReceiveID:     receiveID,
				ReceiveIDType: receiveIDType,
				Text:          text,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"message_id": messageID}
			return state.Printer.Print(payload, fmt.Sprintf("message_id: %s", messageID))
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "chat ID to receive message")
	cmd.Flags().StringVar(&receiveID, "receive-id", "", "receive ID to receive message")
	cmd.Flags().StringVar(&receiveIDType, "receive-id-type", "chat_id", "receive ID type (chat_id, open_id, user_id, email)")
	cmd.Flags().StringVar(&text, "text", "", "text content")
	cmd.MarkFlagsOneRequired("chat-id", "receive-id")
	_ = cmd.MarkFlagRequired("text")
	return cmd
}
