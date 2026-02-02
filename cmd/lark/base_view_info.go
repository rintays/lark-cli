package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseViewInfoCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewID string

	cmd := &cobra.Command{
		Use:   "info <table-id> <view-id>",
		Short: "Get a Bitable view",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			viewID = strings.TrimSpace(args[1])
			if tableID == "" {
				return errors.New("table-id is required")
			}
			if viewID == "" {
				return errors.New("view-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				view, err := sdk.GetBaseView(ctx, token, appToken, tableID, viewID)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"view": view}
				text := tableTextRow([]string{"view_id", "name", "type"}, []string{view.ViewID, view.Name, view.ViewType})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
