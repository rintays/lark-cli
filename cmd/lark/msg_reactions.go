package main

import (
	"context"
	"errors"
	"fmt"

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
		Use:   "add",
		Short: "Add a reaction to a message",
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

	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID")
	cmd.Flags().StringVar(&emojiType, "emoji", "", "emoji type (e.g. SMILE)")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("emoji")
	return cmd
}

func newMsgReactionsDeleteCmd(state *appState) *cobra.Command {
	var messageID string
	var reactionID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a reaction from a message",
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

	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID")
	cmd.Flags().StringVar(&reactionID, "reaction-id", "", "reaction ID")
	_ = cmd.MarkFlagRequired("message-id")
	_ = cmd.MarkFlagRequired("reaction-id")
	return cmd
}
