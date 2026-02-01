package main

import (
	"errors"
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
			if _, err := requireSDK(state); err != nil {
				return err
			}
			nodeToken := strings.TrimSpace(args[0])
			objType := strings.TrimSpace(args[1])
			ctx := cmd.Context()
			tenantToken, err := tokenFor(ctx, state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			node, err := state.SDK.GetWikiNodeV2(ctx, tenantToken, larksdk.GetWikiNodeRequest{
				NodeToken: strings.TrimSpace(nodeToken),
				ObjType:   strings.TrimSpace(objType),
			})
			if err != nil {
				return err
			}
			payload := map[string]any{"node": node}
			text := tableTextRow(
				[]string{"node_token", "obj_type", "title", "obj_token"},
				[]string{node.NodeToken, node.ObjType, node.Title, node.ObjToken},
			)
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
