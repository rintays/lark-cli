package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newWikiNodeGetCmd(state *appState) *cobra.Command {
	var nodeToken string
	var objType string

	cmd := &cobra.Command{
		Use:   "get <node-token> <obj-type>",
		Short: "Get a Wiki node (v2)",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if nodeToken != "" && nodeToken != args[0] {
					return errors.New("node-token provided twice")
				}
				if err := cmd.Flags().Set("node-token", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if objType != "" && objType != args[1] {
					return errors.New("obj-type provided twice")
				}
				if err := cmd.Flags().Set("obj-type", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(nodeToken) == "" {
				return errors.New("node-token is required")
			}
			if strings.TrimSpace(objType) == "" {
				return errors.New("obj-type is required")
			}
			return nil
		},
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
			text := fmt.Sprintf("%s\t%s\t%s\t%s", node.NodeToken, node.ObjType, node.Title, node.ObjToken)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&nodeToken, "node-token", "", "wiki node token (or provide as positional argument)")
	cmd.Flags().StringVar(&objType, "obj-type", "", "object type (docx|doc|sheet|bitable|file|slides|mindnote) (or provide as positional argument)")
	return cmd
}
