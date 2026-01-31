package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newBaseTableDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Bitable table",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	return cmd
}
