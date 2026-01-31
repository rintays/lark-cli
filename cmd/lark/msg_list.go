package main

import (
	"context"
	"encoding/json"
	"errors"
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
				if strings.TrimSpace(containerID) == "" {
					return errors.New("container-id is required")
				}
				return nil
			}
			if containerID != "" && containerID != args[0] {
				return errors.New("container-id provided twice")
			}
			return cmd.Flags().Set("container-id", args[0])
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
			text := "no messages found"
			if len(messages) > 0 {
				lines := make([]string, 0, len(messages))
				for _, message := range messages {
					lines = append(lines, formatMessageBlock(message))
				}
				text = strings.Join(lines, "\n\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&containerIDType, "container-id-type", "chat", "container type (chat or thread)")
	cmd.Flags().StringVar(&containerID, "container-id", "", "chat_id or thread_id (or provide as positional argument)")
	cmd.Flags().StringVar(&startTime, "start-time", "", "start time (unix seconds)")
	cmd.Flags().StringVar(&endTime, "end-time", "", "end time (unix seconds)")
	cmd.Flags().StringVar(&sortType, "sort", "ByCreateTimeAsc", "sort type (ByCreateTimeAsc or ByCreateTimeDesc)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of messages to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 20, "page size per request")
	return cmd
}

func formatMessageBlock(message larksdk.Message) string {
	content := messageContentForDisplay(message)
	if strings.TrimSpace(content) == "" {
		content = "(no content)"
	}
	contentLines := strings.Split(content, "\n")
	lines := make([]string, 0, len(contentLines)+1)
	lines = append(lines, contentLines[0])
	for _, line := range contentLines[1:] {
		lines = append(lines, "  "+line)
	}
	meta := formatMessageMeta(message)
	if meta != "" {
		lines = append(lines, "  "+meta)
	}
	return strings.Join(lines, "\n")
}

func formatMessageMeta(message larksdk.Message) string {
	parts := make([]string, 0, 4)
	if message.MessageID != "" {
		parts = append(parts, "id: "+message.MessageID)
	}
	if message.MsgType != "" {
		parts = append(parts, "type: "+message.MsgType)
	}
	if message.ChatID != "" {
		parts = append(parts, "chat: "+message.ChatID)
	}
	if message.CreateTime != "" {
		parts = append(parts, "time: "+message.CreateTime)
	}
	return strings.Join(parts, "  ")
}

func messageContentForDisplay(message larksdk.Message) string {
	raw := strings.TrimSpace(message.Body.Content)
	if raw == "" {
		return ""
	}
	if message.MsgType == "text" {
		var payload struct {
			Text string `json:"text"`
		}
		if err := json.Unmarshal([]byte(raw), &payload); err == nil {
			if text := strings.TrimSpace(payload.Text); text != "" {
				return text
			}
		}
	}
	return raw
}
