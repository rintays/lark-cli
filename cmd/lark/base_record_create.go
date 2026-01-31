package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseRecordCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldsJSON string
	var fields []string

	cmd := &cobra.Command{
		Use:   "create <table-id>",
		Short: "Create a Bitable record",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(tableID) == "" {
					return errors.New("table-id is required")
				}
				return nil
			}
			if tableID != "" && tableID != args[0] {
				return errors.New("table-id provided twice")
			}
			return cmd.Flags().Set("table-id", args[0])
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			fieldsMap, err := parseBaseRecordFields(fieldsJSON, fields)
			if err != nil {
				return err
			}
			record, err := state.SDK.CreateBaseRecord(context.Background(), token, appToken, tableID, fieldsMap)
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldsJSON, "fields-json", "", "Record fields JSON (object)")
	cmd.Flags().StringArrayVar(&fields, "field", nil, "Record field assignment (repeatable, e.g. --field Title=Task or --field name=Title,value=Task or --field Temp:=12.3)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
