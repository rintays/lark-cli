package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxChatsPageSize = 50

func newChatsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "chats",
		Short: "List chats the bot can access",
	}
	cmd.AddCommand(newChatsListCmd(state))
	cmd.AddCommand(newChatsCreateCmd(state))
	cmd.AddCommand(newChatsGetCmd(state))
	cmd.AddCommand(newChatsUpdateCmd(state))
	cmd.AddCommand(newChatsAnnouncementCmd(state))
	return cmd
}

func newChatsListCmd(state *appState) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List recent chats",
		Args:  cobra.NoArgs,
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			chats := make([]larksdk.Chat, 0, limit)
			pageToken := ""
			remaining := limit
			listChats := state.SDK.ListChats
			for {
				pageSize := remaining
				if pageSize > maxChatsPageSize {
					pageSize = maxChatsPageSize
				}
				result, err := listChats(context.Background(), token, larksdk.ListChatsRequest{
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}
				chats = append(chats, result.Items...)
				if len(chats) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(chats)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(chats) > limit {
				chats = chats[:limit]
			}
			payload := map[string]any{"chats": chats}
			lines := make([]string, 0, len(chats))
			for _, chat := range chats {
				lines = append(lines, fmt.Sprintf("%s\t%s", chat.ChatID, chat.Name))
			}
			text := tableText([]string{"chat_id", "name"}, lines, "no chats found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 20, "max number of chats to return")
	return cmd
}
