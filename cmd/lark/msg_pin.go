package main

import (
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
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			pin, err := state.SDK.PinMessage(cmd.Context(), token, messageID)
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

	return cmd
}

func newMsgUnpinCmd(state *appState) *cobra.Command {
	var messageID string

	cmd := &cobra.Command{
		Use:   "unpin <message-id>",
		Short: "Unpin a message",
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
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if err := state.SDK.UnpinMessage(cmd.Context(), token, messageID); err != nil {
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

	return cmd
}
