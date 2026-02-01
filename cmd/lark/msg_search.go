package main

import (
	"errors"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
	"lark/internal/output"
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
	var pages int
	var userIDType string
	var userAccessToken string

	cmd := &cobra.Command{
		Use:     "search <query>",
		Short:   "Search messages by keyword",
		Example: `  lark messages search "hello"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) == 0 {
				return usageError(cmd, "query is required", `Example:
  lark messages search "hello"`)
			}
			query = strings.TrimSpace(args[0])
			if query == "" {
				return usageError(cmd, "query is required", `Example:
  lark messages search "hello"`)
			}
			return nil
		},
		PreRunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pageSize < 0 {
				return errors.New("page-size must be greater than or equal to 0")
			}
			if pages <= 0 {
				return errors.New("pages must be greater than 0")
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			now := time.Now()
			parsedStart, err := parseTimeArg(startTime, now)
			if err != nil {
				return usageError(cmd, "invalid start-time", err.Error())
			}
			parsedEnd, err := parseTimeArg(endTime, now)
			if err != nil {
				return usageError(cmd, "invalid end-time", err.Error())
			}
			startTime = parsedStart
			endTime = parsedEnd

			token := strings.TrimSpace(userAccessToken)
			if token == "" {
				token = strings.TrimSpace(os.Getenv("LARK_USER_ACCESS_TOKEN"))
			}
			if token != "" {
				var err error
				token, err = tokenForOverride(cmd.Context(), state, tokenTypesUser, tokenOverride{
					Token: token,
					Type:  tokenTypeUser,
				})
				if err != nil {
					return err
				}
			} else {
				var err error
				token, err = tokenFor(cmd.Context(), state, tokenTypesUser)
				if err != nil {
					return err
				}
			}

			items := make([]string, 0, limit)
			pageToken := ""
			pageCount := 0
			remaining := limit
			for {
				if pageCount >= pages {
					break
				}
				pageCount++
				size := remaining
				if pageSize > 0 {
					size = pageSize
				}
				if size > maxMessageSearchPageSize {
					size = maxMessageSearchPageSize
				}
				if size > remaining {
					size = remaining
				}
				if size <= 0 {
					break
				}
				result, err := state.SDK.SearchMessages(cmd.Context(), token, larksdk.MessageSearchRequest{
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

			messages := make([]larksdk.Message, 0, len(items))
			for _, messageID := range items {
				message, err := state.SDK.GetMessage(cmd.Context(), token, messageID, userIDType)
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				messages = append(messages, message)
			}

			payload := map[string]any{
				"message_ids": items,
				"messages":    messages,
			}
			text := output.Notice(output.NoticeInfo, "no messages found", nil)
			if len(messages) > 0 {
				lines := make([]string, 0, len(messages))
				styles := newMessageFormatStyles(state.Printer.Styled)
				for _, message := range messages {
					lines = append(lines, formatMessageBlock(message, styles))
				}
				text = strings.Join(lines, "\n\n")
			}
			return state.Printer.Print(payload, text)
		},
	}
	annotateAuthServices(cmd, "search-message")

	cmd.Flags().StringSliceVar(&fromIDs, "from-id", nil, "filter by sender IDs (repeatable or comma-separated)")
	cmd.Flags().StringSliceVar(&chatIDs, "chat-id", nil, "filter by chat IDs (repeatable or comma-separated)")
	cmd.Flags().StringVar(&messageType, "message-type", "", "message type (file, image, media)")
	cmd.Flags().StringSliceVar(&atChatterIDs, "at-id", nil, "filter by @ user IDs (repeatable or comma-separated)")
	cmd.Flags().StringVar(&fromType, "from-type", "", "sender type (bot or user)")
	cmd.Flags().StringVar(&chatType, "chat-type", "", "chat type (group_chat or p2p_chat)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "start time (unix seconds, RFC3339, or relative like -24h/-7d)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (unix seconds, RFC3339, or relative like -24h/-7d)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of message IDs to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size per request (default: auto)")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")
	cmd.Flags().StringVar(&userIDType, "user-id-type", "open_id", "user id type (open_id, union_id, user_id)")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")
	return cmd
}
