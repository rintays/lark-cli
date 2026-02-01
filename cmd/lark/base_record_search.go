package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseRecordSearchCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewID string
	var filterJSON string
	var sortJSON string
	var fieldsCSV string
	var limit int

	cmd := &cobra.Command{
		Use:   "search <table-id>",
		Short: "Search Bitable records",
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

			automaticFields := true
			req := larksdk.SearchBaseRecordsRequest{
				ViewID:          viewID,
				PageSize:        limit,
				AutomaticFields: &automaticFields,
			}
			fieldNames, err := parseBaseRecordSearchFieldNames(fieldsCSV)
			if err != nil {
				return err
			}
			if len(fieldNames) > 0 {
				req.FieldNames = fieldNames
			}
			if filterJSON != "" {
				req.Filter = json.RawMessage(filterJSON)
			}
			if sortJSON != "" {
				req.Sort = json.RawMessage(sortJSON)
			}

			result, err := state.SDK.SearchBaseRecords(context.Background(), token, appToken, tableID, req)
			if err != nil {
				return err
			}
			records := result.Items
			payload := map[string]any{"records": records}
			headers := buildBaseRecordSearchHeaders(fieldNames, records)
			rows := make([][]string, 0, len(records))
			for _, record := range records {
				row := make([]string, 0, len(headers))
				row = append(row, formatBaseRecordCell(record.RecordID))
				for _, fieldName := range headers[1 : len(headers)-2] {
					row = append(row, formatBaseRecordFieldValue(record.Fields[fieldName]))
				}
				row = append(row, formatBaseRecordCell(record.CreatedTime), formatBaseRecordCell(record.LastModifiedTime))
				rows = append(rows, row)
			}
			text := tableTextFromRows(headers, rows, "no records found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&viewID, "view-id", "", "Bitable view id")
	cmd.Flags().StringVar(&fieldsCSV, "fields", "", "Comma-separated field names to return/display (default: all returned by API)")
	cmd.Flags().StringVar(&filterJSON, "filter", "", "Record filter JSON")
	cmd.Flags().StringVar(&filterJSON, "filter-json", "", "Record filter JSON (raw)")
	cmd.Flags().StringVar(&sortJSON, "sort", "", "Record sort JSON")
	cmd.Flags().StringVar(&sortJSON, "sort-json", "", "Record sort JSON (raw)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max records to return")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func parseBaseRecordSearchFieldNames(raw string) ([]string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	parts := strings.Split(raw, ",")
	names := make([]string, 0, len(parts))
	seen := make(map[string]struct{}, len(parts))
	for _, part := range parts {
		name := strings.TrimSpace(part)
		if name == "" {
			return nil, errors.New("fields must not include empty names")
		}
		if _, exists := seen[name]; exists {
			return nil, fmt.Errorf("field %q provided twice", name)
		}
		seen[name] = struct{}{}
		names = append(names, name)
	}
	if len(names) == 0 {
		return nil, errors.New("fields must include at least one name")
	}
	return names, nil
}

func buildBaseRecordSearchHeaders(fieldNames []string, records []larksdk.BaseRecord) []string {
	headers := []string{"record_id"}
	if len(fieldNames) == 0 {
		fieldNames = collectBaseRecordFieldNames(records)
	}
	headers = append(headers, fieldNames...)
	headers = append(headers, "created_time", "last_modified_time")
	return headers
}

func collectBaseRecordFieldNames(records []larksdk.BaseRecord) []string {
	if len(records) == 0 {
		return nil
	}
	seen := map[string]struct{}{}
	for _, record := range records {
		for name := range record.Fields {
			if strings.TrimSpace(name) == "" {
				continue
			}
			seen[name] = struct{}{}
		}
	}
	if len(seen) == 0 {
		return nil
	}
	names := make([]string, 0, len(seen))
	for name := range seen {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func formatBaseRecordFieldValue(value any) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case string:
		return formatBaseRecordCell(v)
	case bool:
		return strconv.FormatBool(v)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case int:
		return strconv.Itoa(v)
	case int64:
		return strconv.FormatInt(v, 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case json.Number:
		return v.String()
	}
	data, err := json.Marshal(value)
	if err == nil {
		return formatBaseRecordCell(string(data))
	}
	return formatBaseRecordCell(fmt.Sprint(value))
}

func formatBaseRecordCell(value string) string {
	return baseRecordCellReplacer.Replace(value)
}

var baseRecordCellReplacer = strings.NewReplacer("\t", "\\t", "\n", "\\n", "\r", "\\r")
