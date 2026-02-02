package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseRecordDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string

	cmd := &cobra.Command{
		Use:   "delete <table-id> <record-id>",
		Short: "Delete a Bitable record",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			recordID = strings.TrimSpace(args[1])
			if tableID == "" {
				return argsUsageError(cmd, errors.New("table-id is required"))
			}
			if recordID == "" {
				return argsUsageError(cmd, errors.New("record-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete record %s from table %s", recordID, tableID)); err != nil {
				return err
			}
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				result, err := sdk.DeleteBaseRecord(ctx, token, appToken, tableID, recordID)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"record_id": result.RecordID, "deleted": result.Deleted}
				text := "deleted"
				if result.RecordID != "" {
					text = result.RecordID
				}
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
