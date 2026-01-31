package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newChatsAnnouncementCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "announcement",
		Short: "Manage chat announcements",
	}
	cmd.AddCommand(newChatsAnnouncementGetCmd(state))
	cmd.AddCommand(newChatsAnnouncementUpdateCmd(state))
	return cmd
}

func newChatsAnnouncementGetCmd(state *appState) *cobra.Command {
	var chatID string
	var userIDType string

	cmd := &cobra.Command{
		Use:   "get <chat-id>",
		Short: "Get chat announcement",
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
			announcement, err := state.SDK.GetChatAnnouncement(context.Background(), token, chatID, userIDType)
			if err != nil {
				return err
			}
			payload := map[string]any{"announcement": announcement}
			text := "announcement fetched"
			if announcement.Revision != "" || announcement.Content != "" {
				text = strings.TrimSpace(fmt.Sprintf("%s\t%s", announcement.Revision, announcement.Content))
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "chat ID (or provide as positional argument)")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user ID type (open_id, union_id, user_id)")
	_ = cmd.MarkFlagRequired("chat-id")
	return cmd
}

func newChatsAnnouncementUpdateCmd(state *appState) *cobra.Command {
	var chatID string
	var revision string
	var requests []string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update chat announcement",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if err := state.SDK.UpdateChatAnnouncement(context.Background(), token, chatID, revision, requests); err != nil {
				return err
			}
			payload := map[string]any{"chat_id": chatID, "updated": true}
			return state.Printer.Print(payload, chatID)
		},
	}

	cmd.Flags().StringVar(&chatID, "chat-id", "", "chat ID")
	cmd.Flags().StringVar(&revision, "revision", "", "announcement revision")
	cmd.Flags().StringArrayVar(&requests, "request", nil, "announcement update request (repeatable JSON string)")
	_ = cmd.MarkFlagRequired("chat-id")
	_ = cmd.MarkFlagRequired("revision")
	_ = cmd.MarkFlagRequired("request")
	return cmd
}
