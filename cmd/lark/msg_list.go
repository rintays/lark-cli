package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxMessagesPageSize = 50

func newMsgListCmd(state *appState) *cobra.Command {
	var containerIDType string
	var containerID string
	var startTime string
	var endTime string
	var sortType string
	var limit int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list <container-id>",
		Short: "List messages in a chat or thread",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("container-id is required")
			}
			containerID = strings.TrimSpace(args[0])
			if containerID == "" {
				return errors.New("container-id is required")
			}
			return nil
		},
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
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			messages := make([]larksdk.Message, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				size := pageSize
				if size > maxMessagesPageSize {
					size = maxMessagesPageSize
				}
				if size > remaining {
					size = remaining
				}
				result, err := state.SDK.ListMessages(context.Background(), token, larksdk.ListMessagesRequest{
					ContainerIDType: containerIDType,
					ContainerID:     containerID,
					StartTime:       startTime,
					EndTime:         endTime,
					SortType:        sortType,
					PageSize:        size,
					PageToken:       pageToken,
				})
				if err != nil {
					return err
				}
				messages = append(messages, result.Items...)
				if len(messages) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(messages)
				pageToken = result.PageToken
				if strings.TrimSpace(pageToken) == "" {
					break
				}
			}
			if len(messages) > limit {
				messages = messages[:limit]
			}
			payload := map[string]any{"messages": messages}
			lines := make([]string, 0, len(messages))
			for _, message := range messages {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", message.MessageID, message.MsgType, message.ChatID, message.CreateTime))
			}
			text := "no messages found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&containerIDType, "container-id-type", "chat", "container type (chat or thread)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "start time (unix seconds)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (unix seconds)")
	cmd.Flags().StringVar(&sortType, "sort", "ByCreateTimeAsc", "sort type (ByCreateTimeAsc or ByCreateTimeDesc)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of messages to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "page size per request")
	return cmd
}
