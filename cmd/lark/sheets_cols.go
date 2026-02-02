package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
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
		Use:   "insert <spreadsheet-token> <sheet-id> <start-index> <count>",
		Short: "Insert columns into a sheet",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(4)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetID = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return argsUsageError(cmd, errors.New("spreadsheet-token is required"))
			}
			if sheetID == "" {
				return argsUsageError(cmd, errors.New("sheet-id is required"))
			}
			if len(args) > 2 {
				if cmd.Flags().Changed("start-index") && fmt.Sprint(startIndex) != args[2] {
					return argsUsageError(cmd, errors.New("start-index provided twice"))
				}
				if err := cmd.Flags().Set("start-index", args[2]); err != nil {
					return err
				}
			}
			if len(args) > 3 {
				if cmd.Flags().Changed("count") && fmt.Sprint(count) != args[3] {
					return argsUsageError(cmd, errors.New("count provided twice"))
				}
				if err := cmd.Flags().Set("count", args[3]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 {
				return usageErrorWithUsage(cmd, "start-index must be >= 0", "", cmd.UsageString())
			}
			if count <= 0 {
				return usageErrorWithUsage(cmd, "count must be greater than 0", "", cmd.UsageString())
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			result, err := state.SDK.InsertSheetCols(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"insert": result}
			text := fmt.Sprintf("ok: inserted cols start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start column index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of columns to insert (end_index = start_index + count)")
	_ = cmd.MarkFlagRequired("start-index")
	_ = cmd.MarkFlagRequired("count")
	return cmd
}

func newSheetsColsDeleteCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetID string
	var startIndex int
	var count int

	cmd := &cobra.Command{
		Use:   "delete <spreadsheet-token> <sheet-id> <start-index> <count>",
		Short: "Delete columns from a sheet",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(4)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetID = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return argsUsageError(cmd, errors.New("spreadsheet-token is required"))
			}
			if sheetID == "" {
				return argsUsageError(cmd, errors.New("sheet-id is required"))
			}
			if len(args) > 2 {
				if cmd.Flags().Changed("start-index") && fmt.Sprint(startIndex) != args[2] {
					return argsUsageError(cmd, errors.New("start-index provided twice"))
				}
				if err := cmd.Flags().Set("start-index", args[2]); err != nil {
					return err
				}
			}
			if len(args) > 3 {
				if cmd.Flags().Changed("count") && fmt.Sprint(count) != args[3] {
					return argsUsageError(cmd, errors.New("count provided twice"))
				}
				if err := cmd.Flags().Set("count", args[3]); err != nil {
					return err
				}
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if startIndex < 0 {
				return usageErrorWithUsage(cmd, "start-index must be >= 0", "", cmd.UsageString())
			}
			if count <= 0 {
				return usageErrorWithUsage(cmd, "count must be greater than 0", "", cmd.UsageString())
			}
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete columns from %s/%s", spreadsheetID, sheetID)); err != nil {
				return err
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			result, err := state.SDK.DeleteSheetCols(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, sheetID, startIndex, count)
			if err != nil {
				return err
			}
			payload := map[string]any{"delete": result}
			text := fmt.Sprintf("ok: deleted cols start=%d count=%d", result.StartIndex, result.Count)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().IntVar(&startIndex, "start-index", 0, "start column index (0-based)")
	cmd.Flags().IntVar(&count, "count", 0, "number of columns to delete (end_index = start_index + count)")
	_ = cmd.MarkFlagRequired("start-index")
	_ = cmd.MarkFlagRequired("count")
	return cmd
}
