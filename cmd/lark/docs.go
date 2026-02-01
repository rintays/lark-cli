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

- document_id is the docx file token.
- A doc contains blocks (paragraphs, headings, lists, tables, images) that make up its structure and content.
- Documents can live in a Drive folder (folder-id).
- Use info/export/get to inspect or download content.`,
	}
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
				return err
			}
			if len(args) == 0 {
				return errors.New("title is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			title := args[0]
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			doc, err := state.SDK.CreateDocxDocument(context.Background(), token, larksdk.CreateDocxDocumentRequest{
				Title:       title,
				FolderToken: folderID,
			})
			if err != nil {
				return err
			}
			if doc.URL == "" && doc.DocumentID != "" {
				if fetched, err := state.SDK.GetDocxDocument(context.Background(), token, doc.DocumentID); err == nil && fetched.URL != "" {
					doc.URL = fetched.URL
				}
				if doc.URL == "" {
					file, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{
						FileToken: doc.DocumentID,
					})
					if err == nil && file.URL != "" {
						doc.URL = file.URL
					}
				}
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
	var documentID string

	cmd := &cobra.Command{
		Use:   "info <doc-id>",
		Short: "Show Docs (docx) document info",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(documentID) == "" {
					return errors.New("doc-id is required")
				}
				return nil
			}
			if documentID != "" && documentID != args[0] {
				return errors.New("doc-id provided twice")
			}
			return cmd.Flags().Set("doc-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			doc, err := state.SDK.GetDocxDocument(context.Background(), token, documentID)
			if err != nil {
				return err
			}
			if doc.URL == "" {
				file, err := state.SDK.GetDriveFileMetadata(context.Background(), token, larksdk.GetDriveFileRequest{
					FileToken: documentID,
				})
				if err == nil && file.URL != "" {
					doc.URL = file.URL
				}
			}
			payload := map[string]any{"document": doc}
			text := formatDocxInfo(doc)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID (or provide as positional argument)")
	return cmd
}

func newDocsExportCmd(state *appState) *cobra.Command {
	var documentID string
	var format string
	var outPath string

	cmd := &cobra.Command{
		Use:   "export <doc-id> --format pdf --out <path>",
		Short: "Export a Docs (docx) document",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(documentID) == "" {
					return errors.New("doc-id is required")
				}
				return nil
			}
			if documentID != "" && documentID != args[0] {
				return errors.New("doc-id provided twice")
			}
			return cmd.Flags().Set("doc-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if info, err := os.Stat(outPath); err == nil && info.IsDir() {
				return fmt.Errorf("output path is a directory: %s", outPath)
			}
			format = strings.ToLower(format)
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			ticket, err := state.SDK.CreateExportTask(context.Background(), token, larksdk.CreateExportTaskRequest{
				Token:         documentID,
				Type:          "docx",
				FileExtension: format,
			})
			if err != nil {
				return err
			}
			result, err := pollExportTask(context.Background(), state.SDK, token, ticket)
			if err != nil {
				return err
			}
			reader, err := state.SDK.DownloadExportedFile(context.Background(), token, result.FileToken)
			if err != nil {
				return err
			}
			defer reader.Close()
			outFile, err := os.Create(outPath)
			if err != nil {
				return err
			}
			defer outFile.Close()
			written, err := io.Copy(outFile, reader)
			if err != nil {
				return err
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

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID (or provide as positional argument)")
	cmd.Flags().StringVar(&format, "format", "", "export format (pdf)")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDocsGetCmd(state *appState) *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "get <doc-id> [--format md|txt|blocks]",
		Short: "Fetch Docs (docx) document content",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				return errors.New("doc-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			format = strings.ToLower(strings.TrimSpace(format))
			if format == "" {
				format = "md"
			}
			documentID := args[0]
			switch format {
			case "md", "markdown", "txt", "text":
				if format == "markdown" {
					format = "md"
				}
				if format == "text" {
					format = "txt"
				}
				content, err := state.SDK.GetDocxRawContent(context.Background(), token, documentID)
				if err != nil {
					return err
				}
				if state.JSON {
					payload := map[string]any{
						"document_id": documentID,
						"format":      format,
						"content":     content,
					}
					return state.Printer.Print(payload, "")
				}
				_, err = io.WriteString(state.Printer.Writer, content)
				return err
			case "blocks":
				blocks := make([]*larkdocx.Block, 0)
				pageToken := ""
				for {
					items, nextToken, hasMore, err := state.SDK.ListDocxBlocks(
						context.Background(),
						token,
						documentID,
						docxBlocksMaxPageSize,
						pageToken,
						-1,
						"",
					)
					if err != nil {
						return err
					}
					blocks = append(blocks, items...)
					if !hasMore || nextToken == "" {
						break
					}
					pageToken = nextToken
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

type exportTaskClient interface {
	GetExportTask(ctx context.Context, token, ticket string) (larksdk.ExportTaskResult, error)
}

func pollExportTask(ctx context.Context, client exportTaskClient, token, ticket string) (larksdk.ExportTaskResult, error) {
	var lastResult larksdk.ExportTaskResult
	for attempt := 0; attempt < exportTaskMaxAttempts; attempt++ {
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
			time.Sleep(exportTaskPollInterval)
		}
	}
	if lastResult.JobErrorMsg != "" {
		return larksdk.ExportTaskResult{}, fmt.Errorf("export task not ready: %s", lastResult.JobErrorMsg)
	}
	return larksdk.ExportTaskResult{}, fmt.Errorf("export task not ready after %d attempts", exportTaskMaxAttempts)
}
