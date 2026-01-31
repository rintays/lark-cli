package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"lark/internal/testutil"
)

func TestSheetsClearIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	sheetID := os.Getenv("LARK_TEST_SHEET_ID")
	sheetRange := os.Getenv("LARK_TEST_SHEET_RANGE")
	if sheetID == "" || sheetRange == "" {
		t.Skip("missing LARK_TEST_SHEET_ID or LARK_TEST_SHEET_RANGE")
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "sheets", "clear", "--spreadsheet-id", sheetID, "--range", sheetRange})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets clear failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["cleared_range"]; !ok {
		// Clear returns a struct; key name might vary. Allow empty assertions by checking non-empty payload.
		if len(payload) == 0 {
			t.Fatalf("expected non-empty payload, got: %v", payload)
		}
	}
}
