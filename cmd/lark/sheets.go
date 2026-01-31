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

func newSheetsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Read Sheets data",
	}
	cmd.AddCommand(newSheetsReadCmd(state))
	cmd.AddCommand(newSheetsCreateCmd(state))
	cmd.AddCommand(newSheetsUpdateCmd(state))
	cmd.AddCommand(newSheetsAppendCmd(state))
	cmd.AddCommand(newSheetsInfoCmd(state))
	cmd.AddCommand(newSheetsClearCmd(state))
	cmd.AddCommand(newSheetsColsCmd(state))
	cmd.AddCommand(newSheetsRowsCmd(state))
	cmd.AddCommand(newSheetsSearchCmd(state))
	cmd.AddCommand(newSheetsListCmd(state))
	return cmd
}

func newSheetsReadCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a range from Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			valueRange, err := state.SDK.ReadSheetRange(context.Background(), token, spreadsheetID, sheetRange)
			if err != nil {
				return err
			}
			payload := map[string]any{"valueRange": valueRange}
			text := formatSheetValues(valueRange)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func newSheetsInfoCmd(state *appState) *cobra.Command {
	var spreadsheetID string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show spreadsheet info",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			metadata, err := state.SDK.GetSpreadsheetMetadata(context.Background(), token, spreadsheetID)
			if err != nil {
				return err
			}
			payload := map[string]any{"metadata": metadata}
			text := formatSpreadsheetMetadata(metadata)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
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
			values, err := parseSheetValues(valuesRaw)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			update, err := state.SDK.UpdateSheetRange(context.Background(), token, spreadsheetID, sheetRange, values)
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
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("range")
	_ = cmd.MarkFlagRequired("values")
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
			values, err := parseSheetValues(valuesRaw)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			appendResult, err := state.SDK.AppendSheetRange(context.Background(), token, spreadsheetID, sheetRange, values, insertDataOption)
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
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("range")
	_ = cmd.MarkFlagRequired("values")
	return cmd
}

func newSheetsClearCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string

	cmd := &cobra.Command{
		Use:   "clear",
		Short: "Clear a range in Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			result, err := state.SDK.ClearSheetRange(context.Background(), token, spreadsheetID, sheetRange)
			if err != nil {
				return err
			}
			return state.Printer.Print(result, fmt.Sprintf("ok: cleared %s", result.ClearedRange))
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	_ = cmd.MarkFlagRequired("spreadsheet-id")
	_ = cmd.MarkFlagRequired("range")
	return cmd
}

func formatSheetValues(valueRange larksdk.SheetValueRange) string {
	values := valueRange.Values
	if len(values) == 0 {
		return "no values found"
	}
	rows := make([][]string, 0, len(values))
	maxCols := 0
	for _, row := range values {
		cells := make([]string, 0, len(row))
		for _, cell := range row {
			cells = append(cells, fmt.Sprint(cell))
		}
		if len(cells) > maxCols {
			maxCols = len(cells)
		}
		rows = append(rows, cells)
	}
	if maxCols == 0 {
		return "no values found"
	}
	headers := make([]string, maxCols)
	for i := 0; i < maxCols; i++ {
		headers[i] = fmt.Sprintf("col%d", i+1)
	}
	return tableTextFromRows(headers, rows, "no values found")
}

func formatSpreadsheetMetadata(metadata larksdk.SpreadsheetMetadata) string {
	rows := [][]string{
		{"token", infoValue(metadata.Spreadsheet.SpreadsheetToken)},
		{"title", infoValue(metadata.Spreadsheet.Title)},
		{"url", infoValue(metadata.Spreadsheet.URL)},
		{"owner_id", infoValue(metadata.Spreadsheet.OwnerID)},
		{"sheets.count", fmt.Sprintf("%d", len(metadata.Sheets))},
	}
	for i, sheet := range metadata.Sheets {
		prefix := fmt.Sprintf("sheets[%d]", i)
		resourceType := infoValue(sheet.ResourceType)
		frozenRowCount := "-"
		frozenColumnCount := "-"
		rowCount := "-"
		columnCount := "-"
		if sheet.GridProperties != nil {
			frozenRowCount = fmt.Sprintf("%d", sheet.GridProperties.FrozenRowCount)
			frozenColumnCount = fmt.Sprintf("%d", sheet.GridProperties.FrozenColumnCount)
			rowCount = fmt.Sprintf("%d", sheet.GridProperties.RowCount)
			columnCount = fmt.Sprintf("%d", sheet.GridProperties.ColumnCount)
		}
		rows = append(rows,
			[]string{prefix + ".sheet_id", infoValue(sheet.SheetID)},
			[]string{prefix + ".title", infoValue(sheet.Title)},
			[]string{prefix + ".index", fmt.Sprintf("%d", sheet.Index)},
			[]string{prefix + ".hidden", fmt.Sprintf("%t", sheet.Hidden)},
			[]string{prefix + ".resource_type", resourceType},
			[]string{prefix + ".grid_properties.frozen_row_count", frozenRowCount},
			[]string{prefix + ".grid_properties.frozen_column_count", frozenColumnCount},
			[]string{prefix + ".grid_properties.row_count", rowCount},
			[]string{prefix + ".grid_properties.column_count", columnCount},
			[]string{prefix + ".merges.count", fmt.Sprintf("%d", len(sheet.Merges))},
		)
		for j, merge := range sheet.Merges {
			mergePrefix := fmt.Sprintf("%s.merges[%d]", prefix, j)
			rows = append(rows,
				[]string{mergePrefix + ".start_row_index", fmt.Sprintf("%d", merge.StartRowIndex)},
				[]string{mergePrefix + ".end_row_index", fmt.Sprintf("%d", merge.EndRowIndex)},
				[]string{mergePrefix + ".start_column_index", fmt.Sprintf("%d", merge.StartColumnIndex)},
				[]string{mergePrefix + ".end_column_index", fmt.Sprintf("%d", merge.EndColumnIndex)},
			)
		}
	}
	return formatInfoTable(rows, "no metadata found")
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

func formatSheetUpdate(update larksdk.SheetValueUpdate, fallbackRange string) string {
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

func formatSheetAppend(appendResult larksdk.SheetValueAppend) string {
	rangeText := strings.TrimSpace(appendResult.TableRange)
	if rangeText == "" {
		rangeText = strings.TrimSpace(appendResult.Updates.UpdatedRange)
	}
	if rangeText == "" {
		rangeText = "appended"
	}
	return tableTextRow(
		[]string{"range", "updated_rows", "updated_columns", "updated_cells"},
		[]string{
			rangeText,
			fmt.Sprintf("%d", appendResult.Updates.UpdatedRows),
			fmt.Sprintf("%d", appendResult.Updates.UpdatedColumns),
			fmt.Sprintf("%d", appendResult.Updates.UpdatedCells),
		},
	)
}
