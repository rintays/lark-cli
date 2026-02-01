package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const exportTaskMaxAttempts = 20

var exportTaskPollInterval = 200 * time.Millisecond

func newDocsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage Docs (docx) documents",
		Long: `Docs (docx) are document files stored in Drive.

- document_id is the docx file token (use it as FILE_TOKEN for drive permissions).
- A doc contains blocks (paragraphs, headings, lists, tables, images) that make up its structure and content.
- Documents can live in a Drive folder (folder-id).
- Docx is the default API surface; legacy docs scopes are deprecated.
- Use lark drive permissions to manage collaborators for docs.
- Use info/export/get to inspect or download content.`,
	}
	annotateAuthServices(cmd, "docs")
	cmd.AddCommand(newDocsListCmd(state))
	cmd.AddCommand(newDocsCreateCmd(state))
	cmd.AddCommand(newDocsInfoCmd(state))
	cmd.AddCommand(newDocsExportCmd(state))
	cmd.AddCommand(newDocsGetCmd(state))
	cmd.AddCommand(newDocsSearchCmd(state))
	cmd.AddCommand(newDocsBlocksCmd(state))
	cmd.AddCommand(newDocsConvertCmd(state))
	cmd.AddCommand(newDocsOverwriteCmd(state))
	return cmd
}

