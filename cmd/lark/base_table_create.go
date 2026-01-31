package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newBaseTableCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableName string
	var viewName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Bitable table",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			table, err := state.SDK.CreateBaseTable(context.Background(), token, appToken, tableName, viewName)
			if err != nil {
				return err
			}
			payload := map[string]any{"table": table}
			text := fmt.Sprintf("%s\t%s", table.TableID, table.Name)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableName, "name", "", "Table name")
	cmd.Flags().StringVar(&viewName, "view-name", "", "Default view name (optional)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}
