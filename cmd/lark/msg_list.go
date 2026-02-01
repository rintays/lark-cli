package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
	"lark/internal/output"
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
		Use:     "list <container-id>",
		Short:   "List messages in a chat or thread",
		Example: `  lark messages list <chat_id>`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return usageError(cmd, "container-id is required", `Example:
  lark messages list <chat_id>`)
			}
			containerID = strings.TrimSpace(args[0])
			if containerID == "" {
				return usageError(cmd, "container-id is required", `Example:
  lark messages list <chat_id>`)
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
			text := output.Notice(output.NoticeInfo, "no messages found", nil)
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
	meta := formatMessageMeta(message)
	if meta != "" {
		lines = append(lines, meta)
	}
	lines = append(lines, contentLines...)
	return strings.Join(lines, "\n")
}

func formatMessageMeta(message larksdk.Message) string {
	parts := make([]string, 0, 5)
	if message.MessageID != "" {
		parts = append(parts, "id: "+message.MessageID)
	}
	if sender := formatMessageSender(message.Sender); sender != "" {
		parts = append(parts, "from: "+sender)
	}
	if message.MsgType != "" {
		parts = append(parts, "type: "+message.MsgType)
	}
	if message.CreateTime != "" {
		parts = append(parts, "time: "+message.CreateTime)
	}
	return strings.Join(parts, "  ")
}

func formatMessageSender(sender larksdk.MessageSender) string {
	id := strings.TrimSpace(sender.ID)
	if id == "" {
		return ""
	}
	idType := strings.TrimSpace(sender.IDType)
	if idType == "" {
		idType = "user_id"
	}
	senderType := strings.TrimSpace(sender.SenderType)
	if senderType == "" {
		return fmt.Sprintf("%s:%s", idType, id)
	}
	return fmt.Sprintf("%s:%s:%s", senderType, idType, id)
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
				return applyMessageMentions(text, message.Mentions)
			}
		}
	}
	if rendered, ok := renderMessageTemplate(raw); ok {
		return rendered
	}
	return raw
}

func applyMessageMentions(text string, mentions []larksdk.MessageMention) string {
	if len(mentions) == 0 {
		return text
	}
	seen := make(map[string]larksdk.MessageMention, len(mentions))
	for _, mention := range mentions {
		key := strings.TrimSpace(mention.Key)
		if key == "" {
			key = strings.TrimSpace(mention.Name)
			if key != "" && !strings.HasPrefix(key, "@") {
				key = "@" + key
			}
		}
		if key == "" {
			continue
		}
		if _, exists := seen[key]; exists {
			continue
		}
		seen[key] = mention
	}

	rendered := text
	for key, mention := range seen {
		id := strings.TrimSpace(mention.ID)
		if id == "" {
			continue
		}
		scheme := strings.TrimSpace(mention.IDType)
		if scheme == "" {
			scheme = "user_id"
		}
		link := fmt.Sprintf("[%s](%s://%s)", key, scheme, id)
		rendered = strings.ReplaceAll(rendered, key, link)
	}
	return rendered
}

var templatePlaceholderPattern = regexp.MustCompile(`\{[^}]+\}`)

func renderMessageTemplate(raw string) (string, bool) {
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return "", false
	}
	templateValue, ok := payload["template"].(string)
	if !ok || strings.TrimSpace(templateValue) == "" {
		return "", false
	}
	rendered := templateValue
	for key, value := range payload {
		if key == "template" {
			continue
		}
		placeholder := "{" + key + "}"
		if !strings.Contains(rendered, placeholder) {
			continue
		}
		rendered = strings.ReplaceAll(rendered, placeholder, renderTemplateValue(value))
	}
	rendered = templatePlaceholderPattern.ReplaceAllString(rendered, "")
	return strings.TrimSpace(rendered), true
}

func renderTemplateValue(value any) string {
	switch v := value.(type) {
	case string:
		return v
	case []any:
		parts := make([]string, 0, len(v))
		for _, item := range v {
			if part := strings.TrimSpace(renderTemplateValue(item)); part != "" {
				parts = append(parts, part)
			}
		}
		return strings.Join(parts, ", ")
	case map[string]any:
		if text, ok := v["text"].(string); ok {
			return text
		}
		if name, ok := v["name"].(string); ok {
			return name
		}
		if title, ok := v["title"].(string); ok {
			return title
		}
		if value, ok := v["value"].(string); ok {
			return value
		}
		return ""
	default:
		return fmt.Sprint(v)
	}
}
