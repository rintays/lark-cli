package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newSheetsRowsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rows",
		Short: "Row operations",
	}
	cmd.AddCommand(newSheetsRowsInsertCmd(state))
	cmd.AddCommand(newSheetsRowsDeleteCmd(state))
	return cmd
}

func newSheetsRowsInsertCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "insert",
		Short: "Insert rows into a sheet",
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
			result, err := state.SDK.InsertSheetRows(context.Background(), token, spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"insert": result}
			text := fmt.Sprintf("ok: inserted rows start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start row index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of rows to insert")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("sheet-id")
	_ = cmd.MarkFlagRequired("start-index")
	_ = cmd.MarkFlagRequired("count")
	return cmd
}

func newSheetsRowsDeleteCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "delete",
		Short: "Delete rows from a sheet",
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
			result, err := state.SDK.DeleteSheetRows(context.Background(), token, spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"delete": result}
			text := fmt.Sprintf("ok: deleted rows start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start row index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of rows to delete")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("sheet-id")
	_ = cmd.MarkFlagRequired("start-index")
	_ = cmd.MarkFlagRequired("count")
	return cmd
}
