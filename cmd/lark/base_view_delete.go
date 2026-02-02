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

func newBaseViewDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewID string

	cmd := &cobra.Command{
		Use:   "delete <table-id> <view-id>",
		Short: "Delete a Bitable view",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			viewID = strings.TrimSpace(args[1])
			if tableID == "" {
				return argsUsageError(cmd, errors.New("table-id is required"))
			}
			if viewID == "" {
				return argsUsageError(cmd, errors.New("view-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete view %s from table %s", viewID, tableID)); err != nil {
				return err
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				res, err := sdk.DeleteBaseView(ctx, token, appToken, tableID, viewID)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"result": res}
				text := tableTextRow([]string{"deleted", "view_id"}, []string{strconv.FormatBool(res.Deleted), res.ViewID})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
