package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"
	"time"

	"lark/internal/testutil"
)

func TestSheetsAppendIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	sheetID := os.Getenv("LARK_TEST_SHEET_ID")
	sheetRange := os.Getenv("LARK_TEST_SHEET_RANGE")
	if sheetID == "" || sheetRange == "" {
		t.Skip("missing LARK_TEST_SHEET_ID or LARK_TEST_SHEET_RANGE")
	}

	values := "[[\"it-append\",\"" + time.Now().Format(time.RFC3339) + "\"]]"

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "sheets", "append", "--spreadsheet-id", sheetID, "--range", sheetRange, "--values", values})

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
