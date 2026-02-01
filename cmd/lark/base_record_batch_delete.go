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

func newBaseRecordBatchDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordIDs []string
	var recordIDsJSON string
	var recordIDsFile string

	cmd := &cobra.Command{
		Use:   "batch-delete <table-id>",
		Short: "Delete multiple Bitable records in one request",
		Long: `Delete multiple Bitable records in one request.

Provide record ids via repeatable --record-id, or via JSON using --record-ids-json/--record-ids-file.

Example:
  lark base record batch-delete --app-token app_x --table-id tbl_x \
    --record-id rec_1 --record-id rec_2

You can also pass JSON directly:
  --record-ids-json '["rec_1","rec_2"]'

Or pass a file path with @:
  --record-ids-json @record_ids.json`,
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
			ids, err := parseBaseRecordBatchDeleteRecordIDs(recordIDs, recordIDsJSON, recordIDsFile)
			if err != nil {
				return err
			}
			results, err := state.SDK.BatchDeleteBaseRecords(context.Background(), token, appToken, tableID, ids)
			if err != nil {
				return err
			}
			payload := map[string]any{"records": results}
			lines := make([]string, 0, len(results))
			for _, r := range results {
				lines = append(lines, fmt.Sprintf("%s\t%t", r.RecordID, r.Deleted))
			}
			text := tableText([]string{"record_id", "deleted"}, lines, "no records deleted")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringArrayVar(&recordIDs, "record-id", nil, "Bitable record id (repeatable)")
	cmd.Flags().StringVar(&recordIDsJSON, "record-ids-json", "", "JSON array of record ids (or @file)")
	cmd.Flags().StringVar(&recordIDsFile, "record-ids-file", "", "Path to JSON file with record ids (array of strings)")
	_ = cmd.MarkFlagRequired("app-token")
	cmd.MarkFlagsMutuallyExclusive("record-id", "record-ids-json")
	cmd.MarkFlagsMutuallyExclusive("record-id", "record-ids-file")
	cmd.MarkFlagsMutuallyExclusive("record-ids-json", "record-ids-file")
	return cmd
}

func parseBaseRecordBatchDeleteRecordIDs(recordIDs []string, recordIDsJSON string, recordIDsFile string) ([]string, error) {
	raw := strings.TrimSpace(recordIDsJSON)
	filePath := strings.TrimSpace(recordIDsFile)

	if len(recordIDs) == 0 && raw == "" && filePath == "" {
		return nil, errors.New("record ids are required (use --record-id, --record-ids-json, or --record-ids-file)")
	}
	if len(recordIDs) > 0 {
		if raw != "" || filePath != "" {
			return nil, errors.New("record-id cannot be used together with record-ids-json or record-ids-file")
		}
		ids := normalizeBaseRecordIDs(recordIDs)
		if len(ids) == 0 {
			return nil, errors.New("record-id must include at least one non-empty id")
		}
		return ids, nil
	}
	if raw != "" && filePath != "" {
		return nil, errors.New("record-ids-json and record-ids-file cannot both be set")
	}

	var data []byte
	if filePath != "" {
		b, err := os.ReadFile(filePath)
		if err != nil {
			return nil, fmt.Errorf("read record ids file: %w", err)
		}
		data = b
	} else if strings.HasPrefix(raw, "@") {
		path := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
		if path == "" {
			return nil, errors.New("record ids file path is required after @")
		}
		b, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("read record ids file: %w", err)
		}
		data = b
	} else {
		data = []byte(raw)
	}

	var ids []string
	if err := json.Unmarshal(data, &ids); err != nil {
		return nil, fmt.Errorf("record ids must be a JSON array of strings: %w", err)
	}
	ids = normalizeBaseRecordIDs(ids)
	if len(ids) == 0 {
		return nil, errors.New("record ids must include at least one record")
	}
	return ids, nil
}

func normalizeBaseRecordIDs(in []string) []string {
	out := make([]string, 0, len(in))
	seen := map[string]struct{}{}
	for _, v := range in {
		v = strings.TrimSpace(v)
		if v == "" {
			continue
		}
		if _, ok := seen[v]; ok {
			continue
		}
		seen[v] = struct{}{}
		out = append(out, v)
	}
	return out
}
