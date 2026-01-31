package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newMsgPinCmd(state *appState) *cobra.Command {
	var messageID string

	cmd := &cobra.Command{
		Use:   "pin <message-id>",
		Short: "Pin a message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(messageID) == "" {
					return errors.New("message-id is required")
				}
				return nil
			}
			if messageID != "" && messageID != args[0] {
				return errors.New("message-id provided twice")
			}
			return cmd.Flags().Set("message-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			pin, err := state.SDK.PinMessage(context.Background(), token, messageID)
			if err != nil {
				return err
			}
			payload := map[string]any{"pin": pin}
			text := "pinned"
			if pin.MessageID != "" {
				text = fmt.Sprintf("%s\t%s", pin.MessageID, pin.ChatID)
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID (or provide as positional argument)")
	return cmd
}

func newMsgUnpinCmd(state *appState) *cobra.Command {
	var messageID string

	cmd := &cobra.Command{
		Use:   "unpin <message-id>",
		Short: "Unpin a message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(messageID) == "" {
					return errors.New("message-id is required")
				}
				return nil
			}
			if messageID != "" && messageID != args[0] {
				return errors.New("message-id provided twice")
			}
			return cmd.Flags().Set("message-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if err := state.SDK.UnpinMessage(context.Background(), token, messageID); err != nil {
				return err
			}
			payload := map[string]any{"message_id": messageID, "unpinned": true}
			text := "unpinned"
			if messageID != "" {
				text = messageID
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID (or provide as positional argument)")
	return cmd
}
