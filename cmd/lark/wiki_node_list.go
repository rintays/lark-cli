package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeListCmd(state *appState) *cobra.Command {
	var spaceID string
	var parentNodeToken string
	var limit int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Wiki nodes (v2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			if strings.TrimSpace(spaceID) == "" {
				return errors.New("space-id is required")
			}
			if limit <= 0 {
				limit = 50
			}
			tenantToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}

			items := make([]larksdk.WikiNode, 0, limit)
			pageToken := ""
			remaining := limit
			for {
				ps := pageSize
				if ps <= 0 {
					ps = remaining
					if ps > 200 {
						ps = 200
					}
				}
				result, err := state.SDK.ListWikiNodesV2(context.Background(), tenantToken, larksdk.ListWikiNodesRequest{
					SpaceID:         strings.TrimSpace(spaceID),
					ParentNodeToken: strings.TrimSpace(parentNodeToken),
					PageSize:        ps,
					PageToken:       pageToken,
				})
				if err != nil {
					return err
				}
				items = append(items, result.Items...)
				if len(items) >= limit || !result.HasMore {
					break
				}
				remaining = limit - len(items)
				pageToken = result.PageToken
				if pageToken == "" {
					break
				}
			}
			if len(items) > limit {
				items = items[:limit]
			}

			payload := map[string]any{"nodes": items}
			lines := make([]string, 0, len(items))
			for _, n := range items {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", n.NodeToken, n.ObjType, n.Title, n.ObjToken))
			}
			text := "no nodes found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&parentNodeToken, "parent-node-token", "", "parent node token (optional)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of nodes to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
