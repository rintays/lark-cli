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

func newBaseFieldDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldID string

	cmd := &cobra.Command{
		Use:   "delete <table-id> <field-id>",
		Short: "Delete a Bitable field",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			fieldID = strings.TrimSpace(args[1])
			if tableID == "" {
				return argsUsageError(cmd, errors.New("table-id is required"))
			}
			if fieldID == "" {
				return argsUsageError(cmd, errors.New("field-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete field %s from table %s", fieldID, tableID)); err != nil {
				return err
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				res, err := sdk.DeleteBaseField(ctx, token, appToken, tableID, fieldID)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"result": res}
				text := tableTextRow([]string{"deleted", "field_id"}, []string{strconv.FormatBool(res.Deleted), res.FieldID})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
