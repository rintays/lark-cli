package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"
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
				return argsUsageError(cmd, err)
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
			if _, err := requireSDK(state); err != nil {
				return err
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenType, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
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
				result, err := state.SDK.ListMessages(cmd.Context(), token, larksdk.ListMessagesRequest{
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
				var senderNames map[string]string
				if !state.JSON && tokenType == tokenTypeTenant {
					senderNames = resolveMessageSenderNames(cmd.Context(), state, token, messages)
				}
				styles := newMessageFormatStyles(state.Printer.Styled)
				displays := make([]messageDisplay, 0, len(messages))
				for _, message := range messages {
					displays = append(displays, buildMessageDisplay(message, styles, senderNames))
				}
				text = renderMessageTable(displays, styles)
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

type messageFormatStyles struct {
	styled  bool
	content lipgloss.Style
	sender  lipgloss.Style
	time    lipgloss.Style
	dim     lipgloss.Style
	meta    lipgloss.Style
	border  lipgloss.Style
}

func newMessageFormatStyles(styled bool) messageFormatStyles {
	if !styled {
		return messageFormatStyles{styled: false}
	}
	brand := output.BrandColor()
	dim := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "240"})
	strong := lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "0", Dark: "15"})
	meta := lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "61", Dark: "74"})
	return messageFormatStyles{
		styled:  true,
		content: strong,
		sender:  lipgloss.NewStyle().Bold(true).Foreground(brand),
		time:    dim,
		dim:     dim,
		meta:    meta,
		border:  dim,
	}
}

func (s messageFormatStyles) renderContent(text string) string {
	if !s.styled {
		return text
	}
	return s.content.Render(text)
}

func (s messageFormatStyles) renderSender(text string) string {
	if !s.styled {
		return text
	}
	return s.sender.Render(text)
}

func (s messageFormatStyles) renderTime(text string) string {
	if !s.styled {
		return text
	}
	return s.time.Render(text)
}

func (s messageFormatStyles) renderDim(text string) string {
	if !s.styled {
		return text
	}
	return s.dim.Render(text)
}

func (s messageFormatStyles) renderMeta(text string) string {
	if !s.styled {
		return text
	}
	return s.meta.Render(text)
}

type messageDisplay struct {
	leftStyled []string
	rightLines []string
}

func buildMessageDisplay(message larksdk.Message, styles messageFormatStyles, senderNames map[string]string) messageDisplay {
	leftStyled := formatMessageSenderLines(message, styles, senderNames)
	content := messageContentForDisplay(message)
	contentLines := normalizeMessageContentLines(content)
	if len(contentLines) == 0 {
		contentLines = []string{"(no content)"}
	}
	rightLines := make([]string, 0, len(contentLines)+2)
	rightLines = append(rightLines, styles.renderContent(contentLines[0]))
	meta := formatMessageMetaLine(message, styles)
	if meta == "" {
		meta = styles.renderMeta("type - | message id: -")
	}
	rightLines = append(rightLines, meta)
	for _, line := range contentLines[1:] {
		rightLines = append(rightLines, styles.renderContent(line))
	}
	return messageDisplay{
		leftStyled: leftStyled,
		rightLines: rightLines,
	}
}

func renderMessageTable(displays []messageDisplay, styles messageFormatStyles) string {
	if len(displays) == 0 {
		return ""
	}
	rows := make([][]string, 0, len(displays))
	for _, display := range displays {
		left := strings.Join(display.leftStyled, "\n")
		right := strings.Join(display.rightLines, "\n")
		rows = append(rows, []string{left, right})
	}
	table := liptable.New().
		Rows(rows...).
		Border(lipgloss.NormalBorder()).
		BorderTop(false).
		BorderBottom(false).
		BorderLeft(false).
		BorderRight(false).
		BorderHeader(false).
		BorderColumn(true).
		BorderRow(true).
		BorderStyle(styles.border)
	return table.Render()
}

