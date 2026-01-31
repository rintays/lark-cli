package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"lark/internal/testutil"
)

func TestSheetsUpdateIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	sheetID := fx.SpreadsheetToken
	sheetRange := fmt.Sprintf("%s!A1:B1", fx.SheetTitle)

	ts := time.Now().Format(time.RFC3339)
	values := "[[\"it\",\"" + ts + "\"]]"

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "sheets", "update", "--spreadsheet-id", sheetID, "--range", sheetRange, "--values", values})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets update failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["update"]; !ok {
		t.Fatalf("expected update in output, got: %v", payload)
	}

	buf.Reset()
	cmd = newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "sheets", "read", "--spreadsheet-id", sheetID, "--range", sheetRange})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets read failed: %v", err)
	}

	var readPayload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &readPayload); err != nil {
		t.Fatalf("invalid json output (read): %v; out=%q", err, buf.String())
	}
	valueRange, ok := readPayload["valueRange"].(map[string]any)
	if !ok {
		t.Fatalf("expected valueRange in read output, got: %v", readPayload)
	}
	readValues, ok := valueRange["values"]
	if !ok {
		t.Fatalf("expected values in read output, got: %v", valueRange)
	}
	readValuesJSON, err := json.Marshal(readValues)
	if err != nil {
		t.Fatalf("marshal read values: %v", err)
	}
	if !strings.Contains(string(readValuesJSON), ts) {
		t.Fatalf("expected read values to include timestamp %q, got: %s", ts, string(readValuesJSON))
	}
}
