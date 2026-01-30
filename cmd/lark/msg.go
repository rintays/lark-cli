package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
)

func newMsgCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "msg",
		Short: "Send messages",
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
			if receiveID == "" {
				return errors.New("receive_id is required")
			}
			if text == "" {
				return errors.New("text is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			messageID, err := state.Client.SendMessage(context.Background(), token, larkapi.MessageRequest{
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
	return cmd
}