func formatMessageSenderLines(message larksdk.Message, styles messageFormatStyles, senderNames map[string]string) []string {
	timeLabel := formatMessageTime(message.CreateTime)
	if timeLabel == "" {
		timeLabel = "-"
	}
	name, idLabel := messageSenderDisplay(message, senderNames)
	if name == "" {
		name = "unknown"
	}
	line1Styled := styles.renderTime("["+timeLabel+"]") + " " + styles.renderSender(name)
	line2Styled := styles.renderDim("sender -")
	if idLabel != "" {
		line2Styled = styles.renderDim("sender " + idLabel)
	}
	return []string{line1Styled, line2Styled}
}

func messageSenderDisplay(message larksdk.Message, senderNames map[string]string) (string, string) {
	sender := message.Sender
	id := strings.TrimSpace(sender.ID)
	idType := strings.TrimSpace(sender.IDType)
	if idType == "" {
		idType = "user_id"
	}
	senderType := strings.TrimSpace(sender.SenderType)
	if senderType == "" {
		if message.MsgType == "system" {
			senderType = "system"
		} else {
			senderType = "user"
		}
	}
	displayName := ""
	if senderNames != nil && id != "" {
		displayName = senderNames[messageSenderKey(senderType, idType, id)]
	}
	if displayName == "" {
		displayName = senderType
	}
	idLabel := ""
	if id != "" {
		idLabel = fmt.Sprintf("%s://%s", idType, id)
	}
	return displayName, idLabel
}

func formatMessageMetaLine(message larksdk.Message, styles messageFormatStyles) string {
	parts := make([]string, 0, 3)
	if message.MsgType != "" {
		parts = append(parts, "type "+message.MsgType)
	}
	if message.MessageID != "" {
		parts = append(parts, "message id: "+message.MessageID)
	}
	if len(parts) == 0 {
		return ""
	}
	return styles.renderMeta(strings.Join(parts, " | "))
}

func messageSenderKey(senderType, idType, id string) string {
	return fmt.Sprintf("%s:%s:%s", senderType, idType, id)
}

func resolveMessageSenderNames(ctx context.Context, state *appState, token string, messages []larksdk.Message) map[string]string {
	if state == nil || state.SDK == nil || len(messages) == 0 {
		return nil
	}
	type lookup struct {
		key    string
		id     string
		idType string
	}
	lookups := make(map[string]lookup)
	for _, message := range messages {
		sender := message.Sender
		id := strings.TrimSpace(sender.ID)
		if id == "" {
			continue
		}
		senderType := strings.TrimSpace(sender.SenderType)
		if senderType == "" {
			senderType = "user"
		}
		if senderType != "user" {
			continue
		}
		idType := strings.TrimSpace(sender.IDType)
		if idType == "" {
			idType = "user_id"
		}
		key := messageSenderKey(senderType, idType, id)
		if _, exists := lookups[key]; exists {
			continue
		}
		lookups[key] = lookup{key: key, id: id, idType: idType}
	}
	if len(lookups) == 0 {
		return nil
	}
	names := make(map[string]string, len(lookups))
	for _, item := range lookups {
		user, err := state.SDK.GetContactUser(ctx, token, larksdk.GetContactUserRequest{
			UserID:     item.id,
			UserIDType: item.idType,
		})
		if err != nil {
			continue
		}
		name := strings.TrimSpace(user.Name)
		if name == "" {
			continue
		}
		names[item.key] = name
	}
	return names
}

func formatMessageTime(raw string) string {
	trimmed := strings.TrimSpace(raw)
	if trimmed == "" {
		return ""
	}
	value, err := strconv.ParseInt(trimmed, 10, 64)
	if err != nil {
		return trimmed
	}
	var t time.Time
	if value > 1e11 || len(trimmed) >= 13 {
		t = time.UnixMilli(value)
	} else {
		t = time.Unix(value, 0)
	}
	if t.IsZero() {
		return trimmed
	}
	return t.Local().Format("2006-01-02 15:04:05")
}

func normalizeMessageContentLines(content string) []string {
	if strings.TrimSpace(content) == "" {
		return nil
	}
	lines := strings.Split(content, "\n")
	return trimEmptyLines(lines)
}

func trimEmptyLines(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	if start >= end {
		return nil
	}
	return lines[start:end]
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
