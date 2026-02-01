package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
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
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if tableID != "" && tableID != args[0] {
					return errors.New("table-id provided twice")
				}
				if err := cmd.Flags().Set("table-id", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if viewName != "" && viewName != args[1] {
					return errors.New("name provided twice")
				}
				if err := cmd.Flags().Set("name", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(viewType) == "" {
				_ = cmd.Flags().Set("view-type", "grid")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			view, err := state.SDK.CreateBaseView(context.Background(), token, appToken, tableID, viewName, viewType)
			if err != nil {
				return err
			}
			payload := map[string]any{"view": view}
			text := tableTextRow([]string{"view_id", "name", "type"}, []string{view.ViewID, view.Name, view.ViewType})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&viewName, "name", "", "View name (or provide as positional argument)")
	cmd.Flags().StringVar(&viewType, "view-type", "grid", "View type (grid|kanban|gallery|gantt|form)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
