package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newDocsSearchCmd(state *appState) *cobra.Command {
	var query string
	var limit int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Docs (docx) by text",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			ctx := context.Background()
			token, err := resolveDriveSearchToken(ctx, state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}

			files := make([]larksdk.DriveFile, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > maxDrivePageSize {
					pageSize = maxDrivePageSize
				}

				// Thin wrapper over drive search: fetch then filter to docx.
				result, err := state.SDK.SearchDriveFilesWithUserToken(ctx, token, larksdk.SearchDriveFilesRequest{
					Query:     query,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
				}
				for _, file := range result.Files {
					if file.FileType == "docx" {
						files = append(files, file)
						if len(files) >= limit {
							break
						}
					}
				}
				if len(files) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(files)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(files) > limit {
				files = files[:limit]
			}
			payload := map[string]any{"files": files}
			lines := make([]string, 0, len(files))
			for _, file := range files {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", file.Token, file.Name, file.FileType, file.URL))
			}
			text := "no files found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search text")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")
	_ = cmd.MarkFlagRequired("query")

	return cmd
}
