package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeInfoCmd(state *appState) *cobra.Command {
	var nodeToken string
	var objType string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show a Wiki node (v2)",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			tenantToken, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			node, err := state.SDK.GetWikiNodeV2(context.Background(), tenantToken, larksdk.GetWikiNodeRequest{
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

	cmd.Flags().StringVar(&nodeToken, "node-token", "", "wiki node token")
	cmd.Flags().StringVar(&objType, "obj-type", "", "object type (docx|doc|sheet|bitable|file|slides|mindnote)")
	_ = cmd.MarkFlagRequired("node-token")
	_ = cmd.MarkFlagRequired("obj-type")
	return cmd
}
