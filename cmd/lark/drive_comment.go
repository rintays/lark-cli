package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

var supportedDriveCommentFileTypes = map[string]struct{}{
	"doc":    {},
	"docx":   {},
	"sheet":  {},
	"file":   {},
	"slides": {},
}

type driveCommentCommandDefaults struct {
	defaultFileType string
	fileArgName     string
	scopeName       string
}

func newDriveCommentCmd(state *appState) *cobra.Command {
	return newDriveCommentRootCmd(state, driveCommentCommandDefaults{
		defaultFileType: "",
		fileArgName:     "file-token",
		scopeName:       "drive",
	})
}

func newDocsCommentCmd(state *appState) *cobra.Command {
	return newDriveCommentRootCmd(state, driveCommentCommandDefaults{
		defaultFileType: "docx",
		fileArgName:     "document-token",
		scopeName:       "docs",
	})
}

func newSheetsCommentCmd(state *appState) *cobra.Command {
	return newDriveCommentRootCmd(state, driveCommentCommandDefaults{
		defaultFileType: "sheet",
		fileArgName:     "spreadsheet-token",
		scopeName:       "sheets",
	})
}

func newDriveCommentRootCmd(state *appState, defaults driveCommentCommandDefaults) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "comment",
		Short: "Manage anchored comments",
		Long: strings.TrimSpace(`Drive comments are threaded discussions attached to a Drive file.

Anchored comments ("partial comments") use the quote field to locate the target.

Notes about quote:
- quote must be unique in the file to avoid attaching to the wrong location.
- quote must represent a continuous selection (avoid disjoint ranges).
- if the referenced content changes, the anchor may fail to resolve.`),
	}
	annotateAuthServices(cmd, "drive")

	var fileType string
	var userIDType string
	if defaults.defaultFileType == "" {
		cmd.PersistentFlags().StringVar(&fileType, "type", "", "file type (docx|sheet|file|doc|slides); inferred from URL when possible")
	} else {
		fileType = defaults.defaultFileType
		cmd.PersistentFlags().StringVar(&fileType, "type", defaults.defaultFileType, "file type (default: "+defaults.defaultFileType+")")
		_ = cmd.PersistentFlags().MarkHidden("type")
	}
	cmd.PersistentFlags().StringVar(&userIDType, "user-id-type", "open_id", "user id type (open_id|union_id|user_id)")

	cmd.AddCommand(newDriveCommentAddCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentListCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentGetCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentUpdateCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentRepliesCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentReplyCmd(state, defaults, &fileType, &userIDType))
	cmd.AddCommand(newDriveCommentReplyUpdateCmd(state, defaults, &fileType, &userIDType))
	return cmd
}

func newDriveCommentAddCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var quote string
	var contentJSON string
	var text string

	example := "lark " + defaults.scopeName + " comment add <" + defaults.fileArgName + "> --text 'Looks good'\n" +
		"lark " + defaults.scopeName + " comment add <" + defaults.fileArgName + "> --quote '{...}' --content-json '{\"elements\":[{\"type\":\"text_run\",\"text_run\":{\"text\":\"Please check this\"}}]}'"

	cmd := &cobra.Command{
		Use:     "add <" + defaults.fileArgName + ">",
		Short:   "Add a comment (whole or anchored)",
		Args:    cobra.ExactArgs(1),
		Example: example,
		RunE: func(cmd *cobra.Command, args []string) error {
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			content, err := parseDriveCommentReplyContent(text, contentJSON)
			if err != nil {
				return flagUsage(cmd, err.Error())
			}
			quote = strings.TrimSpace(quote)
			comment, err := buildDriveFileCommentCreate(nil, content, quote)
			if err != nil {
				return flagUsage(cmd, err.Error())
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			created, err := state.SDK.CreateDriveFileComment(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.CreateDriveFileCommentRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				UserIDType: *userIDType,
				Comment:    comment,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"comment": created}
			textOut := ""
			if created != nil && created.CommentId != nil {
				textOut = fmt.Sprintf("%s", *created.CommentId)
			}
			return state.Printer.Print(payload, textOut)
		},
	}

	cmd.Flags().StringVar(&quote, "quote", "", "anchor quote string for partial comments (must be unique and continuous in the file)")
	cmd.Flags().StringVar(&text, "text", "", "plain text content (mutually exclusive with --content-json)")
	cmd.Flags().StringVar(&contentJSON, "content-json", "", driveCommentContentJSONHelp())
	return cmd
}

func newDriveCommentListCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "list <" + defaults.fileArgName + ">",
		Short: "List comments",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize <= 0 {
				return flagUsage(cmd, "page-size must be greater than 0")
			}
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			result, err := state.SDK.ListDriveFileComments(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.ListDriveFileCommentsRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				UserIDType: *userIDType,
				PageSize:   pageSize,
				PageToken:  strings.TrimSpace(pageToken),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"items": result.Items, "has_more": result.HasMore, "page_token": result.PageToken}
			lines := make([]string, 0, len(result.Items))
			for _, item := range result.Items {
				id := derefString(item.CommentId)
				solved := fmt.Sprintf("%v", derefBool(item.IsSolved))
				whole := fmt.Sprintf("%v", derefBool(item.IsWhole))
				quote := strings.TrimSpace(derefString(item.Quote))
				if len(quote) > 40 {
					quote = quote[:40] + "..."
				}
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", id, solved, whole, quote))
			}
			text := tableText([]string{"comment_id", "solved", "whole", "quote"}, lines, "no comments found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&pageSize, "page-size", 50, "page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "page token")
	return cmd
}

func newDriveCommentGetCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get <" + defaults.fileArgName + "> <comment-id>",
		Short: "Get a comment thread",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			commentID := strings.TrimSpace(args[1])
			if commentID == "" {
				return argsUsageError(cmd, errors.New("comment-id is required"))
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			comment, err := state.SDK.GetDriveFileComment(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.GetDriveFileCommentRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				CommentID:  commentID,
				UserIDType: *userIDType,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"comment": comment}
			return state.Printer.Print(payload, "")
		},
	}
	return cmd
}

func newDriveCommentUpdateCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var solved bool

	cmd := &cobra.Command{
		Use:   "update <" + defaults.fileArgName + "> <comment-id>",
		Short: "Update a comment (currently: solved status)",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if !cmd.Flags().Changed("solved") {
				return flagUsage(cmd, "--solved is required")
			}
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			commentID := strings.TrimSpace(args[1])
			if commentID == "" {
				return argsUsageError(cmd, errors.New("comment-id is required"))
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			err = state.SDK.PatchDriveFileComment(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.PatchDriveFileCommentRequest{
				FileToken: fileToken,
				CommentID: commentID,
				FileType:  resolvedType,
				IsSolved:  solved,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"ok": true}
			return state.Printer.Print(payload, "ok")
		},
	}

	cmd.Flags().BoolVar(&solved, "solved", false, "mark comment as solved (true) or unsolved (false)")
	return cmd
}

func newDriveCommentRepliesCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var pageSize int
	var pageToken string

	cmd := &cobra.Command{
		Use:   "replies <" + defaults.fileArgName + "> <comment-id>",
		Short: "List replies in a comment thread",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if pageSize <= 0 {
				return flagUsage(cmd, "page-size must be greater than 0")
			}
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			commentID := strings.TrimSpace(args[1])
			if commentID == "" {
				return argsUsageError(cmd, errors.New("comment-id is required"))
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			result, err := state.SDK.ListDriveFileCommentReplies(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.ListDriveFileCommentRepliesRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				CommentID:  commentID,
				UserIDType: *userIDType,
				PageSize:   pageSize,
				PageToken:  strings.TrimSpace(pageToken),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"items": result.Items, "has_more": result.HasMore, "page_token": result.PageToken}
			lines := make([]string, 0, len(result.Items))
			for _, item := range result.Items {
				replyID := derefString(item.ReplyId)
				userID := derefString(item.UserId)
				text := summarizeReplyContent(item.Content)
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", replyID, userID, text))
			}
			textOut := tableText([]string{"reply_id", "user_id", "text"}, lines, "no replies found")
			return state.Printer.Print(payload, textOut)
		},
	}

	cmd.Flags().IntVar(&pageSize, "page-size", 50, "page size")
	cmd.Flags().StringVar(&pageToken, "page-token", "", "page token")
	return cmd
}

func newDriveCommentReplyCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var contentJSON string
	var text string

	cmd := &cobra.Command{
		Use:   "reply <" + defaults.fileArgName + "> <comment-id>",
		Short: "Add a reply to a comment thread",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			commentID := strings.TrimSpace(args[1])
			if commentID == "" {
				return argsUsageError(cmd, errors.New("comment-id is required"))
			}
			content, err := parseDriveCommentReplyContent(text, contentJSON)
			if err != nil {
				return flagUsage(cmd, err.Error())
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			// Create reply
			if _, err := state.SDK.CreateDriveFileCommentReply(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.CreateDriveFileCommentReplyRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				CommentID:  commentID,
				UserIDType: *userIDType,
				Content:    content,
			}); err != nil {
				return err
			}

			// Best-effort: fetch replies and return the latest reply_id.
			latest, err := fetchLatestDriveCommentReply(ctx, state, token, larksdk.AccessTokenType(tokenTypeValue), fileToken, resolvedType, commentID, *userIDType)
			if err != nil {
				return err
			}
			replyID := ""
			if latest != nil && latest.ReplyId != nil {
				replyID = *latest.ReplyId
			}
			payload := map[string]any{"ok": true, "reply_id": replyID, "reply": latest}
			textOut := "ok"
			if replyID != "" {
				textOut = replyID
			}
			return state.Printer.Print(payload, textOut)
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "plain text content (mutually exclusive with --content-json)")
	cmd.Flags().StringVar(&contentJSON, "content-json", "", driveCommentContentJSONHelp())
	return cmd
}

