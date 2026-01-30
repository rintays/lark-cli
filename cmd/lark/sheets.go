package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
)

func newSheetsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Read Sheets data",
	}
	cmd.AddCommand(newSheetsReadCmd(state))
	cmd.AddCommand(newSheetsUpdateCmd(state))
	cmd.AddCommand(newSheetsAppendCmd(state))
	cmd.AddCommand(newSheetsMetadataCmd(state))
	cmd.AddCommand(newSheetsClearCmd(state))
	return cmd
}

func newSheetsReadCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a range from Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			valueRange, err := state.Client.ReadSheetRange(context.Background(), token, spreadsheetID, sheetRange)
			if err != nil {
				return err
			}
			payload := map[string]any{"valueRange": valueRange}
			text := formatSheetValues(valueRange.Values)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	return cmd
}

func newSheetsMetadataCmd(state *appState) *cobra.Command {
	var spreadsheetID string

	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Get spreadsheet metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			metadata, err := state.Client.GetSpreadsheetMetadata(context.Background(), token, spreadsheetID)
			if err != nil {
				return err
			}
			payload := map[string]any{"metadata": metadata}
			text := formatSpreadsheetMetadata(metadata)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	return cmd
}

func newSheetsUpdateCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string
	var valuesRaw string

	cmd := &cobra.Command{
		Use:   "update",
		Short: "Update a range in Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			values, err := parseSheetValues(valuesRaw)
			if err != nil {
				return err
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			update, err := state.Client.UpdateSheetRange(context.Background(), token, spreadsheetID, sheetRange, values)
			if err != nil {
				return err
			}
			payload := map[string]any{"update": update}
			text := formatSheetUpdate(update, sheetRange)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	cmd.Flags().StringVar(&valuesRaw, "values", "", "JSON array of rows, e.g. '[[\"Name\",\"Amount\"],[\"Ada\",42]]'")
	return cmd
}

func newSheetsAppendCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string
	var valuesRaw string
	var insertDataOption string

	cmd := &cobra.Command{
		Use:   "append",
		Short: "Append rows to Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			values, err := parseSheetValues(valuesRaw)
			if err != nil {
				return err
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			appendResult, err := state.Client.AppendSheetRange(context.Background(), token, spreadsheetID, sheetRange, values, insertDataOption)
			if err != nil {
				return err
			}
			payload := map[string]any{"append": appendResult}
			text := formatSheetAppend(appendResult)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	cmd.Flags().StringVar(&valuesRaw, "values", "", "JSON array of rows, e.g. '[[\"Name\",\"Amount\"],[\"Ada\",42]]'")
	cmd.Flags().StringVar(&insertDataOption, "insert-data-option", "", "insert data option (for example: INSERT_ROWS)")
	return cmd
}

func newSheetsClearCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear a range in Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			clearedRange, err := state.Client.ClearSheetRange(context.Background(), token, spreadsheetID, sheetRange)
			if err != nil {
				return err
			}
			payload := map[string]any{"clearedRange": clearedRange}
			return state.Printer.Print(payload, fmt.Sprintf("ok: cleared %s", clearedRange))
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	return cmd
}

func formatSheetValues(values [][]any) string {
	if len(values) == 0 {
		return "no values found"
	}
	lines := make([]string, 0, len(values))
	for _, row := range values {
		cells := make([]string, 0, len(row))
		for _, cell := range row {
			cells = append(cells, fmt.Sprint(cell))
		}
		lines = append(lines, strings.Join(cells, "\t"))
	}
	return strings.Join(lines, "\n")
}

func formatSpreadsheetMetadata(metadata larkapi.SpreadsheetMetadata) string {
	lines := make([]string, 0, len(metadata.Sheets)+1)
	if title := strings.TrimSpace(metadata.Properties.Title); title != "" {
		lines = append(lines, title)
	}
	for _, sheet := range metadata.Sheets {
		if name := strings.TrimSpace(sheet.Title); name != "" {
			lines = append(lines, name)
		}
	}
	if len(lines) == 0 {
		return "no metadata found"
	}
	return strings.Join(lines, "\n")
}

func parseSheetValues(valuesRaw string) ([][]any, error) {
	if strings.TrimSpace(valuesRaw) == "" {
		return nil, errors.New("values is required")
	}
	var values [][]any
	if err := json.Unmarshal([]byte(valuesRaw), &values); err != nil {
		return nil, fmt.Errorf("values must be a JSON array of arrays: %w", err)
	}
	if len(values) == 0 {
		return nil, errors.New("values must include at least one row")
	}
	return values, nil
}

func formatSheetUpdate(update larkapi.SheetValueUpdate, fallbackRange string) string {
	rangeText := strings.TrimSpace(update.UpdatedRange)
	if rangeText == "" {
		rangeText = strings.TrimSpace(fallbackRange)
	}
	if rangeText == "" {
		rangeText = "range"
	}
	text := fmt.Sprintf("ok: updated %s", rangeText)
	if update.UpdatedCells > 0 {
		text = fmt.Sprintf("%s (updated_cells=%d)", text, update.UpdatedCells)
	}
	return text
}

func formatSheetAppend(appendResult larkapi.SheetValueAppend) string {
	rangeText := strings.TrimSpace(appendResult.TableRange)
	if rangeText == "" {
		rangeText = strings.TrimSpace(appendResult.Updates.UpdatedRange)
	}
	if rangeText == "" {
		rangeText = "appended"
	}
	return fmt.Sprintf("%s\t%d\t%d\t%d", rangeText, appendResult.Updates.UpdatedRows, appendResult.Updates.UpdatedColumns, appendResult.Updates.UpdatedCells)
}
