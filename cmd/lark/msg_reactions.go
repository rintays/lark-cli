package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newMsgReactionsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reactions",
		Short: "Manage message reactions",
	}
	cmd.AddCommand(newMsgReactionsAddCmd(state))
	cmd.AddCommand(newMsgReactionsDeleteCmd(state))
	return cmd
}

func newMsgReactionsAddCmd(state *appState) *cobra.Command {
	var messageID string
	var emojiType string

	cmd := &cobra.Command{
		Use:   "add <message-id> <emoji>",
		Short: "Add a reaction to a message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("message-id is required")
			}
			if len(args) < 2 {
				return errors.New("emoji is required")
			}
			messageID = strings.TrimSpace(args[0])
			emojiType = strings.TrimSpace(args[1])
			if messageID == "" {
				return errors.New("message-id is required")
			}
			if emojiType == "" {
				return errors.New("emoji is required")
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
			reaction, err := state.SDK.CreateMessageReaction(context.Background(), token, messageID, emojiType)
			if err != nil {
				return err
			}
			payload := map[string]any{"reaction": reaction}
			text := "reaction added"
			if reaction.ReactionID != "" {
				text = fmt.Sprintf("reaction_id: %s", reaction.ReactionID)
			}
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}

func newMsgReactionsDeleteCmd(state *appState) *cobra.Command {
	var messageID string
	var reactionID string

	cmd := &cobra.Command{
		Use:   "delete <message-id> <reaction-id>",
		Short: "Delete a reaction from a message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("message-id is required")
			}
			if len(args) < 2 {
				return errors.New("reaction-id is required")
			}
			messageID = strings.TrimSpace(args[0])
			reactionID = strings.TrimSpace(args[1])
			if messageID == "" {
				return errors.New("message-id is required")
			}
			if reactionID == "" {
				return errors.New("reaction-id is required")
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
			reaction, err := state.SDK.DeleteMessageReaction(context.Background(), token, messageID, reactionID)
			if err != nil {
				return err
			}
			payload := map[string]any{"reaction": reaction}
			text := "reaction deleted"
			if reaction.ReactionID != "" {
				text = fmt.Sprintf("reaction_id: %s", reaction.ReactionID)
			}
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}
