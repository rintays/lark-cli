package main

import (
	"context"
	"errors"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxMessageSearchPageSize = 100

func newMsgSearchCmd(state *appState) *cobra.Command {
	var query string
	var fromIDs []string
	var chatIDs []string
	var messageType string
	var atChatterIDs []string
	var fromType string
	var chatType string
	var startTime string
	var endTime string
	var limit int
	var pageSize int
	var userIDType string
	var userAccessToken string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search messages by keyword",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pageSize <= 0 {
				return errors.New("page-size must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token := strings.TrimSpace(userAccessToken)
			if token == "" {
				token = strings.TrimSpace(os.Getenv("LARK_USER_ACCESS_TOKEN"))
			}
			if token != "" {
				var err error
				token, err = tokenForOverride(context.Background(), state, tokenTypesUser, tokenOverride{
					Token: token,
					Type:  tokenTypeUser,
				})
				if err != nil {
					return err
				}
			} else {
				var err error
				token, err = tokenFor(context.Background(), state, tokenTypesUser)
				if err != nil {
					return err
				}
			}

			items := make([]string, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				size := pageSize
				if size > maxMessageSearchPageSize {
					size = maxMessageSearchPageSize
				}
				if size > remaining {
					size = remaining
				}
				result, err := state.SDK.SearchMessages(context.Background(), token, larksdk.MessageSearchRequest{
					Query:        query,
					FromIDs:      fromIDs,
					ChatIDs:      chatIDs,
					MessageType:  messageType,
					AtChatterIDs: atChatterIDs,
					FromType:     fromType,
					ChatType:     chatType,
					StartTime:    startTime,
					EndTime:      endTime,
					PageSize:     size,
					PageToken:    pageToken,
					UserIDType:   userIDType,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				items = append(items, result.Items...)
				if len(items) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(items)
				pageToken = result.PageToken
				if strings.TrimSpace(pageToken) == "" {
					break
				}
			}
			if len(items) > limit {
				items = items[:limit]
			}

			payload := map[string]any{"message_ids": items}
			text := "no messages found"
			if len(items) > 0 {
				text = strings.Join(items, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search keyword")
	cmd.Flags().StringArrayVar(&fromIDs, "from-id", nil, "filter by sender IDs (repeatable)")
	cmd.Flags().StringArrayVar(&chatIDs, "chat-id", nil, "filter by chat IDs (repeatable)")
	cmd.Flags().StringVar(&messageType, "message-type", "", "message type (file, image, media)")
	cmd.Flags().StringArrayVar(&atChatterIDs, "at-id", nil, "filter by @ user IDs (repeatable)")
	cmd.Flags().StringVar(&fromType, "from-type", "", "sender type (bot or user)")
	cmd.Flags().StringVar(&chatType, "chat-type", "", "chat type (group_chat or p2p_chat)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "start time (unix seconds)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (unix seconds)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of message IDs to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "page size per request")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user id type (open_id, union_id, user_id)")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")
	_ = cmd.MarkFlagRequired("query")
	return cmd
}
