package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func newBaseViewInfoCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewID string

	cmd := &cobra.Command{
		Use:   "info <table-id> <view-id>",
		Short: "Get a Bitable view",
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
				if viewID != "" && viewID != args[1] {
					return errors.New("view-id provided twice")
				}
				if err := cmd.Flags().Set("view-id", args[1]); err != nil {
					return err
				}
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
			view, err := state.SDK.GetBaseView(context.Background(), token, appToken, tableID, viewID)
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
	cmd.Flags().StringVar(&viewID, "view-id", "", "Bitable view id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("view-id")
	return cmd
}
