package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

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
		Use:   "insert <spreadsheet-id> <sheet-id> <start-index> <count>",
		Short: "Insert rows into a sheet",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(4)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if spreadsheetID != "" && spreadsheetID != args[0] {
					return errors.New("spreadsheet-id provided twice")
				}
				if err := cmd.Flags().Set("spreadsheet-id", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if sheetID != "" && sheetID != args[1] {
					return errors.New("sheet-id provided twice")
				}
				if err := cmd.Flags().Set("sheet-id", args[1]); err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if cmd.Flags().Changed("start-index") && fmt.Sprint(startIndex) != args[2] {
					return errors.New("start-index provided twice")
				}
				if err := cmd.Flags().Set("start-index", args[2]); err != nil {
					return err
				}
			}
			if len(args) > 3 {
				if cmd.Flags().Changed("count") && fmt.Sprint(count) != args[3] {
					return errors.New("count provided twice")
				}
				if err := cmd.Flags().Set("count", args[3]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(spreadsheetID) == "" {
				return errors.New("spreadsheet-id is required")
			}
			if strings.TrimSpace(sheetID) == "" {
				return errors.New("sheet-id is required")
			}
			if !cmd.Flags().Changed("start-index") {
				return errors.New("start-index is required")
			}
			if !cmd.Flags().Changed("count") {
				return errors.New("count is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 {
				return errors.New("start-index must be >= 0")
			}
			if count <= 0 {
				return errors.New("count must be greater than 0")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
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

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token (or provide as positional argument)")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id (or provide as positional argument)")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start row index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of rows to insert")
	return cmd
}

func newSheetsRowsDeleteCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "delete <spreadsheet-id> <sheet-id> <start-index> <count>",
		Short: "Delete rows from a sheet",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(4)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if spreadsheetID != "" && spreadsheetID != args[0] {
					return errors.New("spreadsheet-id provided twice")
				}
				if err := cmd.Flags().Set("spreadsheet-id", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if sheetID != "" && sheetID != args[1] {
					return errors.New("sheet-id provided twice")
				}
				if err := cmd.Flags().Set("sheet-id", args[1]); err != nil {
					return err
				}
			}
			if len(args) > 2 {
				if cmd.Flags().Changed("start-index") && fmt.Sprint(startIndex) != args[2] {
					return errors.New("start-index provided twice")
				}
				if err := cmd.Flags().Set("start-index", args[2]); err != nil {
					return err
				}
			}
			if len(args) > 3 {
				if cmd.Flags().Changed("count") && fmt.Sprint(count) != args[3] {
					return errors.New("count provided twice")
				}
				if err := cmd.Flags().Set("count", args[3]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(spreadsheetID) == "" {
				return errors.New("spreadsheet-id is required")
			}
			if strings.TrimSpace(sheetID) == "" {
				return errors.New("sheet-id is required")
			}
			if !cmd.Flags().Changed("start-index") {
				return errors.New("start-index is required")
			}
			if !cmd.Flags().Changed("count") {
				return errors.New("count is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 {
				return errors.New("start-index must be >= 0")
			}
			if count <= 0 {
				return errors.New("count must be greater than 0")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
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

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token (or provide as positional argument)")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id (or provide as positional argument)")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start row index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of rows to delete")
	return cmd
}
