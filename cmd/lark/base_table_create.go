package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseTableCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableName string
	var viewName string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Bitable table",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(tableName) == "" {
					return errors.New("name is required")
				}
				return nil
			}
			if tableName != "" && tableName != args[0] {
				return errors.New("name provided twice")
			}
			return cmd.Flags().Set("name", args[0])
		},
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
			text := tableTextRow([]string{"table_id", "name"}, []string{table.TableID, table.Name})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableName, "name", "", "Table name (or provide as positional argument)")
	cmd.Flags().StringVar(&viewName, "view-name", "", "Default view name (optional)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
