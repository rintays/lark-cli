package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
)

func newDocsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "docs",
		Short: "Manage Docs (docx) documents",
	}
	cmd.AddCommand(newDocsCreateCmd(state))
	cmd.AddCommand(newDocsGetCmd(state))
	return cmd
}

func newDocsCreateCmd(state *appState) *cobra.Command {
	var title string
	var folderID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Docs (docx) document",
		RunE: func(cmd *cobra.Command, args []string) error {
			if title == "" {
				return errors.New("title is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			doc, err := state.Client.CreateDocxDocument(context.Background(), token, larkapi.CreateDocxDocumentRequest{
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
	return cmd
}

func newDocsGetCmd(state *appState) *cobra.Command {
	var documentID string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get Docs (docx) document metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if documentID == "" {
				return errors.New("doc-id is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			doc, err := state.Client.GetDocxDocument(context.Background(), token, documentID)
			if err != nil {
				return err
			}
			payload := map[string]any{"document": doc}
			text := fmt.Sprintf("%s\t%s\t%s", doc.DocumentID, doc.Title, doc.URL)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&documentID, "doc-id", "", "document ID")
	return cmd
}
