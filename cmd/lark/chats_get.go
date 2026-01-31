package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newChatsGetCmd(state *appState) *cobra.Command {
	var chatID string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get <chat-id>",
		Short: "Get chat information",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if chatID != "" && chatID != args[0] {
					return errors.New("chat-id provided twice")
				}
				if err := cmd.Flags().Set("chat-id", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			chat, err := state.SDK.GetChatInfo(context.Background(), token, larksdk.GetChatRequest{
				ChatID:     chatID,
				UserIDType: userIDType,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"chat": chat}
			text := chat.ChatID
			if chat.Name != "" {
				text = fmt.Sprintf("%s\t%s", chat.ChatID, chat.Name)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "chat ID (or provide as positional argument)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	_ = cmd.MarkFlagRequired("chat-id")
	return cmd
}
