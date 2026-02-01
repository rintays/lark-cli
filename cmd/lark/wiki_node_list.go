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
			if strings.TrimSpace(spaceID) == "" {
				return errors.New("space-id is required")
			}
			if limit <= 0 {
				limit = 50
			}
			spaceID = strings.TrimSpace(spaceID)
			parentNodeToken = strings.TrimSpace(parentNodeToken)
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
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
					req := larksdk.ListWikiNodesRequest{
						SpaceID:         spaceID,
						ParentNodeToken: parentNodeToken,
						PageSize:        ps,
						PageToken:       pageToken,
					}
					var result larksdk.ListWikiNodesResult
					var err error
					switch tokenType {
					case tokenTypeTenant:
						result, err = sdk.ListWikiNodesV2(ctx, token, req)
					case tokenTypeUser:
						result, err = sdk.ListWikiNodesV2WithUserToken(ctx, token, req)
					default:
						return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
					}
					if err != nil {
						return nil, "", err
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
				text := tableText([]string{"node_token", "obj_type", "title", "obj_token"}, lines, "no nodes found")
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&parentNodeToken, "parent-node-token", "", "parent node token (optional)")
	cmd.Flags().IntVar(&limit, "limit", 50, "max number of nodes to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