func newDriveCommentReplyUpdateCmd(state *appState, defaults driveCommentCommandDefaults, fileType *string, userIDType *string) *cobra.Command {
	var contentJSON string
	var text string

	cmd := &cobra.Command{
		Use:   "reply-update <" + defaults.fileArgName + "> <comment-id> <reply-id>",
		Short: "Update a reply",
		Args:  cobra.ExactArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			fileToken, resolvedType, err := resolveDriveCommentFileRef(args[0], *fileType, defaults.defaultFileType)
			if err != nil {
				return argsUsageError(cmd, err)
			}
			commentID := strings.TrimSpace(args[1])
			if commentID == "" {
				return argsUsageError(cmd, errors.New("comment-id is required"))
			}
			replyID := strings.TrimSpace(args[2])
			if replyID == "" {
				return argsUsageError(cmd, errors.New("reply-id is required"))
			}
			content, err := parseDriveCommentReplyContent(text, contentJSON)
			if err != nil {
				return flagUsage(cmd, err.Error())
			}
			ctx := cmd.Context()
			token, tokenTypeValue, err := resolveAccessToken(ctx, state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			err = state.SDK.UpdateDriveFileCommentReply(ctx, token, larksdk.AccessTokenType(tokenTypeValue), larksdk.UpdateDriveFileCommentReplyRequest{
				FileToken:  fileToken,
				FileType:   resolvedType,
				CommentID:  commentID,
				ReplyID:    replyID,
				UserIDType: *userIDType,
				Content:    content,
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"ok": true}
			return state.Printer.Print(payload, "ok")
		},
	}

	cmd.Flags().StringVar(&text, "text", "", "plain text content (mutually exclusive with --content-json)")
	cmd.Flags().StringVar(&contentJSON, "content-json", "", driveCommentContentJSONHelp())
	return cmd
}

func driveCommentContentJSONHelp() string {
	return strings.TrimSpace(`reply content JSON (ReplyContent). Example:
  {"elements":[{"type":"text_run","text_run":{"text":"hello"}}]}

Each element supports:
- {"type":"text_run","text_run":{"text":"..."}}
- {"type":"person","person":{"user_id":"<open_id|union_id|user_id>"}}
- {"type":"docs_link","docs_link":{"url":"https://..."}}`)
}

func resolveDriveCommentFileRef(raw string, fileType string, defaultFileType string) (string, string, error) {
	token, kind, err := parseResourceRef(raw)
	if err != nil {
		return "", "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", "", errors.New("file token is required")
	}
	resolvedType := strings.TrimSpace(fileType)
	if resolvedType == "" {
		resolvedType = strings.TrimSpace(defaultFileType)
	}
	if resolvedType == "" {
		resolvedType = resourceKindToDriveCommentFileType(kind)
	}
	resolvedType = strings.TrimSpace(resolvedType)
	if resolvedType == "" {
		return "", "", errors.New("file type is required (use --type)")
	}
	if _, ok := supportedDriveCommentFileTypes[resolvedType]; !ok {
		return "", "", fmt.Errorf("unsupported file type: %s", resolvedType)
	}
	return token, resolvedType, nil
}

func resourceKindToDriveCommentFileType(kind string) string {
	switch strings.ToLower(strings.TrimSpace(kind)) {
	case "docx":
		return "docx"
	case "sheet":
		return "sheet"
	case "slides":
		return "slides"
	case "file":
		return "file"
	case "doc":
		return "doc"
	default:
		return ""
	}
}

func parseDriveCommentReplyContent(text string, contentJSON string) (*larkdrive.ReplyContent, error) {
	text = strings.TrimSpace(text)
	contentJSON = strings.TrimSpace(contentJSON)
	if text != "" && contentJSON != "" {
		return nil, errors.New("--text and --content-json are mutually exclusive")
	}
	if text != "" {
		typeTextRun := "text_run"
		return &larkdrive.ReplyContent{Elements: []*larkdrive.ReplyElement{{
			Type:    &typeTextRun,
			TextRun: &larkdrive.TextRun{Text: &text},
		}}}, nil
	}
	if contentJSON == "" {
		return nil, errors.New("either --text or --content-json is required")
	}
	var content larkdrive.ReplyContent
	if err := json.Unmarshal([]byte(contentJSON), &content); err != nil {
		return nil, fmt.Errorf("invalid --content-json: %w", err)
	}
	if len(content.Elements) == 0 {
		return nil, errors.New("--content-json must include at least one element")
	}
	return &content, nil
}

func buildDriveFileCommentCreate(parentCommentID *string, content *larkdrive.ReplyContent, quote string) (*larkdrive.FileComment, error) {
	if content == nil {
		return nil, errors.New("content is required")
	}
	quote = strings.TrimSpace(quote)
	if parentCommentID != nil {
		id := strings.TrimSpace(*parentCommentID)
		if id == "" {
			return nil, errors.New("comment-id is required")
		}
		return &larkdrive.FileComment{
			CommentId: &id,
			ReplyList: &larkdrive.ReplyList{Replies: []*larkdrive.FileCommentReply{{Content: content}}},
		}, nil
	}

	if quote == "" {
		whole := true
		return &larkdrive.FileComment{
			IsWhole:   &whole,
			ReplyList: &larkdrive.ReplyList{Replies: []*larkdrive.FileCommentReply{{Content: content}}},
		}, nil
	}
	whole := false
	quoteValue := quote
	return &larkdrive.FileComment{
		IsWhole:   &whole,
		Quote:     &quoteValue,
		ReplyList: &larkdrive.ReplyList{Replies: []*larkdrive.FileCommentReply{{Content: content}}},
	}, nil
}

func summarizeReplyContent(content *larkdrive.ReplyContent) string {
	if content == nil || len(content.Elements) == 0 {
		return ""
	}
	var parts []string
	for _, el := range content.Elements {
		if el == nil {
			continue
		}
		switch strings.ToLower(derefString(el.Type)) {
		case "text_run":
			if el.TextRun != nil {
				parts = append(parts, derefString(el.TextRun.Text))
			}
		case "docs_link":
			if el.DocsLink != nil {
				parts = append(parts, derefString(el.DocsLink.Url))
			}
		case "person":
			if el.Person != nil {
				parts = append(parts, "@"+derefString(el.Person.UserId))
			}
		}
	}
	joined := strings.Join(parts, "")
	joined = strings.TrimSpace(joined)
	if len(joined) > 60 {
		return joined[:60] + "..."
	}
	return joined
}

func derefString(ptr *string) string {
	if ptr == nil {
		return ""
	}
	return *ptr
}

func derefBool(ptr *bool) bool {
	if ptr == nil {
		return false
	}
	return *ptr
}

func fetchLatestDriveCommentReply(ctx context.Context, state *appState, token string, tokenType larksdk.AccessTokenType, fileToken string, fileType string, commentID string, userIDType string) (*larkdrive.FileCommentReply, error) {
	pageToken := ""
	pageSize := 50
	var latest *larkdrive.FileCommentReply
	for {
		result, err := state.SDK.ListDriveFileCommentReplies(ctx, token, tokenType, larksdk.ListDriveFileCommentRepliesRequest{
			FileToken:  fileToken,
			FileType:   fileType,
			CommentID:  commentID,
			UserIDType: userIDType,
			PageSize:   pageSize,
			PageToken:  pageToken,
		})
		if err != nil {
			return nil, err
		}
		latest = findLatestDriveCommentReply(latest, result.Items)
		if !result.HasMore {
			break
		}
		if strings.TrimSpace(result.PageToken) == "" {
			break
		}
		pageToken = result.PageToken
	}
	return latest, nil
}

func findLatestDriveCommentReply(current *larkdrive.FileCommentReply, candidates []*larkdrive.FileCommentReply) *larkdrive.FileCommentReply {
	best := current
	bestTime := driveCommentReplyTime(current)
	for _, r := range candidates {
		t := driveCommentReplyTime(r)
		if t >= bestTime {
			bestTime = t
			best = r
		}
	}
	return best
}

func driveCommentReplyTime(r *larkdrive.FileCommentReply) int {
	if r == nil {
		return -1
	}
	if r.CreateTime != nil {
		return *r.CreateTime
	}
	if r.UpdateTime != nil {
		return *r.UpdateTime
	}
	return -1
}
