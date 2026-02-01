package main

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

type baseListItem struct {
	AppToken  string `json:"app_token"`
	Title     string `json:"title"`
	FileToken string `json:"file_token"`
	URL       string `json:"url"`
}

var (
	bitableAppTokenPattern  = regexp.MustCompile(`(?i)(basc[a-z0-9]+)`) // basc... tokens are canonical.
	bitableLegacyAppPattern = regexp.MustCompile(`(?i)(app[a-z0-9]+)`)  // legacy fallback.
)

func extractBitableAppTokenFromURL(rawURL string) string {
	rawURL = strings.TrimSpace(rawURL)
	if rawURL == "" {
		return ""
	}
	if m := bitableAppTokenPattern.FindStringSubmatch(rawURL); len(m) == 2 {
		return m[1]
	}
	if m := bitableLegacyAppPattern.FindStringSubmatch(rawURL); len(m) == 2 {
		return m[1]
	}
	return ""
}

func newBaseListCmd(state *appState) *cobra.Command {
	return newBaseDiscoverListCmd(state, "list", "List Bitable bases (apps) via Drive search")
}

func newBaseAppListCmd(state *appState) *cobra.Command {
	return newBaseDiscoverListCmd(state, "list", "List Bitable apps via Drive search")
}

func newBaseDiscoverListCmd(state *appState, use, short string) *cobra.Command {
	var query string
	var limit int
	var pages int

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
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
			if state == nil {
				return errors.New("state is required")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pages <= 0 {
				return errors.New("pages must be greater than 0")
			}
			ctx := context.Background()
			userToken, err := resolveDriveSearchToken(ctx, state)
			if err != nil {
				return err
			}

			files, err := searchBitableFiles(ctx, state, userToken, query, limit, pages)
			if err != nil {
				return err
			}

			items := make([]baseListItem, 0, len(files))
			lines := make([]string, 0, len(files))
			for _, file := range files {
				appToken := extractBitableAppTokenFromURL(file.URL)
				items = append(items, baseListItem{
					AppToken:  appToken,
					Title:     file.Name,
					FileToken: file.Token,
					URL:       file.URL,
				})
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", appToken, file.Name, file.Token, file.URL))
			}
			payload := map[string]any{"bases": items}
			text := tableText([]string{"app_token", "title", "file_token", "url"}, lines, "no bases found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search text (or provide as positional argument)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of bases to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")
	return cmd
}

func searchBitableFiles(ctx context.Context, state *appState, userToken, query string, limit, pages int) ([]larksdk.DriveFile, error) {
	if state == nil {
		return nil, errors.New("state is required")
	}
	if state.SDK == nil {
		return nil, errors.New("sdk client is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}
	if pages <= 0 {
		return nil, errors.New("pages must be greater than 0")
	}

	files := make([]larksdk.DriveFile, 0, limit)
	pageToken := ""
	remaining := limit
	pageCount := 0
	for {
		if pageCount >= pages {
			break
		}
		pageCount++

		pageSize := remaining
		if pageSize > maxDrivePageSize {
			pageSize = maxDrivePageSize
		}

		result, err := state.SDK.SearchDriveFilesWithUserToken(ctx, userToken, larksdk.SearchDriveFilesRequest{
			Query:     query,
			FileTypes: []string{"bitable"},
			PageSize:  pageSize,
			PageToken: pageToken,
		})
		if err != nil {
			return nil, withUserScopeHintForCommand(state, err)
		}
		files = append(files, result.Files...)
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
	return files, nil
}
