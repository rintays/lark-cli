package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newDocsListCmd(state *appState) *cobra.Command {
	var folderID string
	var limit int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Docs (docx) in a Drive folder",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
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

				// Thin wrapper over drive list: fetch then filter to docx.
				result, err := state.SDK.ListDriveFiles(context.Background(), token, larksdk.ListDriveFilesRequest{
					FolderToken: folderID,
					PageSize:    pageSize,
					PageToken:   pageToken,
				})
				if err != nil {
					return err
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
			text := tableText([]string{"token", "name", "type", "url"}, lines, "no files found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of files to return")

	return cmd
}
