package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func newBaseRecordDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete a Bitable record",
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("record-id")
	return cmd
}
