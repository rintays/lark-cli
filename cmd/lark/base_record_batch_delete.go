package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
		Use:   "batch-delete <table-id> [record-id ...]",
		Short: "Batch delete Bitable records",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				return errors.New("table-id is required")
			}
			if tableID != "" {
				return errors.New("table-id provided twice")
			}
			tableID = args[0]
			if len(args) > 1 {
				recordIDs = append(recordIDs, args[1:]...)
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if strings.TrimSpace(tableID) == "" {
				return errors.New("table-id is required")
			}
			if strings.TrimSpace(appToken) == "" {
				return errors.New("app token is required")
			}

			if len(recordIDs) > 0 && (recordIDsJSON != "" || recordIDsFile != "") {
				return errors.New("record-id positional/flag is mutually exclusive with --record-ids-json/--record-ids-file")
			}
			if recordIDsJSON != "" && recordIDsFile != "" {
				return errors.New("record-ids-json and record-ids-file are mutually exclusive")
			}
			if len(recordIDs) == 0 {
				raw, err := readInput(recordIDsJSON, recordIDsFile, "record-ids")
				if err != nil {
					return err
				}
				raw = strings.TrimSpace(raw)
				if raw != "" {
					if err := json.Unmarshal([]byte(raw), &recordIDs); err != nil {
						return fmt.Errorf("invalid record ids JSON: %w", err)
					}
				}
			}
			recordIDs = normalizeStringSliceLocal(recordIDs)
			if len(recordIDs) == 0 {
				return errors.New("at least one record id is required")
			}

			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			results, err := state.SDK.BatchDeleteBaseRecords(context.Background(), token, appToken, tableID, recordIDs)
			if err != nil {
				return err
			}

			payload := map[string]any{"results": results}
			rows := make([][]string, 0, len(results))
			for _, r := range results {
				rows = append(rows, []string{r.RecordID, fmt.Sprintf("%t", r.Deleted)})
			}
			text := tableTextFromRows([]string{"record_id", "deleted"}, rows, "no records deleted")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (positional preferred)")
	_ = cmd.MarkFlagRequired("app-token")
	cmd.Flags().StringArrayVar(&recordIDs, "record-id", nil, "record id to delete (repeatable) (or provide as positional arguments)")
	cmd.Flags().StringVar(&recordIDsJSON, "record-ids-json", "", "record ids JSON (array of strings; or @file)")
	cmd.Flags().StringVar(&recordIDsFile, "record-ids-file", "", "path to file containing record ids JSON (array of strings)")
	cmd.MarkFlagsMutuallyExclusive("record-id", "record-ids-json")
	cmd.MarkFlagsMutuallyExclusive("record-id", "record-ids-file")
	cmd.MarkFlagsMutuallyExclusive("record-ids-json", "record-ids-file")

	return cmd
}

func normalizeStringSliceLocal(in []string) []string {
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
