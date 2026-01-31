package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func newBaseRecordUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string
	var fieldsJSON string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a Bitable record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			fields, err := parseBaseRecordFields(fieldsJSON)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			record, err := state.SDK.UpdateBaseRecord(context.Background(), token, appToken, tableID, recordID, fields)
			if err != nil {
				return err
			}
			payload := map[string]any{"record": record}
			text := tableTextRow(
				[]string{"record_id", "created_time", "last_modified_time"},
				[]string{record.RecordID, record.CreatedTime, record.LastModifiedTime},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id")
	cmd.Flags().StringVar(&fieldsJSON, "fields-json", "", "Record fields JSON (raw)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("record-id")
	_ = cmd.MarkFlagRequired("fields-json")
	return cmd
}
