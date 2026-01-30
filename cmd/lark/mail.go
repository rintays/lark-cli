package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
	"lark/internal/larksdk"
)

const maxMailPageSize = 200

func newMailCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Manage Mail messages",
	}
	cmd.AddCommand(newMailFoldersCmd(state))
	cmd.AddCommand(newMailListCmd(state))
	cmd.AddCommand(newMailGetCmd(state))
	cmd.AddCommand(newMailSendCmd(state))
	return cmd
}

func newMailFoldersCmd(state *appState) *cobra.Command {
	var mailboxID string

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List mail folders",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			folders, err := state.SDK.ListMailFolders(context.Background(), token, mailboxID)
			if err != nil {
				return err
			}
			payload := map[string]any{"folders": folders}
			lines := make([]string, 0, len(folders))
			for _, folder := range folders {
				lines = append(lines, formatMailFolderLine(folder))
			}
			text := "no folders found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID")
	_ = cmd.MarkFlagRequired("mailbox-id")
	return cmd
}

func newMailListCmd(state *appState) *cobra.Command {
	var mailboxID string
	var folderID string
	var limit int
	var onlyUnread bool

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mail messages",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			messages := make([]larksdk.MailMessage, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxMailPageSize {
					pageSize = maxMailPageSize
				}
				result, err := state.SDK.ListMailMessages(context.Background(), token, larksdk.ListMailMessagesRequest{
					MailboxID:  mailboxID,
					FolderID:   folderID,
					PageSize:   pageSize,
					PageToken:  pageToken,
					OnlyUnread: onlyUnread,
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
				if pageToken == "" {
					break
				}
			}
			if len(messages) > limit {
				messages = messages[:limit]
			}
			payload := map[string]any{"messages": messages}
			lines := make([]string, 0, len(messages))
			for _, message := range messages {
				lines = append(lines, formatMailMessageLine(message.MessageID, message.Subject))
			}
			text := "no messages found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "filter by folder ID")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of messages to return")
	cmd.Flags().BoolVar(&onlyUnread, "only-unread", false, "only return unread messages")
	_ = cmd.MarkFlagRequired("mailbox-id")
	return cmd
}

func newMailGetCmd(state *appState) *cobra.Command {
	var mailboxID string
	var messageID string

	cmd := &cobra.Command{
		Use:   "get <message-id>",
		Short: "Get a mail message",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if messageID != "" && messageID != args[0] {
					return errors.New("message-id provided twice")
				}
				if err := cmd.Flags().Set("message-id", args[0]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			message, err := state.SDK.GetMailMessage(context.Background(), token, mailboxID, messageID)
			if err != nil {
				return err
			}
			payload := map[string]any{"message": message}
			return state.Printer.Print(payload, formatMailMessageLine(message.MessageID, message.Subject))
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID")
	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("mailbox-id")
	_ = cmd.MarkFlagRequired("message-id")
	return cmd
}

func newMailSendCmd(state *appState) *cobra.Command {
	var mailboxID string
	var subject string
	var to []string
	var cc []string
	var bcc []string
	var bodyText string
	var bodyHTML string
	var headFrom string
	var userAccessToken string
	const userTokenHint = "mail send requires a user access token; pass --user-access-token, set LARK_USER_ACCESS_TOKEN, or run `lark auth user login`"

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email message",
		RunE: func(cmd *cobra.Command, args []string) error {
			token := userAccessToken
			if token == "" {
				token = os.Getenv("LARK_USER_ACCESS_TOKEN")
			}
			if token == "" {
				var err error
				token, err = ensureUserToken(context.Background(), state)
				if err != nil || token == "" {
					return errors.New(userTokenHint)
				}
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			request := larksdk.SendMailRequest{
				Subject:       subject,
				To:            buildMailAddressInputs(to),
				CC:            buildMailAddressInputs(cc),
				BCC:           buildMailAddressInputs(bcc),
				HeadFromName:  headFrom,
				BodyPlainText: bodyText,
				BodyHTML:      bodyHTML,
			}
			messageID, err := state.SDK.SendMail(context.Background(), token, mailboxID, request)
			if err != nil {
				return err
			}
			payload := map[string]any{"message_id": messageID}
			return state.Printer.Print(payload, fmt.Sprintf("message_id: %s", messageID))
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID")
	cmd.Flags().StringVar(&subject, "subject", "", "message subject")
	cmd.Flags().StringArrayVar(&to, "to", nil, "recipient email (repeatable)")
	cmd.Flags().StringArrayVar(&cc, "cc", nil, "cc email (repeatable)")
	cmd.Flags().StringArrayVar(&bcc, "bcc", nil, "bcc email (repeatable)")
	cmd.Flags().StringVar(&bodyText, "text", "", "plain text body")
	cmd.Flags().StringVar(&bodyHTML, "html", "", "HTML body")
	cmd.Flags().StringVar(&headFrom, "from-name", "", "display name for From header")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")
	_ = cmd.MarkFlagRequired("mailbox-id")
	_ = cmd.MarkFlagRequired("subject")
	_ = cmd.MarkFlagRequired("to")
	cmd.MarkFlagsOneRequired("text", "html")
	return cmd
}

func formatMailFolderLine(folder larkapi.MailFolder) string {
	parts := []string{folder.FolderID, folder.Name}
	if folder.FolderType != "" {
		parts = append(parts, folder.FolderType)
	}
	return strings.Join(parts, "\t")
}

func formatMailMessageLine(messageID, subject string) string {
	subject = strings.TrimSpace(subject)
	if subject == "" {
		subject = "(no subject)"
	}
	return fmt.Sprintf("%s\t%s", messageID, subject)
}

func buildMailAddressInputs(values []string) []larksdk.MailAddressInput {
	if len(values) == 0 {
		return nil
	}
	addresses := make([]larksdk.MailAddressInput, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		addresses = append(addresses, larksdk.MailAddressInput{MailAddress: value})
	}
	return addresses
}
