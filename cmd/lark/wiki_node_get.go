package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeInfoCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info <node-token> <obj-type>",
		Short: "Show a Wiki node (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if strings.TrimSpace(args[0]) == "" {
				return errors.New("node-token is required")
			}
			if strings.TrimSpace(args[1]) == "" {
				return errors.New("obj-type is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			nodeToken := strings.TrimSpace(args[0])
			objType := strings.TrimSpace(args[1])
			return runWithToken(cmd, state, tokenTypesTenantOrUser, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				var node larksdk.WikiNode
				var err error
				req := larksdk.GetWikiNodeRequest{
					NodeToken: nodeToken,
					ObjType:   objType,
				}
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
				payload := map[string]any{"node": node}
				text := tableTextRow(
					[]string{"node_token", "obj_type", "title", "obj_token"},
					[]string{node.NodeToken, node.ObjType, node.Title, node.ObjToken},
				)
				return payload, text, nil
			})
		},
	}
	return cmd
}