func newDocsCreateCmd(state *appState) *cobra.Command {
	var folderID string

	cmd := &cobra.Command{
		Use:   "create <title>",
		Short: "Create a Docs (docx) document",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) == 0 {
				return errors.New("title is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			title := args[0]
			token, tokenType, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			doc, err := state.SDK.CreateDocxDocument(cmd.Context(), token, larksdk.CreateDocxDocumentRequest{
				Title:       title,
				FolderToken: folderID,
			})
			if err != nil {
				return err
			}
			if doc.URL == "" && doc.DocumentID != "" {
				doc.URL = docxDriveURL(cmd.Context(), state, tokenType, token, doc.DocumentID)
			}
			payload := map[string]any{"document": doc}
			text := tableTextRow(
				[]string{"document_id", "title", "url"},
				[]string{doc.DocumentID, doc.Title, doc.URL},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	return cmd
}

func newDocsInfoCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <document-id>",
		Short: "Show Docs (docx) document info",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			if strings.TrimSpace(token) == "" {
				return errors.New("document-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if _, err := requireSDK(state); err != nil {
				return err
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			documentID := strings.TrimSpace(token)
			token, tokenType, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			doc, err := state.SDK.GetDocxDocument(cmd.Context(), token, documentID)
			if err != nil {
				return err
			}
			if doc.URL == "" {
				doc.URL = docxDriveURL(cmd.Context(), state, tokenType, token, documentID)
			}
			payload := map[string]any{"document": doc}
			text := formatDocxInfo(doc)
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newDocsExportCmd(state *appState) *cobra.Command {
	var format string
	var outPath string

	cmd := &cobra.Command{
		Use:   "export <document-id> --format pdf --out <path>",
		Short: "Export a Docs (docx) document",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			if strings.TrimSpace(token) == "" {
				return errors.New("document-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			outPath = strings.TrimSpace(outPath)
			writeStdout := outPath == "-"
			if !writeStdout {
				if info, err := os.Stat(outPath); err == nil && info.IsDir() {
					return fmt.Errorf("output path is a directory: %s", outPath)
				}
			}
			format = strings.ToLower(format)
			if _, err := requireSDK(state); err != nil {
				return err
			}
			refToken, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			documentID := strings.TrimSpace(refToken)
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			ticket, err := state.SDK.CreateExportTask(cmd.Context(), token, larksdk.CreateExportTaskRequest{
				Token:         documentID,
				Type:          "docx",
				FileExtension: format,
			})
			if err != nil {
				return err
			}
			result, err := pollExportTask(cmd.Context(), state.SDK, token, ticket)
			if err != nil {
				return err
			}
			reader, err := state.SDK.DownloadExportedFile(cmd.Context(), token, result.FileToken)
			if err != nil {
				return err
			}
			defer reader.Close()
			var out io.Writer
			var outFile *os.File
			if writeStdout {
				out = cmd.OutOrStdout()
			} else {
				outFile, err = os.Create(outPath)
				if err != nil {
					return err
				}
				defer outFile.Close()
				out = outFile
			}
			written, err := io.Copy(out, reader)
			if err != nil {
				return err
			}
			if writeStdout {
				if state.Verbose {
					fmt.Fprintf(errWriter(state), "wrote %d bytes to stdout\n", written)
				}
				return nil
			}
			payload := map[string]any{
				"document_id":   documentID,
				"format":        format,
				"file_token":    result.FileToken,
				"output_path":   outPath,
				"bytes_written": written,
			}
			text := tableTextRow(
				[]string{"document_id", "output_path", "bytes_written"},
				[]string{documentID, outPath, fmt.Sprintf("%d", written)},
			)
			return state.Printer.Print(payload, text)
		},
	}
	annotateAuthServices(cmd, "drive-export")

	cmd.Flags().StringVar(&format, "format", "", "export format (pdf)")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path (or - for stdout)")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDocsGetCmd(state *appState) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "get <document-id> [--format md|txt|blocks]",
		Short: "Fetch Docs (docx) document content",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			if strings.TrimSpace(token) == "" {
				return errors.New("document-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			accessToken, err := tokenFor(cmd.Context(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			format = strings.ToLower(strings.TrimSpace(format))
			if format == "" {
				format = "md"
			}
			refToken, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			documentID := strings.TrimSpace(refToken)
			switch format {
			case "md", "markdown":
				blocks, err := listDocxBlocks(cmd.Context(), state.SDK, accessToken, documentID)
				if err != nil {
					return err
				}
				content := docxBlocksMarkdown(documentID, blocks)
				if state.JSON {
					payload := map[string]any{
						"document_id": documentID,
						"format":      "md",
						"content":     content,
					}
					return state.Printer.Print(payload, "")
				}
				_, err = io.WriteString(state.Printer.Writer, content)
				return err
			case "txt", "text":
				content, err := state.SDK.GetDocxRawContent(cmd.Context(), accessToken, documentID)
				if err != nil {
					return err
				}
				content = normalizeDocxContentEscapes(content)
				if content != "" {
					doc, err := state.SDK.GetDocxDocument(cmd.Context(), accessToken, documentID)
					if err == nil && doc.Title != "" {
						content = stripDocxTitlePrefix(content, doc.Title)
					}
				}
				if state.JSON {
					payload := map[string]any{
						"document_id": documentID,
						"format":      "txt",
						"content":     content,
					}
					return state.Printer.Print(payload, "")
				}
				_, err = io.WriteString(state.Printer.Writer, content)
				return err
			case "blocks":
				blocks, err := listDocxBlocks(cmd.Context(), state.SDK, accessToken, documentID)
				if err != nil {
					return err
				}
				payload := map[string]any{
					"document_id": documentID,
					"format":      "blocks",
					"blocks":      blocks,
				}
				text := docxBlocksTable(blocks, "no blocks found")
				return state.Printer.Print(payload, text)
			default:
				return fmt.Errorf("format must be md, txt, or blocks")
			}
		},
	}

	cmd.Flags().StringVar(&format, "format", "md", "output format (md, txt, or blocks)")
	return cmd
}

func listDocxBlocks(ctx context.Context, sdk *larksdk.Client, token, documentID string) ([]*larkdocx.Block, error) {
	if sdk == nil {
		return nil, errors.New("sdk client is required")
	}
	blocks := make([]*larkdocx.Block, 0)
	pageToken := ""
	for {
		items, nextToken, hasMore, err := sdk.ListDocxBlocks(
			ctx,
			token,
			documentID,
			docxBlocksMaxPageSize,
			pageToken,
			-1,
			"",
		)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, items...)
		if !hasMore || nextToken == "" {
			break
		}
		pageToken = nextToken
	}
	return blocks, nil
}

func formatDocxInfo(doc larksdk.DocxDocument) string {
	rows := [][]string{
		{"document_id", infoValue(doc.DocumentID)},
		{"title", infoValue(doc.Title)},
		{"url", infoValue(doc.URL)},
		{"revision_id", infoValue(string(doc.RevisionID))},
	}
	setting := doc.DisplaySetting
	rows = append(rows,
		[]string{"display_setting.show_authors", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_create_time", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_pv", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_uv", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_like_count", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_comment_count", infoValueBoolPtr(nil)},
		[]string{"display_setting.show_related_matters", infoValueBoolPtr(nil)},
	)
	if setting != nil {
		rows[4][1] = infoValueBoolPtr(setting.ShowAuthors)
		rows[5][1] = infoValueBoolPtr(setting.ShowCreateTime)
		rows[6][1] = infoValueBoolPtr(setting.ShowPv)
		rows[7][1] = infoValueBoolPtr(setting.ShowUv)
		rows[8][1] = infoValueBoolPtr(setting.ShowLikeCount)
		rows[9][1] = infoValueBoolPtr(setting.ShowCommentCount)
		rows[10][1] = infoValueBoolPtr(setting.ShowRelatedMatters)
	}
	cover := doc.Cover
	rows = append(rows,
		[]string{"cover.token", infoValue("")},
		[]string{"cover.offset_ratio_x", infoValueFloatPtr(nil)},
		[]string{"cover.offset_ratio_y", infoValueFloatPtr(nil)},
	)
	if cover != nil {
		rows[len(rows)-3][1] = infoValue(cover.Token)
		rows[len(rows)-2][1] = infoValueFloatPtr(cover.OffsetRatioX)
		rows[len(rows)-1][1] = infoValueFloatPtr(cover.OffsetRatioY)
	}
	return formatInfoTable(rows, "no document found")
}

func docxDriveURL(ctx context.Context, state *appState, accessType tokenType, token, documentID string) string {
	if state == nil || state.SDK == nil || documentID == "" {
		return ""
	}
	file, err := driveFileMetadataWithToken(ctx, state.SDK, accessType, token, documentID)
	if err == nil && file.URL != "" {
		return file.URL
	}
	if accessType == tokenTypeUser {
		return ""
	}
	userToken, err := resolveDriveSearchToken(ctx, state)
	if err != nil || userToken == "" {
		return ""
	}
	if state.Verbose {
		fmt.Fprintln(errWriter(state), "retrying docs url lookup with user access token")
	}
	fallbackFile, err := driveFileMetadataWithToken(ctx, state.SDK, tokenTypeUser, userToken, documentID)
	if err != nil || fallbackFile.URL == "" {
		return ""
	}
	return fallbackFile.URL
}

func stripDocxTitlePrefix(content, title string) string {
	title = strings.TrimSpace(title)
	if title == "" || content == "" {
		return content
	}
	if content == title {
		return ""
	}
	if strings.HasPrefix(content, title+"\r\n") {
		return strings.TrimPrefix(content, title+"\r\n")
	}
	if strings.HasPrefix(content, title+"\n") {
		return strings.TrimPrefix(content, title+"\n")
	}
	return content
}

type exportTaskClient interface {
	GetExportTask(ctx context.Context, token, ticket string) (larksdk.ExportTaskResult, error)
}

func pollExportTask(ctx context.Context, client exportTaskClient, token, ticket string) (larksdk.ExportTaskResult, error) {
	var lastResult larksdk.ExportTaskResult
	for attempt := 0; attempt < exportTaskMaxAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			return larksdk.ExportTaskResult{}, err
		}
		result, err := client.GetExportTask(ctx, token, ticket)
		if err != nil {
			return larksdk.ExportTaskResult{}, err
		}
		lastResult = result
		switch result.JobStatus {
		case 0:
			if result.FileToken == "" {
				return larksdk.ExportTaskResult{}, errors.New("export task completed without file token")
			}
			return result, nil
		case 1:
		default:
			if result.JobErrorMsg != "" {
				return larksdk.ExportTaskResult{}, fmt.Errorf("export task failed: %s", result.JobErrorMsg)
			}
			return larksdk.ExportTaskResult{}, fmt.Errorf("export task failed with status %d", result.JobStatus)
		}
		if exportTaskPollInterval > 0 {
			select {
			case <-ctx.Done():
				return larksdk.ExportTaskResult{}, ctx.Err()
			case <-time.After(exportTaskPollInterval):
			}
		}
	}
	if lastResult.JobErrorMsg != "" {
		return larksdk.ExportTaskResult{}, fmt.Errorf("export task not ready: %s", lastResult.JobErrorMsg)
	}
	return larksdk.ExportTaskResult{}, fmt.Errorf("export task not ready after %d attempts", exportTaskMaxAttempts)
}
