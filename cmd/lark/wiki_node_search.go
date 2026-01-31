package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeSearchCmd(state *appState) *cobra.Command {
	var query string
	var spaceID string
	var limit int
	var userAccessToken string

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Wiki nodes (v1)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}

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

			nodes := make([]larksdk.WikiNodeSearchV1Item, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				pageSize := remaining
				if pageSize > 200 {
					pageSize = 200
				}
				result, err := state.SDK.SearchWikiNodesV1(context.Background(), token, larksdk.WikiNodeSearchV1Request{
					Query:     query,
					SpaceID:   spaceID,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return err
				}
				nodes = append(nodes, result.Items...)
				if len(nodes) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(nodes)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(nodes) > limit {
				nodes = nodes[:limit]
			}

			payload := map[string]any{"nodes": nodes}
			lines := make([]string, 0, len(nodes))
			for _, n := range nodes {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", n.NodeToken, n.ObjType, n.Title, n.URL))
			}
			text := "no nodes found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&query, "query", "", "search text")
	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of nodes to return")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")
	_ = cmd.MarkFlagRequired("query")

	return cmd
}
