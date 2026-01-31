package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseRecordSearchCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var viewID string
	var filterJSON string
	var sortJSON string
	var limit int

	cmd := &cobra.Command{
		Use:   "search <table-id>",
		Short: "Search Bitable records",
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

			req := larksdk.SearchBaseRecordsRequest{
				ViewID:   viewID,
				PageSize: limit,
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
			lines := make([]string, 0, len(records))
			for _, record := range records {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", record.RecordID, record.CreatedTime, record.LastModifiedTime))
			}
			text := tableText([]string{"record_id", "created_time", "last_modified_time"}, lines, "no records found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&viewID, "view-id", "", "Bitable view id")
	cmd.Flags().StringVar(&filterJSON, "filter", "", "Record filter JSON")
	cmd.Flags().StringVar(&filterJSON, "filter-json", "", "Record filter JSON (raw)")
	cmd.Flags().StringVar(&sortJSON, "sort", "", "Record sort JSON")
	cmd.Flags().StringVar(&sortJSON, "sort-json", "", "Record sort JSON (raw)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max records to return")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
