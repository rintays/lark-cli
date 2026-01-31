package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSheetsColsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cols",
		Short: "Column operations",
	}
	cmd.AddCommand(newSheetsColsInsertCmd(state))
	cmd.AddCommand(newSheetsColsDeleteCmd(state))
	return cmd
}

func newSheetsColsInsertCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "insert <spreadsheet-id> <sheet-id> <start-index> <count>",
		Short: "Insert columns into a sheet",
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
			result, err := state.SDK.InsertSheetCols(context.Background(), token, spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"insert": result}
			text := fmt.Sprintf("ok: inserted cols start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token (or provide as positional argument)")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id (or provide as positional argument)")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start column index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of columns to insert")
	return cmd
}

func newSheetsColsDeleteCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "delete <spreadsheet-id> <sheet-id> <start-index> <count>",
		Short: "Delete columns from a sheet",
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
			result, err := state.SDK.DeleteSheetCols(context.Background(), token, spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"delete": result}
			text := fmt.Sprintf("ok: deleted cols start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token (or provide as positional argument)")
	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id (or provide as positional argument)")
	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start column index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of columns to delete")
	return cmd
}
