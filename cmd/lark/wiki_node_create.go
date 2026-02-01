package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeCreateCmd(state *appState) *cobra.Command {
	var spaceID string
	var objType string
	var objToken string
	var parentNodeToken string
	var nodeType string
	var originNodeToken string
	var originSpaceID string
	var title string

	cmd := &cobra.Command{
		Use:   "create [obj-type] [obj-token]",
		Short: "Create a Wiki node (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if len(args) > 0 {
				if strings.TrimSpace(objType) != "" && objType != args[0] {
					return errors.New("obj-type provided twice")
				}
				objType = args[0]
			}
			if len(args) > 1 {
				if strings.TrimSpace(objToken) != "" && objToken != args[1] {
					return errors.New("obj-token provided twice")
				}
				objToken = args[1]
			}

			objType = strings.TrimSpace(objType)
			objToken = strings.TrimSpace(objToken)
			originNodeToken = strings.TrimSpace(originNodeToken)
			nodeTypeValue := strings.TrimSpace(nodeType)

			if objType == "" {
				return errors.New("obj-type is required")
			}
			if nodeTypeValue == "" {
				if originNodeToken != "" {
					nodeTypeValue = "shortcut"
				} else {
					nodeTypeValue = "origin"
				}
			}
			switch nodeTypeValue {
			case "origin":
				if objToken == "" {
					return errors.New("obj-token is required for origin nodes")
				}
			case "shortcut":
				if originNodeToken == "" {
					return errors.New("origin-node-token is required for shortcut nodes")
				}
			default:
				return errors.New("node-type must be origin or shortcut")
			}
			nodeType = nodeTypeValue
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			req := larksdk.CreateWikiNodeRequest{
				SpaceID:         strings.TrimSpace(spaceID),
				ObjType:         strings.TrimSpace(objType),
				ObjToken:        strings.TrimSpace(objToken),
				ParentNodeToken: strings.TrimSpace(parentNodeToken),
				NodeType:        strings.TrimSpace(nodeType),
				OriginNodeToken: strings.TrimSpace(originNodeToken),
				OriginSpaceID:   strings.TrimSpace(originSpaceID),
				Title:           strings.TrimSpace(title),
			}
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var node larksdk.WikiNode
				var err error
				switch tokenType {
				case tokenTypeTenant:
					node, err = sdk.CreateWikiNodeV2(ctx, token, req)
				case tokenTypeUser:
					node, err = sdk.CreateWikiNodeV2WithUserToken(ctx, token, req)
				default:
					return nil, "", fmt.Errorf("unsupported token type %s", tokenType)
				}
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"node": node}
				text := tableTextRow(
					[]string{"node_token", "obj_type", "title", "obj_token"},
					[]string{node.NodeToken, node.ObjType, node.Title, node.ObjToken},
				)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&spaceID, "space-id", "", "Wiki space ID")
	cmd.Flags().StringVar(&objType, "obj-type", "", "object type (doc, docx, sheet, slides, bitable, mindnote, file)")
	cmd.Flags().StringVar(&objToken, "obj-token", "", "object token (required for origin nodes)")
	cmd.Flags().StringVar(&parentNodeToken, "parent-node-token", "", "parent node token (optional)")
	cmd.Flags().StringVar(&nodeType, "node-type", "", "node type (origin or shortcut)")
	cmd.Flags().StringVar(&originNodeToken, "origin-node-token", "", "origin node token (required for shortcut nodes)")
	cmd.Flags().StringVar(&originSpaceID, "origin-space-id", "", "origin space ID (optional, shortcut nodes)")
	cmd.Flags().StringVar(&title, "title", "", "node title (optional)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
