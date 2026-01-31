package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

const exportTaskMaxAttempts = 20

var exportTaskPollInterval = 200 * time.Millisecond

func newDocsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage Docs (docx) documents",
	}
	cmd.AddCommand(newDocsCreateCmd(state))
	cmd.AddCommand(newDocsGetCmd(state))
	cmd.AddCommand(newDocsExportCmd(state))
	cmd.AddCommand(newDocsCatCmd(state))
	cmd.AddCommand(newDocsSearchCmd(state))
	return cmd
}

func newDocsCreateCmd(state *appState) *cobra.Command {
	var title string
	var folderID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Docs (docx) document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
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
			payload := map[string]any{"document": doc}
			text := fmt.Sprintf("%s\t%s\t%s", doc.DocumentID, doc.Title, doc.URL)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "document title")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}

func newDocsGetCmd(state *appState) *cobra.Command {
	var documentID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Docs (docx) document metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			doc, err := state.SDK.GetDocxDocument(context.Background(), token, documentID)
			if err != nil {
				return err
			}
			payload := map[string]any{"document": doc}
			text := fmt.Sprintf("%s\t%s\t%s", doc.DocumentID, doc.Title, doc.URL)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID")
	_ = cmd.MarkFlagRequired("doc-id")
	return cmd
}

func newDocsExportCmd(state *appState) *cobra.Command {
	var documentID string
	var format string
	var outPath string

	cmd := &cobra.Command{
		Use:   "export --doc-id <DOCUMENT_ID> --format pdf --out <path>",
		Short: "Export a Docs (docx) document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if info, err := os.Stat(outPath); err == nil && info.IsDir() {
				return fmt.Errorf("output path is a directory: %s", outPath)
			}
			format = strings.ToLower(format)
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
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
			text := fmt.Sprintf("%s\t%s\t%d", documentID, outPath, written)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID")
	cmd.Flags().StringVar(&format, "format", "", "export format (pdf)")
	cmd.Flags().StringVar(&outPath, "out", "", "output file path")
	_ = cmd.MarkFlagRequired("doc-id")
	_ = cmd.MarkFlagRequired("format")
	_ = cmd.MarkFlagRequired("out")
	return cmd
}

func newDocsCatCmd(state *appState) *cobra.Command {
	var documentID string
	var format string

	cmd := &cobra.Command{
		Use:   "cat --doc-id <DOCUMENT_ID> [--format txt|md]",
		Short: "Print Docs (docx) document content",
		RunE: func(cmd *cobra.Command, args []string) error {
			format = strings.ToLower(strings.TrimSpace(format))
			if format == "" {
				format = "txt"
			}
			switch format {
			case "txt", "md":
			default:
				return fmt.Errorf("format must be txt or md")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
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
			if state.JSON {
				content, err := io.ReadAll(reader)
				if err != nil {
					return err
				}
				payload := map[string]any{
					"document_id": documentID,
					"format":      format,
					"file_token":  result.FileToken,
					"file_name":   result.FileName,
					"content":     string(content),
				}
				return state.Printer.Print(payload, "")
			}
			_, err = io.Copy(state.Printer.Writer, reader)
			return err
		},
	}

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID")
	cmd.Flags().StringVar(&format, "format", "txt", "output format (txt or md)")
	_ = cmd.MarkFlagRequired("doc-id")
	return cmd
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
