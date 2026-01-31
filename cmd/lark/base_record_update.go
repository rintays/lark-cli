package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseRecordUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string
	var fieldsJSON string

	cmd := &cobra.Command{
		Use:   "update <table-id> <record-id>",
		Short: "Update a Bitable record",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if tableID != "" && tableID != args[0] {
					return errors.New("table-id provided twice")
				}
				if err := cmd.Flags().Set("table-id", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if recordID != "" && recordID != args[1] {
					return errors.New("record-id provided twice")
				}
				if err := cmd.Flags().Set("record-id", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(tableID) == "" {
				return errors.New("table-id is required")
			}
			if strings.TrimSpace(recordID) == "" {
				return errors.New("record-id is required")
			}
			return nil
		},
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
			text := fmt.Sprintf("%s\t%s\t%s", record.RecordID, record.CreatedTime, record.LastModifiedTime)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldsJSON, "fields-json", "", "Record fields JSON (raw)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("fields-json")
	return cmd
}
