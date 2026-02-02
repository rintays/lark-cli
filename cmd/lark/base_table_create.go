package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseTableCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableName string

	cmd := &cobra.Command{
		Use:   "create <name>",
		Short: "Create a Bitable table",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableName = strings.TrimSpace(args[0])
			if tableName == "" {
				return argsUsageError(cmd, errors.New("name is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				table, err := sdk.CreateBaseTable(ctx, token, appToken, tableName)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"table": table}
				text := tableTextRow([]string{"table_id", "name"}, []string{table.TableID, table.Name})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
