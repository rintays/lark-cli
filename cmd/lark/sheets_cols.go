package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newSheetsColsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cols",
		Short: "Column operations",
	}
	cmd.AddCommand(newSheetsColsInsertCmd(state))
	return cmd
}

func newSheetsColsInsertCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert columns into a sheet",
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 {
				return errors.New("start-index must be >= 0")
			}
			if count <= 0 {
				return errors.New("count must be greater than 0")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			result, err := state.SDK.InsertSheetCols(context.Background(), token, spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"insert": result}
			text := fmt.Sprintf("ok: inserted cols start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start column index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of columns to insert")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("sheet-id")
	_ = cmd.MarkFlagRequired("start-index")
	_ = cmd.MarkFlagRequired("count")
	return cmd
}
