package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"lark/internal/testutil"
)

func TestSheetsAppendIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	sheetID := fx.SpreadsheetToken
	// Sheets v2 values endpoints expect ranges prefixed with sheet_id.
	sheetRange := fmt.Sprintf("%s!A1:B1", fx.SheetID)

	values := "[[\"it-append\",\"" + time.Now().Format(time.RFC3339) + "\"]]"

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "sheets", "append", "--spreadsheet-id", sheetID, "--range", sheetRange, "--values", values})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets append failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["append"]; !ok {
		t.Fatalf("expected append in output, got: %v", payload)
	}
}
