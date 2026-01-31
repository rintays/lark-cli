package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"testing"

	"lark/internal/testutil"
)

func TestSheetsClearIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	sheetID := fx.SpreadsheetToken
	// Sheets v2 values endpoints expect ranges prefixed with sheet_id.
	sheetRange := fmt.Sprintf("%s!A1:B10", fx.SheetID)

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "sheets", "clear", "--spreadsheet-id", sheetID, "--range", sheetRange})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets clear failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["cleared_range"]; !ok {
		if len(payload) == 0 {
			t.Fatalf("expected non-empty payload, got: %v", payload)
		}
	}
}
