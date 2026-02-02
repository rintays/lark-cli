package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseViewCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewName string
	var viewType string

	cmd := &cobra.Command{
		Use:   "create <table-id> <name>",
		Short: "Create a Bitable view",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			if tableID == "" {
				return errors.New("table-id is required")
			}
			if len(args) > 1 {
				if viewName != "" && viewName != args[1] {
					return errors.New("name provided twice")
				}
				if err := cmd.Flags().Set("name", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(viewName) == "" {
				return errors.New("name is required")
			}
			if strings.TrimSpace(viewType) == "" {
				_ = cmd.Flags().Set("view-type", "grid")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				view, err := sdk.CreateBaseView(ctx, token, appToken, tableID, viewName, viewType)
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
	cmd.Flags().StringVar(&viewName, "name", "", "View name (or provide as positional argument)")
	cmd.Flags().StringVar(&viewType, "view-type", "grid", "View type (grid|kanban|gallery|gantt|form)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
