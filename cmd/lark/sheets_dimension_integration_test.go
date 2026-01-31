package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"lark/internal/testutil"
)

func runIntegrationCLI(t *testing.T, fx integrationFixtures, args ...string) map[string]any {
	t.Helper()

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(append([]string{"--config", fx.ConfigPath, "--json"}, args...))

	if err := cmd.Execute(); err != nil {
		t.Fatalf("command failed: %v; args=%v; out=%q", err, args, buf.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; args=%v; out=%q", err, args, buf.String())
	}
	return payload
}

func readFirstColumnStrings(t *testing.T, payload map[string]any) []string {
	t.Helper()

	vr, ok := payload["valueRange"].(map[string]any)
	if !ok {
		t.Fatalf("expected valueRange in output, got: %v", payload)
	}
	vals, ok := vr["values"].([]any)
	if !ok {
		t.Fatalf("expected valueRange.values in output, got: %v", vr)
	}

	out := make([]string, 0, len(vals))
	for _, rowAny := range vals {
		row, ok := rowAny.([]any)
		if !ok || len(row) == 0 {
			out = append(out, "")
			continue
		}
		cell, _ := row[0].(string)
		out = append(out, cell)
	}
	return out
}

func readRowStrings(t *testing.T, payload map[string]any) []string {
	t.Helper()

	vr, ok := payload["valueRange"].(map[string]any)
	if !ok {
		t.Fatalf("expected valueRange in output, got: %v", payload)
	}
	vals, ok := vr["values"].([]any)
	if !ok {
		t.Fatalf("expected valueRange.values in output, got: %v", vr)
	}
	if len(vals) == 0 {
		return nil
	}
	row, ok := vals[0].([]any)
	if !ok {
		t.Fatalf("expected valueRange.values[0] to be an array, got: %T", vals[0])
	}
	out := make([]string, 0, len(row))
	for _, cellAny := range row {
		cell, _ := cellAny.(string)
		out = append(out, cell)
	}
	return out
}

func TestSheetsDimensionIntegration_RowsAndColsInsertDelete(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	spreadsheetToken := fx.SpreadsheetToken
	sheetID := fx.SheetID

	t.Run("rows insert+delete shifts values", func(t *testing.T) {
		// Put two sentinel values in A1 and A2.
		_ = runIntegrationCLI(t, fx,
			"sheets", "update",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:A2", sheetID),
			"--values", `[["top"],["bottom"]]`,
		)

		// Insert 1 row at index 1 (between row 1 and row 2).
		insertPayload := runIntegrationCLI(t, fx,
			"sheets", "rows", "insert",
			"--spreadsheet-id", spreadsheetToken,
			"--sheet-id", sheetID,
			"--start-index", "1",
			"--count", "1",
		)
		if _, ok := insertPayload["insert"]; !ok {
			t.Fatalf("expected insert in output, got: %v", insertPayload)
		}

		// Write into the newly inserted row so the subsequent read is stable.
		_ = runIntegrationCLI(t, fx,
			"sheets", "update",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A2:A2", sheetID),
			"--values", `[["middle"]]`,
		)

		readPayload := runIntegrationCLI(t, fx,
			"sheets", "read",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:A3", sheetID),
		)
		col := readFirstColumnStrings(t, readPayload)
		if len(col) < 3 || col[0] != "top" || col[1] != "middle" || col[2] != "bottom" {
			t.Fatalf("unexpected values after insert: %v", col)
		}

		deletePayload := runIntegrationCLI(t, fx,
			"sheets", "rows", "delete",
			"--spreadsheet-id", spreadsheetToken,
			"--sheet-id", sheetID,
			"--start-index", "1",
			"--count", "1",
		)
		if _, ok := deletePayload["delete"]; !ok {
			t.Fatalf("expected delete in output, got: %v", deletePayload)
		}

		readPayload = runIntegrationCLI(t, fx,
			"sheets", "read",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:A2", sheetID),
		)
		col = readFirstColumnStrings(t, readPayload)
		if len(col) < 2 || col[0] != "top" || col[1] != "bottom" {
			t.Fatalf("unexpected values after delete: %v", col)
		}
	})

	t.Run("cols insert+delete shifts values", func(t *testing.T) {
		// Put two sentinel values in A1 and B1.
		_ = runIntegrationCLI(t, fx,
			"sheets", "update",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:B1", sheetID),
			"--values", `[["left","right"]]`,
		)

		insertPayload := runIntegrationCLI(t, fx,
			"sheets", "cols", "insert",
			"--spreadsheet-id", spreadsheetToken,
			"--sheet-id", sheetID,
			"--start-index", "1",
			"--count", "1",
		)
		if _, ok := insertPayload["insert"]; !ok {
			t.Fatalf("expected insert in output, got: %v", insertPayload)
		}

		// Write into the newly inserted column.
		_ = runIntegrationCLI(t, fx,
			"sheets", "update",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!B1:B1", sheetID),
			"--values", `[["middlecol"]]`,
		)

		readPayload := runIntegrationCLI(t, fx,
			"sheets", "read",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:C1", sheetID),
		)
		row := readRowStrings(t, readPayload)
		if len(row) < 3 || row[0] != "left" || row[1] != "middlecol" || row[2] != "right" {
			t.Fatalf("unexpected values after insert: %v", row)
		}

		deletePayload := runIntegrationCLI(t, fx,
			"sheets", "cols", "delete",
			"--spreadsheet-id", spreadsheetToken,
			"--sheet-id", sheetID,
			"--start-index", "1",
			"--count", "1",
		)
		if _, ok := deletePayload["delete"]; !ok {
			t.Fatalf("expected delete in output, got: %v", deletePayload)
		}

		readPayload = runIntegrationCLI(t, fx,
			"sheets", "read",
			"--spreadsheet-id", spreadsheetToken,
			"--range", fmt.Sprintf("%s!A1:B1", sheetID),
		)
		row = readRowStrings(t, readPayload)
		if len(row) < 2 || row[0] != "left" || row[1] != "right" {
			t.Fatalf("unexpected values after delete: %v", row)
		}
	})
}
