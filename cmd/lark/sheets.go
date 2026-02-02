package main

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newSheetsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Read Sheets data",
		Long: `Sheets are spreadsheet files stored in Drive.

- spreadsheet_token identifies the file (use it as FILE_TOKEN for drive permissions).
- Each spreadsheet contains sheets (tabs) with sheet_id.
- Use lark drive permissions to manage collaborators for sheets.
- Ranges use A1 notation: <sheet_id>!A1:B2; rows/cols act on a sheet.
- Single cell ranges are allowed (A1 or <sheet_id>!A1) and normalized to A1:A1.`,
	}
	annotateAuthServices(cmd, "sheets")
	cmd.AddCommand(newSheetsReadCmd(state))
	cmd.AddCommand(newSheetsCreateCmd(state))
	cmd.AddCommand(newSheetsUpdateCmd(state))
	cmd.AddCommand(newSheetsAppendCmd(state))
	cmd.AddCommand(newSheetsInfoCmd(state))
	cmd.AddCommand(newSheetsDeleteCmd(state))
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
	var sheetID string

	cmd := &cobra.Command{
		Use:   "read <spreadsheet-token> <range>",
		Short: "Read a range from Sheets",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetRange = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return errors.New("spreadsheet-token is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			resolvedRange, err := resolveSheetRange(sheetRange, sheetID)
			if err != nil {
				return err
			}
			valueRange, err := state.SDK.ReadSheetRange(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, resolvedRange)
			if err != nil {
				return err
			}
			if valueRange.MajorDimension == "" && (valueRange.Range != "" || len(valueRange.Values) > 0) {
				valueRange.MajorDimension = "ROWS"
			}
			payload := map[string]any{"valueRange": valueRange}
			text := formatSheetValues(valueRange)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id to prefix the range (use with range like A1:B2 or single cell A1)")
	return cmd
}

func newSheetsInfoCmd(state *appState) *cobra.Command {
	var spreadsheetID string

	cmd := &cobra.Command{
		Use:   "info <spreadsheet-token>",
		Short: "Show spreadsheet info",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			if spreadsheetID == "" {
				return errors.New("spreadsheet-token is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			metadata, err := state.SDK.GetSpreadsheetMetadata(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID)
			if err != nil {
				return err
			}
			payload := map[string]any{"metadata": metadata}
			text := formatSpreadsheetMetadata(metadata)
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}

func newSheetsDeleteCmd(state *appState) *cobra.Command {
	var spreadsheetID string

	cmd := &cobra.Command{
		Use:   "delete <spreadsheet-token>",
		Short: "Delete a spreadsheet",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			if spreadsheetID == "" {
				return argsUsageError(cmd, errors.New("spreadsheet-token is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := confirmDestructive(cmd, state, fmt.Sprintf("delete spreadsheet %s", spreadsheetID)); err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			token, err := tokenFor(cmd.Context(), state, tokenTypesTenant)
			if err != nil {
				return err
			}
			result, err := state.SDK.DeleteDriveFile(cmd.Context(), token, spreadsheetID, "sheet")
			if err != nil {
				return err
			}
			payload := map[string]any{
				"delete":            result,
				"spreadsheet_token": spreadsheetID,
				"type":              "sheet",
			}
			text := tableTextRow(
				[]string{"spreadsheet_token", "type", "task_id"},
				[]string{spreadsheetID, "sheet", result.TaskID},
			)
			return state.Printer.Print(payload, text)
		},
	}

	return cmd
}

func newSheetsUpdateCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string
	var sheetID string
	var valuesRaw string
	var valuesFile string
	var valuesFormat string

	cmd := &cobra.Command{
		Use:   "update <spreadsheet-token> <range>",
		Short: "Update a range in Sheets",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetRange = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return errors.New("spreadsheet-token is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := parseSheetValues(valuesRaw, valuesFile, valuesFormat)
			if err != nil {
				return err
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			resolvedRange, err := resolveSheetRange(sheetRange, sheetID)
			if err != nil {
				return err
			}
			if err := validateSheetRangeForValues(resolvedRange, values); err != nil {
				return err
			}
			update, err := state.SDK.UpdateSheetRange(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, resolvedRange, values)
			if err != nil {
				return err
			}
			payload := map[string]any{"update": update}
			text := formatSheetUpdate(update, resolvedRange)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id to prefix the range (use with range like A1:B2 or single cell A1)")
	cmd.Flags().StringVar(&valuesRaw, "values", "", "JSON rows (or @file), e.g. '[[\"Name\",\"Amount\"],[\"Ada\",42]]'; use --values-format for inline CSV/TSV")
	cmd.Flags().StringVar(&valuesFile, "values-file", "", "Read values from JSON/CSV/TSV file")
	cmd.Flags().StringVar(&valuesFormat, "values-format", "json", "values format for --values (json, csv, tsv)")
	return cmd
}

func newSheetsAppendCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string
	var sheetID string
	var valuesRaw string
	var valuesFile string
	var valuesFormat string
	var insertDataOption string

	cmd := &cobra.Command{
		Use:   "append <spreadsheet-token> <range>",
		Short: "Append rows to Sheets",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetRange = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return errors.New("spreadsheet-token is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			values, err := parseSheetValues(valuesRaw, valuesFile, valuesFormat)
			if err != nil {
				return err
			}
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			resolvedRange, err := resolveSheetRange(sheetRange, sheetID)
			if err != nil {
				return err
			}
			appendResult, err := state.SDK.AppendSheetRange(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, resolvedRange, values, insertDataOption)
			if err != nil {
				return err
			}
			payload := map[string]any{"append": appendResult}
			text := formatSheetAppend(appendResult)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id to prefix the range (use with range like A1:B2 or single cell A1)")
	cmd.Flags().StringVar(&valuesRaw, "values", "", "JSON rows (or @file), e.g. '[[\"Name\",\"Amount\"],[\"Ada\",42]]'; use --values-format for inline CSV/TSV")
	cmd.Flags().StringVar(&valuesFile, "values-file", "", "Read values from JSON/CSV/TSV file")
	cmd.Flags().StringVar(&valuesFormat, "values-format", "json", "values format for --values (json, csv, tsv)")
	cmd.Flags().StringVar(&insertDataOption, "insert-data-option", "", "insert data option (for example: INSERT_ROWS)")
	return cmd
}

func newSheetsClearCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string
	var sheetID string

	cmd := &cobra.Command{
		Use:   "clear <spreadsheet-token> <range>",
		Short: "Clear a range in Sheets",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			token, _, err := parseResourceRef(args[0])
			if err != nil {
				return err
			}
			spreadsheetID = strings.TrimSpace(token)
			sheetRange = strings.TrimSpace(args[1])
			if spreadsheetID == "" {
				return errors.New("spreadsheet-token is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			token, tokenTypeValue, err := resolveAccessToken(cmd.Context(), state, tokenTypesTenantOrUser, nil)
			if err != nil {
				return err
			}
			if _, err := requireSDK(state); err != nil {
				return err
			}
			resolvedRange, err := resolveSheetRange(sheetRange, sheetID)
			if err != nil {
				return err
			}
			result, err := state.SDK.ClearSheetRange(cmd.Context(), token, larksdk.AccessTokenType(tokenTypeValue), spreadsheetID, resolvedRange)
			if err != nil {
				return err
			}
			return state.Printer.Print(result, fmt.Sprintf("ok: cleared %s", result.ClearedRange))
		},
	}

	cmd.Flags().StringVar(&sheetID, "sheet-id", "", "sheet id to prefix the range (use with range like A1:B2 or single cell A1)")
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

func resolveSheetRange(sheetRange string, sheetID string) (string, error) {
	trimmedRange := strings.TrimSpace(sheetRange)
	if trimmedRange == "" {
		return "", errors.New("range is required")
	}
	trimmedSheetID := strings.TrimSpace(sheetID)
	if strings.Contains(trimmedRange, "!") {
		if trimmedSheetID != "" {
			return "", errors.New("range already includes sheet reference; omit --sheet-id")
		}
		normalized, _ := normalizeSheetRange(trimmedRange)
		return normalized, nil
	}
	if trimmedSheetID == "" {
		return "", errors.New("range must include sheet reference (e.g. <sheet_id>!A1:B2 or <sheet_id>!A1) or set --sheet-id")
	}
	resolved := fmt.Sprintf("%s!%s", trimmedSheetID, trimmedRange)
	normalized, _ := normalizeSheetRange(resolved)
	return normalized, nil
}

var a1CellRe = regexp.MustCompile(`^([A-Za-z]+)([0-9]+)$`)

func normalizeSheetRange(sheetRange string) (string, bool) {
	trimmed := strings.TrimSpace(sheetRange)
	if trimmed == "" {
		return "", false
	}
	prefix, cellRange := splitSheetRange(trimmed)
	if strings.Contains(cellRange, ":") {
		return trimmed, false
	}
	if !isA1Cell(cellRange) {
		return trimmed, false
	}
	return fmt.Sprintf("%s%s:%s", prefix, cellRange, cellRange), true
}

func splitSheetRange(sheetRange string) (string, string) {
	if idx := strings.LastIndex(sheetRange, "!"); idx >= 0 {
		return sheetRange[:idx+1], sheetRange[idx+1:]
	}
	return "", sheetRange
}

func isA1Cell(cell string) bool {
	return a1CellRe.MatchString(strings.TrimSpace(cell))
}

func a1ColToNumber(col string) int {
	col = strings.ToUpper(strings.TrimSpace(col))
	if col == "" {
		return 0
	}
	n := 0
	for i := 0; i < len(col); i++ {
		c := col[i]
		if c < 'A' || c > 'Z' {
			return 0
		}
		n = n*26 + int(c-'A'+1)
	}
	return n
}

func a1NumberToCol(num int) string {
	if num <= 0 {
		return ""
	}
	var out []byte
	for num > 0 {
		num--
		out = append([]byte{byte('A' + (num % 26))}, out...)
		num /= 26
	}
	return string(out)
}

func parseA1Cell(cell string) (col int, row int) {
	cell = strings.TrimSpace(cell)
	m := a1CellRe.FindStringSubmatch(cell)
	if len(m) != 3 {
		return 0, 0
	}
	col = a1ColToNumber(m[1])
	row, err := strconv.Atoi(m[2])
	if err != nil || row <= 0 {
		return 0, 0
	}
	return col, row
}

func a1RangeShape(a1 string) (rows int, cols int) {
	a1 = strings.TrimSpace(a1)
	if a1 == "" {
		return 0, 0
	}
	_, cellRange := splitSheetRange(a1)
	parts := strings.Split(cellRange, ":")
	start := parts[0]
	end := start
	if len(parts) == 2 {
		end = parts[1]
	}
	sc, sr := parseA1Cell(start)
	ec, er := parseA1Cell(end)
	if sc <= 0 || sr <= 0 || ec <= 0 || er <= 0 {
		return 0, 0
	}
	if ec < sc {
		sc, ec = ec, sc
	}
	if er < sr {
		sr, er = er, sr
	}
	return er - sr + 1, ec - sc + 1
}

func valuesShape(values [][]any) (rows int, cols int) {
	rows = len(values)
	for _, row := range values {
		if len(row) > cols {
			cols = len(row)
		}
	}
	return rows, cols
}

func suggestRangeForValues(sheetRange string, rows int, cols int) string {
	if rows <= 0 || cols <= 0 {
		return ""
	}
	prefix, cellRange := splitSheetRange(sheetRange)
	start := cellRange
	if idx := strings.Index(start, ":"); idx >= 0 {
		start = start[:idx]
	}
	sc, sr := parseA1Cell(start)
	if sc <= 0 || sr <= 0 {
		return ""
	}
	endCol := a1NumberToCol(sc + cols - 1)
	endRow := sr + rows - 1
	if endCol == "" || endRow <= 0 {
		return ""
	}
	return fmt.Sprintf("%s%s:%s%d", prefix, start, endCol, endRow)
}

func validateSheetRangeForValues(sheetRange string, values [][]any) error {
	rows, cols := valuesShape(values)
	if rows <= 0 || cols <= 0 {
		return nil
	}
	rangeRows, rangeCols := a1RangeShape(sheetRange)
	if rangeRows <= 0 || rangeCols <= 0 {
		return nil
	}
	if rows <= rangeRows && cols <= rangeCols {
		return nil
	}
	suggested := suggestRangeForValues(sheetRange, rows, cols)
	if suggested != "" && suggested != sheetRange {
		return fmt.Errorf("range too small for values: range=%s values=%dx%d; try %s", sheetRange, rows, cols, suggested)
	}
	return fmt.Errorf("range too small for values: range=%s values=%dx%d", sheetRange, rows, cols)
}

func parseSheetValues(valuesRaw string, valuesFile string, valuesFormat string) ([][]any, error) {
	raw := strings.TrimSpace(valuesRaw)
	filePath := strings.TrimSpace(valuesFile)
	if raw == "" && filePath == "" {
		return nil, errors.New("values is required (use --values or --values-file)")
	}
	if raw != "" && filePath != "" {
		return nil, errors.New("values and values-file cannot both be set")
	}
	if filePath != "" {
		return parseSheetValuesFile(filePath)
	}
	if strings.HasPrefix(raw, "@") {
		path := strings.TrimSpace(strings.TrimPrefix(raw, "@"))
		if path == "" {
			return nil, errors.New("values file path is required after @")
		}
		return parseSheetValuesFile(path)
	}
	return parseSheetValuesInline(raw, valuesFormat)
}

func parseSheetValuesFile(path string) ([][]any, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read values file: %w", err)
	}
	ext := strings.ToLower(filepath.Ext(path))
	if ext == ".csv" {
		return parseSheetValuesCSV(data)
	}
	if ext == ".tsv" || ext == ".tab" {
		return parseSheetValuesTSV(data)
	}
	return parseSheetValuesJSON(data)
}

func parseSheetValuesInline(raw string, valuesFormat string) ([][]any, error) {
	format := strings.ToLower(strings.TrimSpace(valuesFormat))
	if format == "" {
		format = "json"
	}
	switch format {
	case "json":
		return parseSheetValuesJSON([]byte(raw))
	case "csv":
		return parseSheetValuesCSV([]byte(raw))
	case "tsv":
		return parseSheetValuesTSV([]byte(raw))
	default:
		return nil, fmt.Errorf("values-format must be one of json, csv, tsv")
	}
}

func parseSheetValuesJSON(data []byte) ([][]any, error) {
	var values [][]any
	if err := json.Unmarshal(data, &values); err != nil {
		return nil, fmt.Errorf("values must be a JSON array of arrays: %w", err)
	}
	if len(values) == 0 {
		return nil, errors.New("values must include at least one row")
	}
	return values, nil
}

func parseSheetValuesCSV(data []byte) ([][]any, error) {
	return parseSheetValuesDelimited(data, ',')
}

func parseSheetValuesTSV(data []byte) ([][]any, error) {
	return parseSheetValuesDelimited(data, '\t')
}

func parseSheetValuesDelimited(data []byte, comma rune) ([][]any, error) {
	reader := csv.NewReader(bytes.NewReader(data))
	reader.TrimLeadingSpace = true
	reader.Comma = comma
	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("values must be valid delimited data: %w", err)
	}
	if len(records) == 0 {
		return nil, errors.New("values must include at least one row")
	}
	values := make([][]any, 0, len(records))
	for _, record := range records {
		row := make([]any, len(record))
		for i, cell := range record {
			row[i] = cell
		}
		values = append(values, row)
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
