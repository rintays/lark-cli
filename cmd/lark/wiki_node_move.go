package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeMoveCmd(state *appState) *cobra.Command {
	var spaceID string
	var targetParentNodeToken string
	var targetSpaceID string

	cmd := &cobra.Command{
		Use:   "move <node-token>",
		Short: "Move a Wiki node (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("node-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeToken := strings.TrimSpace(args[0])
			spaceID = strings.TrimSpace(spaceID)
			targetParentNodeToken = strings.TrimSpace(targetParentNodeToken)
			targetSpaceID = strings.TrimSpace(targetSpaceID)
			if targetParentNodeToken == "" && targetSpaceID == "" {
				return errors.New("target-parent-node-token or target-space-id is required")
			}
			req := larksdk.MoveWikiNodeRequest{
				SpaceID:               spaceID,
				NodeToken:             nodeToken,
				TargetParentNodeToken: targetParentNodeToken,
				TargetSpaceID:         targetSpaceID,
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var node larksdk.WikiNode
				var err error
				switch tokenType {
				case tokenTypeTenant:
					node, err = sdk.MoveWikiNodeV2(ctx, token, req)
				case tokenTypeUser:
					node, err = sdk.MoveWikiNodeV2WithUserToken(ctx, token, req)
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
	cmd.Flags().StringVar(&targetParentNodeToken, "target-parent-node-token", "", "target parent node token")
	cmd.Flags().StringVar(&targetSpaceID, "target-space-id", "", "target space ID (optional)")
	_ = cmd.MarkFlagRequired("space-id")
	return cmd
}
