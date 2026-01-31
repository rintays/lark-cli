package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const maxMailPageSize = 200

func newMailCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mail",
		Short: "Manage Mail messages",
	}
	cmd.AddCommand(newMailMailboxCmd(state))
	cmd.AddCommand(newMailPublicMailboxesCmd(state))
	// Backwards-compatible alias: historically users ran `lark mail mailboxes list`.
	// The user_mailboxes list endpoint is not available in Feishu OpenAPI, so we
	// map this to public mailboxes discovery.
	cmd.AddCommand(newMailMailboxesCmd(state))
	cmd.AddCommand(newMailFoldersCmd(state))
	cmd.AddCommand(newMailListCmd(state))
	cmd.AddCommand(newMailInfoCmd(state))
	cmd.AddCommand(newMailGetCmd(state))
	cmd.AddCommand(newMailSendCmd(state))
	return cmd
}

func newMailMailboxCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailbox",
		Short: "Manage a mailbox",
	}
	cmd.AddCommand(newMailMailboxInfoCmd(state))
	cmd.AddCommand(newMailMailboxSetCmd(state))
	return cmd
}

func newMailMailboxInfoCmd(state *appState) *cobra.Command {
	var mailboxID string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show mailbox details",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			// Default mailbox resolution: flag > config default > "me".
			mailboxID = resolveMailboxID(state, mailboxID)

			mailbox, err := state.SDK.GetMailbox(context.Background(), token, mailboxID)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			payload := map[string]any{"mailbox": mailbox}
			text := formatMailMailboxInfo(mailbox)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
	// mailbox-id is optional; defaults to config default_mailbox_id or me.
	// (still allow explicit mailbox-id when needed)

	return cmd
}

func newMailMailboxSetCmd(state *appState) *cobra.Command {
	var mailboxID string

	cmd := &cobra.Command{
		Use:   "set <mailbox-id>",
		Short: "Set the default mailbox",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(mailboxID) == "" {
					return errors.New("mailbox-id is required")
				}
				return nil
			}
			if mailboxID != "" && mailboxID != args[0] {
				return errors.New("mailbox-id provided twice")
			}
			return cmd.Flags().Set("mailbox-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			state.Config.DefaultMailboxID = mailboxID
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path":        state.ConfigPath,
				"default_mailbox_id": mailboxID,
			}
			return state.Printer.Print(payload, fmt.Sprintf("default mailbox set to %s", mailboxID))
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me'; or provide as positional argument)")
	return cmd
}

func newMailPublicMailboxesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "public-mailboxes",
		Short: "Discover public mailboxes",
	}
	cmd.AddCommand(newMailPublicMailboxesListCmd(state))
	return cmd
}

func newMailMailboxesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "mailboxes",
		Short: "List mailboxes (alias for public-mailboxes)",
		Long:  "This is a backwards-compatible alias. Feishu OpenAPI does not provide a user mailbox list endpoint; use public mailboxes discovery or pass --mailbox-id me for user mailbox operations.",
	}
	cmd.AddCommand(newMailMailboxesListCmd(state))
	return cmd
}

func newMailMailboxesListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List mailboxes (public mailboxes)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			mailboxes, err := state.SDK.ListPublicMailboxes(context.Background(), token)
			if err != nil {
				return err
			}
			payload := map[string]any{"public_mailboxes": mailboxes}
			lines := make([]string, 0, len(mailboxes))
			for _, mailbox := range mailboxes {
				lines = append(lines, formatMailMailboxLine(mailbox))
			}
			text := tableText([]string{"mailbox_id", "name", "address"}, lines, "no public mailboxes found")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newMailPublicMailboxesListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List public mailboxes",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			mailboxes, err := state.SDK.ListPublicMailboxes(context.Background(), token)
			if err != nil {
				return err
			}
			payload := map[string]any{"public_mailboxes": mailboxes}
			lines := make([]string, 0, len(mailboxes))
			for _, mailbox := range mailboxes {
				lines = append(lines, formatMailMailboxLine(mailbox))
			}
			text := tableText([]string{"mailbox_id", "name", "address"}, lines, "no public mailboxes found")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newMailFoldersCmd(state *appState) *cobra.Command {
	var mailboxID string

	cmd := &cobra.Command{
		Use:   "folders",
		Short: "List mail folders",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			mailboxID = resolveMailboxID(state, mailboxID)
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			folders, err := state.SDK.ListMailFolders(context.Background(), token, mailboxID)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			payload := map[string]any{"folders": folders}
			lines := make([]string, 0, len(folders))
			for _, folder := range folders {
				lines = append(lines, formatMailFolderLine(folder))
			}
			text := tableText([]string{"folder_id", "name", "type"}, lines, "no folders found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
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
			mailboxID = resolveMailboxID(state, mailboxID)
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			folderID, err = resolveMailFolderID(context.Background(), state, token, mailboxID, folderID)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			debugf(state, "mail list: mailbox_id=%q folder_id=%q limit=%d only_unread=%t\n", mailboxID, folderID, limit, onlyUnread)
			ctx := context.Background()
			messages := make([]larksdk.MailMessage, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxMailPageSize {
					pageSize = maxMailPageSize
				}
				result, err := state.SDK.ListMailMessages(ctx, token, larksdk.ListMailMessagesRequest{
					MailboxID:  mailboxID,
					FolderID:   folderID,
					PageSize:   pageSize,
					PageToken:  pageToken,
					OnlyUnread: onlyUnread,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				for _, message := range result.Items {
					if len(messages) >= limit {
						break
					}
					if message.MessageID == "" {
						continue
					}
					item, err := state.SDK.GetMailMessage(ctx, token, mailboxID, message.MessageID)
					if err != nil {
						return withUserScopeHintForCommand(state, err)
					}
					if item.MessageID == "" {
						item.MessageID = message.MessageID
					}
					stripMailMessageContent(&item)
					messages = append(messages, item)
				}
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
				lines = append(lines, formatMailMessageListLine(message))
			}
			text := tableText([]string{"message_id", "subject", "from", "internal_date"}, lines, "no messages found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "filter by folder ID (system aliases: INBOX/SENT/DRAFT/TRASH/SPAM/ARCHIVED)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max number of messages to return")
	cmd.Flags().BoolVar(&onlyUnread, "only-unread", false, "only return unread messages")
	return cmd
}

func newMailInfoCmd(state *appState) *cobra.Command {
	var mailboxID string
	var messageID string

	cmd := &cobra.Command{
		Use:   "info <message-id>",
		Short: "Show mail message metadata",
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
			mailboxID = resolveMailboxID(state, mailboxID)
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			message, err := state.SDK.GetMailMessage(context.Background(), token, mailboxID, messageID)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			stripMailMessageContent(&message)
			payload := map[string]any{"message": message}
			text := formatMailMessageInfo(message)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("message-id")
	return cmd
}

func newMailGetCmd(state *appState) *cobra.Command {
	var mailboxID string
	var messageID string

	cmd := &cobra.Command{
		Use:   "get <message-id>",
		Short: "Get a mail message (full content)",
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
			mailboxID = resolveMailboxID(state, mailboxID)
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesUser)
			if err != nil {
				return err
			}
			message, err := state.SDK.GetMailMessage(context.Background(), token, mailboxID, messageID)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			payload := map[string]any{"message": message}
			text := formatMailMessageGet(message)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
	cmd.Flags().StringVar(&messageID, "message-id", "", "message ID (or provide as positional argument)")
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
	var raw string
	var rawFile string

	cmd := &cobra.Command{
		Use:   "send",
		Short: "Send an email message",
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

			// Default mailbox resolution: flag > config default > "me".
			mailboxID = resolveMailboxID(state, mailboxID)

			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			rawValue := strings.TrimSpace(raw)
			rawFile = strings.TrimSpace(rawFile)
			if rawValue != "" && rawFile != "" {
				return errors.New("raw and raw-file are mutually exclusive")
			}
			if rawFile != "" {
				data, readErr := os.ReadFile(rawFile)
				if readErr != nil {
					return fmt.Errorf("read raw file: %w", readErr)
				}
				rawValue = base64.URLEncoding.EncodeToString(data)
			}
			if rawValue != "" {
				if subject != "" || len(to) > 0 || len(cc) > 0 || len(bcc) > 0 || bodyText != "" || bodyHTML != "" {
					return errors.New("raw is mutually exclusive with subject/to/cc/bcc/text/html")
				}
			} else {
				if subject == "" {
					return errors.New("subject is required")
				}
				if len(to) == 0 {
					return errors.New("to is required")
				}
				if bodyText == "" && bodyHTML == "" {
					return errors.New("text or html is required")
				}
			}

			toInputs := buildMailAddressInputs(to)
			ccInputs := buildMailAddressInputs(cc)
			bccInputs := buildMailAddressInputs(bcc)
			request := larksdk.SendMailRequest{
				Subject:       subject,
				To:            toInputs,
				CC:            ccInputs,
				BCC:           bccInputs,
				HeadFromName:  headFrom,
				BodyPlainText: bodyText,
				BodyHTML:      bodyHTML,
				Raw:           rawValue,
			}
			messageID, err := state.SDK.SendMail(context.Background(), token, mailboxID, request)
			if err != nil {
				return withUserScopeHintForCommand(state, err)
			}
			payload := map[string]any{"message_id": messageID}
			return state.Printer.Print(payload, fmt.Sprintf("message_id: %s", messageID))
		},
	}

	cmd.Flags().StringVar(&mailboxID, "mailbox-id", "", "user mailbox ID (defaults to config default_mailbox_id or 'me')")
	cmd.Flags().StringVar(&subject, "subject", "", "message subject")
	cmd.Flags().StringArrayVar(&to, "to", nil, "recipient email (repeatable)")
	cmd.Flags().StringArrayVar(&cc, "cc", nil, "cc email (repeatable)")
	cmd.Flags().StringArrayVar(&bcc, "bcc", nil, "bcc email (repeatable)")
	cmd.Flags().StringVar(&bodyText, "text", "", "plain text body")
	cmd.Flags().StringVar(&bodyHTML, "html", "", "HTML body")
	cmd.Flags().StringVar(&raw, "raw", "", "raw EML content (base64url-encoded)")
	cmd.Flags().StringVar(&rawFile, "raw-file", "", "path to .eml file (will be base64url-encoded)")
	cmd.Flags().StringVar(&headFrom, "from-name", "", "display name for From header")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")
	return cmd
}

func formatMailFolderLine(folder larksdk.MailFolder) string {
	id := folder.FolderID
	if id == "" {
		id = "-"
	}
	folderType := folder.FolderType.String()
	parts := []string{id, folder.Name, folderType}
	return strings.Join(parts, "\t")
}

func formatMailMailboxInfo(mailbox larksdk.Mailbox) string {
	rows := [][]string{
		{"mailbox_id", infoValue(mailbox.MailboxID)},
		{"name", infoValue(mailbox.Name)},
		{"display_name", infoValue(mailbox.DisplayName)},
		{"mail_address", infoValue(mailbox.MailAddress)},
		{"primary_email", infoValue(mailbox.PrimaryEmail)},
		{"email", infoValue(mailbox.Email)},
		{"user_id", infoValue(mailbox.UserID)},
		{"mailbox_status", infoValue(mailbox.MailboxStatus)},
	}
	return formatInfoTable(rows, "no mailbox found")
}

func stripMailMessageContent(message *larksdk.MailMessage) {
	if message == nil {
		return
	}
	message.Raw = ""
	message.BodyHTML = ""
	message.BodyPlainText = ""
	if len(message.Attachments) == 0 {
		return
	}
	for i := range message.Attachments {
		message.Attachments[i].Body = ""
	}
}

func formatMailMessageInfo(message larksdk.MailMessage) string {
	rows := [][]string{
		{"message_id", infoValue(message.MessageID)},
		{"thread_id", infoValue(message.ThreadID)},
		{"subject", infoValue(message.Subject)},
		{"snippet", infoValue(message.Snippet)},
		{"folder_id", infoValue(message.FolderID)},
		{"internal_date", infoValue(message.InternalDate)},
		{"message_state", infoValueIntZeroDash(message.MessageState)},
		{"smtp_message_id", infoValue(message.SMTPMessageID)},
		{"from.mail_address", infoValue(message.From.MailAddress)},
		{"from.name", infoValue(message.From.Name)},
		{"to.count", fmt.Sprintf("%d", len(message.To))},
	}
	for i, addr := range message.To {
		prefix := fmt.Sprintf("to[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"cc.count", fmt.Sprintf("%d", len(message.CC))})
	for i, addr := range message.CC {
		prefix := fmt.Sprintf("cc[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"bcc.count", fmt.Sprintf("%d", len(message.BCC))})
	for i, addr := range message.BCC {
		prefix := fmt.Sprintf("bcc[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"attachments.count", fmt.Sprintf("%d", len(message.Attachments))})
	for i, attachment := range message.Attachments {
		prefix := fmt.Sprintf("attachments[%d]", i)
		rows = append(rows,
			[]string{prefix + ".id", infoValue(attachment.ID)},
			[]string{prefix + ".filename", infoValue(attachment.Filename)},
			[]string{prefix + ".attachment_type", infoValueIntZeroDash(attachment.AttachmentType)},
			[]string{prefix + ".is_inline", fmt.Sprintf("%t", attachment.IsInline)},
			[]string{prefix + ".cid", infoValue(attachment.CID)},
		)
	}
	return formatInfoTable(rows, "no message found")
}

func formatMailMessageGet(message larksdk.MailMessage) string {
	rows := [][]string{
		{"message_id", infoValue(message.MessageID)},
		{"thread_id", infoValue(message.ThreadID)},
		{"subject", infoValue(message.Subject)},
		{"snippet", infoValue(message.Snippet)},
		{"folder_id", infoValue(message.FolderID)},
		{"internal_date", infoValue(message.InternalDate)},
		{"message_state", infoValueIntZeroDash(message.MessageState)},
		{"smtp_message_id", infoValue(message.SMTPMessageID)},
		{"raw", infoValue(message.Raw)},
		{"body_html", infoValue(message.BodyHTML)},
		{"body_plain_text", infoValue(message.BodyPlainText)},
		{"from.mail_address", infoValue(message.From.MailAddress)},
		{"from.name", infoValue(message.From.Name)},
		{"to.count", fmt.Sprintf("%d", len(message.To))},
	}
	for i, addr := range message.To {
		prefix := fmt.Sprintf("to[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"cc.count", fmt.Sprintf("%d", len(message.CC))})
	for i, addr := range message.CC {
		prefix := fmt.Sprintf("cc[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"bcc.count", fmt.Sprintf("%d", len(message.BCC))})
	for i, addr := range message.BCC {
		prefix := fmt.Sprintf("bcc[%d]", i)
		rows = append(rows,
			[]string{prefix + ".mail_address", infoValue(addr.MailAddress)},
			[]string{prefix + ".name", infoValue(addr.Name)},
		)
	}
	rows = append(rows, []string{"attachments.count", fmt.Sprintf("%d", len(message.Attachments))})
	for i, attachment := range message.Attachments {
		prefix := fmt.Sprintf("attachments[%d]", i)
		rows = append(rows,
			[]string{prefix + ".id", infoValue(attachment.ID)},
			[]string{prefix + ".filename", infoValue(attachment.Filename)},
			[]string{prefix + ".attachment_type", infoValueIntZeroDash(attachment.AttachmentType)},
			[]string{prefix + ".is_inline", fmt.Sprintf("%t", attachment.IsInline)},
			[]string{prefix + ".cid", infoValue(attachment.CID)},
			[]string{prefix + ".body", infoValue(attachment.Body)},
		)
	}
	return formatInfoTable(rows, "no message found")
}

func formatMailMessageListLine(message larksdk.MailMessage) string {
	subject := strings.TrimSpace(message.Subject)
	if subject == "" {
		subject = "(no subject)"
	}
	from := formatMailAddressLine(message.From)
	if from == "" {
		from = "-"
	}
	internalDate := strings.TrimSpace(message.InternalDate)
	if internalDate == "" {
		internalDate = "-"
	}
	return strings.Join([]string{message.MessageID, subject, from, internalDate}, "\t")
}

func formatMailAddressLine(addr larksdk.MailAddress) string {
	name := strings.TrimSpace(addr.Name)
	mail := strings.TrimSpace(addr.MailAddress)
	if name != "" && mail != "" {
		return fmt.Sprintf("%s <%s>", name, mail)
	}
	if name != "" {
		return name
	}
	return mail
}

func formatMailMailboxLine(mailbox larksdk.Mailbox) string {
	primary := mailbox.DisplayName
	if primary == "" {
		primary = mailbox.Name
	}
	address := mailbox.PrimaryEmail
	if address == "" {
		address = mailbox.MailAddress
	}
	if address == "" {
		address = mailbox.Email
	}
	id := mailbox.MailboxID
	if id == "" {
		id = address
	}
	if primary == id {
		primary = ""
	}
	if address == id || address == primary {
		address = ""
	}
	parts := []string{id, primary, address}
	return strings.Join(parts, "\t")
}

func resolveMailboxID(state *appState, mailboxID string) string {
	if mailboxID != "" {
		return mailboxID
	}
	if state != nil && state.Config != nil && state.Config.DefaultMailboxID != "" {
		return state.Config.DefaultMailboxID
	}
	// Feishu Mail OpenAPI supports using literal "me" as the mailbox id for the
	// current authenticated user (user access token).
	return "me"
}

var mailFolderAliasLookup = map[string]string{
	"inbox":    "INBOX",
	"收件箱":      "INBOX",
	"sent":     "SENT",
	"已发送":      "SENT",
	"draft":    "DRAFT",
	"drafts":   "DRAFT",
	"草稿":       "DRAFT",
	"草稿箱":      "DRAFT",
	"trash":    "TRASH",
	"deleted":  "TRASH",
	"垃圾箱":      "TRASH",
	"废纸篓":      "TRASH",
	"spam":     "SPAM",
	"junk":     "SPAM",
	"垃圾邮件":     "SPAM",
	"archive":  "ARCHIVED",
	"archived": "ARCHIVED",
	"归档":       "ARCHIVED",
	"已归档":      "ARCHIVED",
}

func mailFolderAliasKey(value string) string {
	trimmed := strings.TrimSpace(value)
	if trimmed == "" {
		return ""
	}
	if canonical, ok := mailFolderAliasLookup[strings.ToLower(trimmed)]; ok {
		return canonical
	}
	return ""
}

func resolveMailFolderAlias(ctx context.Context, state *appState, token, mailboxID, folderID string) (string, error) {
	canonical := mailFolderAliasKey(folderID)
	if canonical == "" {
		return folderID, nil
	}
	if state == nil || state.SDK == nil {
		return canonical, nil
	}
	folders, err := state.SDK.ListMailFolders(ctx, token, mailboxID)
	if err != nil {
		return "", err
	}
	for _, folder := range folders {
		if folder.FolderID != "" && strings.EqualFold(folder.FolderID, folderID) {
			return folder.FolderID, nil
		}
	}
	for _, folder := range folders {
		if matchesMailFolderAlias(folder, canonical, folderID) {
			if folder.FolderID != "" {
				return folder.FolderID, nil
			}
		}
	}
	return canonical, nil
}

func matchesMailFolderAlias(folder larksdk.MailFolder, canonical, rawInput string) bool {
	if strings.EqualFold(folder.FolderType.String(), canonical) {
		return true
	}
	if strings.EqualFold(folder.Name, rawInput) {
		return true
	}
	if mailFolderAliasKey(folder.Name) == canonical {
		return true
	}
	return false
}

func resolveMailFolderID(ctx context.Context, state *appState, token, mailboxID, folderID string) (string, error) {
	if folderID != "" {
		return resolveMailFolderAlias(ctx, state, token, mailboxID, folderID)
	}
	if state == nil || state.SDK == nil {
		return "", errors.New("sdk client is required")
	}
	folders, err := state.SDK.ListMailFolders(ctx, token, mailboxID)
	if err != nil {
		return "", err
	}
	for _, folder := range folders {
		if strings.EqualFold(folder.FolderType.String(), "INBOX") {
			if folder.FolderID != "" {
				return folder.FolderID, nil
			}
			return "INBOX", nil
		}
	}
	for _, folder := range folders {
		if mailFolderAliasKey(folder.Name) == "INBOX" {
			if folder.FolderID != "" {
				return folder.FolderID, nil
			}
			return "INBOX", nil
		}
	}
	if len(folders) > 0 && folders[0].FolderID != "" {
		return folders[0].FolderID, nil
	}
	return "", errors.New("folder id is required; run `lark mail folders` to get IDs, then use `lark mail list --folder-id <id>` (system aliases: INBOX/SENT/DRAFT/TRASH/SPAM/ARCHIVED)")
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
