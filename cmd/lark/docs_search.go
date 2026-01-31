package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newDocsSearchCmd(state *appState) *cobra.Command {
	var query string
	var limit int
	var pages int

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Docs (docx) by text",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(query) == "" {
					return errors.New("query is required")
				}
				return nil
			}
			if query != "" && query != args[0] {
				return errors.New("query provided twice")
			}
			return cmd.Flags().Set("query", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pages <= 0 {
				return errors.New("pages must be greater than 0")
			}
			ctx := context.Background()
			token, err := resolveDriveSearchToken(ctx, state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			files, err := docsSearchDriveFiles(ctx, state, token, "docs", query, []string{"doc"}, limit, pages)
			if err != nil {
				return err
			}
			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := tableText([]string{"token", "name", "type", "url"}, lines, "no files found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search text (or provide as positional argument)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")

	return cmd
}
