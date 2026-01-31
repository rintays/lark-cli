package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseRecordCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldsJSON string

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
			fields, err := parseBaseRecordFields(fieldsJSON)
			if err != nil {
				return err
			}
			record, err := state.SDK.CreateBaseRecord(context.Background(), token, appToken, tableID, fields)
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
	cmd.Flags().StringVar(&fieldsJSON, "fields-json", "", "Record fields JSON (raw)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("fields-json")
	return cmd
}

func parseBaseRecordFields(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("fields-json is required")
	}
	var fields map[string]any
	if err := json.Unmarshal([]byte(raw), &fields); err != nil {
		return nil, fmt.Errorf("fields-json must be a JSON object: %w", err)
	}
	if fields == nil {
		return nil, errors.New("fields-json must be a JSON object")
	}
	return fields, nil
}
