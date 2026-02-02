package main

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseTableDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "delete <table-id>",
		Short: "Delete a Bitable table",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			if tableID == "" {
				return argsUsageError(cmd, errors.New("table-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete table %s", tableID)); err != nil {
				return err
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				res, err := sdk.DeleteBaseTable(ctx, token, appToken, tableID)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"result": res}
				text := tableTextRow([]string{"deleted", "table_id"}, []string{strconv.FormatBool(res.Deleted), res.TableID})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
