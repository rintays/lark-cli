package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseRecordBatchUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordsRaw string
	var recordsFile string
	var clientToken string
	var ignoreConsistencyCheck bool

	cmd := &cobra.Command{
		Use:   "batch-update <table-id>",
		Short: "Update multiple Bitable records in one request",
		Long: `Update multiple Bitable records in one request.

Provide records via --records (JSON array) or --records-file. Each record must include record_id and fields.

Example:
  lark base record batch-update --app-token app_x --table-id tbl_x \
    --records '[{"record_id":"rec_x","fields":{"Title":"A"}},{"record_id":"rec_y","fields":{"Done":true}}]'

You can also pass a file path with @:
  --records @records.json`,
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
			records, err := parseBaseRecordBatchUpdateRecords(recordsRaw, recordsFile)
			if err != nil {
				return err
			}
			updated, err := state.SDK.BatchUpdateBaseRecords(context.Background(), token, appToken, tableID, records, clientToken, ignoreConsistencyCheck)
			if err != nil {
				return err
			}
			payload := map[string]any{"records": updated}
			lines := make([]string, 0, len(updated))
			for _, record := range updated {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", record.RecordID, record.CreatedTime, record.LastModifiedTime))
			}
			text := tableText([]string{"record_id", "created_time", "last_modified_time"}, lines, "no records updated")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&recordsRaw, "records", "", "JSON array of record objects (or @file)")
	cmd.Flags().StringVar(&recordsFile, "records-file", "", "Path to JSON file with records (array of objects)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token (best-effort; uses core fallback when SDK lacks support)")
	cmd.Flags().BoolVar(&ignoreConsistencyCheck, "ignore-consistency-check", false, "Ignore field consistency checks")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func parseBaseRecordBatchUpdateRecords(recordsRaw string, recordsFile string) ([]larksdk.BaseRecordUpdate, error) {
	raw := strings.TrimSpace(recordsRaw)
	filePath := strings.TrimSpace(recordsFile)
	if raw == "" && filePath == "" {
		return nil, errors.New("records is required (use --records or --records-file)")
	}
	if raw != "" && filePath != "" {
		return nil, errors.New("records and records-file cannot both be set")
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read records file: %w", err)
		}
		return parseBaseRecordBatchUpdateRecordsJSON(data)
	}
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
		if path == "" {
			return nil, errors.New("records file path is required after @")
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read records file: %w", err)
		}
		return parseBaseRecordBatchUpdateRecordsJSON(data)
	}
	return parseBaseRecordBatchUpdateRecordsJSON([]byte(recordsRaw))
}

func parseBaseRecordBatchUpdateRecordsJSON(data []byte) ([]larksdk.BaseRecordUpdate, error) {
	var records []larksdk.BaseRecordUpdate
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("records must be a JSON array of objects: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("records must include at least one record")
	}
	for i := range records {
		record := &records[i]
		record.RecordID = strings.TrimSpace(record.RecordID)
		if record.RecordID == "" {
			return nil, fmt.Errorf("records[%d].record_id is required", i)
		}
		if record.Fields == nil {
			return nil, fmt.Errorf("records[%d].fields is required", i)
		}
		if len(record.Fields) == 0 {
			return nil, fmt.Errorf("records[%d].fields must not be empty", i)
		}
	}
	return records, nil
}
