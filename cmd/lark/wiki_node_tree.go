package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

type wikiTreeNode struct {
	Node     larksdk.WikiNode `json:"node"`
	Children []wikiTreeNode   `json:"children,omitempty"`
}

type wikiTreeBuilder struct {
	sdk       *larksdk.Client
	token     string
	tokenType tokenType
	spaceID   string
	pageSize  int
	depth     int
	remaining int
}

func (b *wikiTreeBuilder) build(ctx context.Context, parentToken string, depth int) ([]wikiTreeNode, error) {
	if b.remaining == 0 {
		return nil, nil
	}
	if b.depth > 0 && depth > b.depth {
		return nil, nil
	}
	nodes, err := b.listNodes(ctx, parentToken)
	if err != nil {
		return nil, err
	}
	out := make([]wikiTreeNode, 0, len(nodes))
	for _, node := range nodes {
		item := wikiTreeNode{Node: node}
		if b.remaining == 0 {
			out = append(out, item)
			break
		}
		if b.depth == 0 || depth < b.depth {
			children, err := b.build(ctx, node.NodeToken, depth+1)
			if err != nil {
				return nil, err
			}
			item.Children = children
		}
		out = append(out, item)
		if b.remaining == 0 {
			break
		}
	}
	return out, nil
}

func (b *wikiTreeBuilder) listNodes(ctx context.Context, parentToken string) ([]larksdk.WikiNode, error) {
	items := make([]larksdk.WikiNode, 0)
	pageToken := ""
	for {
		if b.remaining == 0 {
			break
		}
		ps := b.pageSize
		if ps <= 0 {
			ps = b.remaining
			if ps > 200 {
				ps = 200
			}
		} else if ps > 200 {
			ps = 200
		}
		req := larksdk.ListWikiNodesRequest{
			SpaceID:         b.spaceID,
			ParentNodeToken: parentToken,
			PageSize:        ps,
			PageToken:       pageToken,
		}
		var result larksdk.ListWikiNodesResult
		var err error
		switch b.tokenType {
		case tokenTypeTenant:
			result, err = b.sdk.ListWikiNodesV2(ctx, b.token, req)
		case tokenTypeUser:
			result, err = b.sdk.ListWikiNodesV2WithUserToken(ctx, b.token, req)
		default:
			return nil, fmt.Errorf("unsupported token type %s", b.tokenType)
		}
		if err != nil {
			return nil, err
		}
		for _, node := range result.Items {
			if b.remaining == 0 {
				break
			}
			items = append(items, node)
			b.remaining--
		}
		if b.remaining == 0 || !result.HasMore || result.PageToken == "" {
			break
		}
		pageToken = result.PageToken
	}
	return items, nil
}

func newWikiNodeTreeCmd(state *appState) *cobra.Command {
	var spaceID string
	var rootNodeToken string
	var rootObjType string
	var depth int
	var limit int
	var pageSize int

	cmd := &cobra.Command{
		Use:   "tree",
		Short: "Render a Wiki node tree (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.NoArgs(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(rootObjType) != "" && strings.TrimSpace(rootNodeToken) == "" {
				return errors.New("root-node-token is required when root-obj-type is provided")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			spaceID = strings.TrimSpace(spaceID)
			rootNodeToken = strings.TrimSpace(rootNodeToken)
			rootObjType = strings.TrimSpace(rootObjType)
			if limit <= 0 {
				limit = 200
			}
			builder := &wikiTreeBuilder{
				spaceID:   spaceID,
				pageSize:  pageSize,
				depth:     depth,
				remaining: limit,
			}

			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				builder.sdk = sdk
				builder.token = token
				builder.tokenType = tokenType

				var rootNode *larksdk.WikiNode
				if rootNodeToken != "" && rootObjType != "" {
					req := larksdk.GetWikiNodeRequest{NodeToken: rootNodeToken, ObjType: rootObjType}
					var node larksdk.WikiNode
					var err error
					switch tokenType {
					case tokenTypeTenant:
						node, err = sdk.GetWikiNodeV2(ctx, token, req)
					case tokenTypeUser:
						node, err = sdk.GetWikiNodeV2WithUserToken(ctx, token, req)
					default:
						return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
					}
					if err != nil {
						return nil, "", err
					}
					rootNode = &node
				}

				parent := rootNodeToken
				nodes, err := builder.build(ctx, parent, 1)
				if err != nil {
					return nil, "", err
				}

				lines := make([]string, 0, len(nodes))
				if rootNodeToken != "" {
					rootLabel := "root"
					if rootNode != nil {
						rootLabel = formatWikiTreeLabel(rootNode.NodeToken, rootNode.Title, rootNode.ObjType)
					} else {
						rootLabel = fmt.Sprintf("root [%s]", rootNodeToken)
					}
					lines = append(lines, rootLabel)
					lines = append(lines, renderWikiTree(nodes, "  ")...)
				} else {
					lines = append(lines, renderWikiTree(nodes, "")...)
				}
				text := strings.Join(lines, "\n")
				if text == "" {
					text = "no nodes found"
				}

				payload := map[string]any{"nodes": nodes}
				if rootNode != nil {
					payload["root"] = rootNode
				} else if rootNodeToken != "" {
					payload["root"] = map[string]any{"node_token": rootNodeToken}
				}
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&rootNodeToken, "root-node-token", "", "root node token (optional)")
	cmd.Flags().StringVar(&rootObjType, "root-obj-type", "", "root node obj type (optional)")
	cmd.Flags().IntVar(&depth, "depth", 0, "max depth to traverse (0 = unlimited)")
	cmd.Flags().IntVar(&limit, "limit", 200, "max number of nodes to return")
	cmd.Flags().IntVar(&pageSize, "page-size", 0, "page size (default: auto)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}

func renderWikiTree(nodes []wikiTreeNode, prefix string) []string {
	lines := make([]string, 0)
	for i, node := range nodes {
		last := i == len(nodes)-1
		branch := "|- "
		if last {
			branch = "`- "
		}
		label := formatWikiTreeLabel(node.Node.NodeToken, node.Node.Title, node.Node.ObjType)
		lines = append(lines, prefix+branch+label)
		childPrefix := prefix + "|  "
		if last {
			childPrefix = prefix + "   "
		}
		if len(node.Children) > 0 {
			lines = append(lines, renderWikiTree(node.Children, childPrefix)...)
		}
	}
	return lines
}

func formatWikiTreeLabel(nodeToken, title, objType string) string {
	label := strings.TrimSpace(title)
	if label == "" {
		label = "untitled"
	}
	out := fmt.Sprintf("%s [%s]", label, nodeToken)
	if strings.TrimSpace(objType) != "" {
		out = fmt.Sprintf("%s (%s)", out, strings.TrimSpace(objType))
	}
	return out
}
