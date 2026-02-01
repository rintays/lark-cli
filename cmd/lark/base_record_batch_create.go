package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseRecordBatchCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordsRaw string
	var recordsFile string
	var clientToken string
	var ignoreConsistencyCheck bool

	cmd := &cobra.Command{
		Use:   "batch-create <table-id>",
		Short: "Create multiple Bitable records in one request",
		Long: `Create multiple Bitable records in one request.

Provide record fields via --records (JSON array of objects) or --records-file.

Example:
  lark base record batch-create tbl_x --app-token app_x \
    --records '[{"Title":"A"},{"Title":"B","Done":true}]'

You can also pass a file path with @:
  --records @records.json`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			if tableID == "" {
				return errors.New("table-id is required")
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
			records, err := parseBaseRecordBatchCreateRecords(recordsRaw, recordsFile)
			if err != nil {
				return err
			}
			created, err := state.SDK.BatchCreateBaseRecords(context.Background(), token, appToken, tableID, records, clientToken, ignoreConsistencyCheck)
			if err != nil {
				return err
			}
			payload := map[string]any{"records": created}
			lines := make([]string, 0, len(created))
			for _, record := range created {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", record.RecordID, record.CreatedTime.String(), record.LastModifiedTime.String()))
			}
			text := tableText([]string{"record_id", "created_time", "last_modified_time"}, lines, "no records created")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&recordsRaw, "records", "", "JSON array of field objects (or @file)")
	cmd.Flags().StringVar(&recordsFile, "records-file", "", "Path to JSON file with records (array of objects)")
	cmd.Flags().StringVar(&clientToken, "client-token", "", "idempotency token")
	cmd.Flags().BoolVar(&ignoreConsistencyCheck, "ignore-consistency-check", false, "Ignore field consistency checks")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func parseBaseRecordBatchCreateRecords(recordsRaw string, recordsFile string) ([]map[string]any, error) {
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
		return parseBaseRecordBatchCreateRecordsJSON(data)
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
		return parseBaseRecordBatchCreateRecordsJSON(data)
	}
	return parseBaseRecordBatchCreateRecordsJSON([]byte(recordsRaw))
}

func parseBaseRecordBatchCreateRecordsJSON(data []byte) ([]map[string]any, error) {
	var records []map[string]any
	if err := json.Unmarshal(data, &records); err != nil {
		return nil, fmt.Errorf("records must be a JSON array of objects: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("records must include at least one record")
	}
	for _, fields := range records {
		if fields == nil {
			return nil, errors.New("records include null record")
		}
		if len(fields) == 0 {
			return nil, errors.New("records include empty record")
		}
	}
	return records, nil
}
