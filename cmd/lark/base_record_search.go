package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
	var limit int

	cmd := &cobra.Command{
		Use:   "search",
		Short: "Search Bitable records",
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
			text := "no records found"
			if len(lines) > 0 {
				text = strings.Join(lines, "\n")
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	cmd.Flags().StringVar(&viewID, "view-id", "", "Bitable view id")
	cmd.Flags().StringVar(&filterJSON, "filter", "", "Record filter JSON")
	cmd.Flags().StringVar(&filterJSON, "filter-json", "", "Record filter JSON (raw)")
	cmd.Flags().StringVar(&sortJSON, "sort", "", "Record sort JSON")
	cmd.Flags().StringVar(&sortJSON, "sort-json", "", "Record sort JSON (raw)")
	cmd.Flags().IntVar(&limit, "limit", 20, "max records to return")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	return cmd
}
