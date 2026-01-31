package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseTableDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "delete <table-id>",
		Short: "Delete a Bitable table",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(tableID) == "" {
					return errors.New("table-id is required")
				}
				return nil
			}
			if tableID != "" && tableID != args[0] {
				return errors.New("table-id provided twice")
			}
			return cmd.Flags().Set("table-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			res, err := state.SDK.DeleteBaseTable(context.Background(), token, appToken, tableID)
			if err != nil {
				return err
			}
			payload := map[string]any{"result": res}
			text := fmt.Sprintf("%t\t%s", res.Deleted, res.TableID)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
