package main

import (
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
	var pages int
	var userAccessToken string

	cmd := &cobra.Command{
		Use:   "search <query>",
		Short: "Search Wiki nodes (v1)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			query = strings.TrimSpace(args[0])
			if query == "" {
				return errors.New("query is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if limit <= 0 {
				return errors.New("limit must be greater than 0")
			}
			if pages <= 0 {
				return errors.New("pages must be greater than 0")
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}

			token := strings.TrimSpace(userAccessToken)
			if token == "" {
				token = strings.TrimSpace(os.Getenv("LARK_USER_ACCESS_TOKEN"))
			}
			if token != "" {
				var err error
				token, err = tokenForOverride(cmd.Context(), state, tokenTypesUser, tokenOverride{
					Token: token,
					Type:  tokenTypeUser,
				})
				if err != nil {
					return err
				}
			} else {
				var err error
				token, err = tokenFor(cmd.Context(), state, tokenTypesUser)
				if err != nil {
					return err
				}
			}

			nodes := make([]larksdk.WikiNodeSearchV1Item, 0, limit)
			pageToken := ""
			pageCount := 0
			remaining := limit
			for {
				if pageCount >= pages {
					break
				}
				pageCount++
				pageSize := remaining
				if pageSize > 200 {
					pageSize = 200
				}
				result, err := state.SDK.SearchWikiNodesV1(cmd.Context(), token, larksdk.WikiNodeSearchV1Request{
					Query:     query,
					SpaceID:   spaceID,
					PageSize:  pageSize,
					PageToken: pageToken,
				})
				if err != nil {
					return withUserScopeHintForCommand(state, err)
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
			text := tableText([]string{"node_token", "obj_type", "title", "url"}, lines, "no nodes found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of nodes to return")
	cmd.Flags().IntVar(&pages, "pages", 1, "max number of pages to fetch")
	cmd.Flags().StringVar(&userAccessToken, "user-access-token", "", "user access token (OAuth)")

	return cmd
}
