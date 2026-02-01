package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
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
				return err
			}
			tableID = strings.TrimSpace(args[0])
			recordID = strings.TrimSpace(args[1])
			if tableID == "" {
				return errors.New("table-id is required")
			}
			if recordID == "" {
				return errors.New("record-id is required")
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
			result, err := state.SDK.DeleteBaseRecord(context.Background(), token, appToken, tableID, recordID)
			if err != nil {
				return err
			}
			payload := map[string]any{"record_id": result.RecordID, "deleted": result.Deleted}
			text := "deleted"
			if result.RecordID != "" {
				text = result.RecordID
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
