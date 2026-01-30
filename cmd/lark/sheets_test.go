package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestSheetsReadCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values/Sheet1!A1:B2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"valueRange": map[string]any{
					"range": "Sheet1!A1:B2",
					"values": [][]any{
						{"Name", "Amount"},
						{"Ada", 42},
					},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"read", "--spreadsheet-id", "spreadsheet", "--range", "Sheet1!A1:B2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets read error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Name\tAmount") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "Ada\t42") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsUpdateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected valueInputOption: %s", r.URL.Query().Get("valueInputOption"))
		}
		w.Header().Set("Content-Type", "application/json")
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		valueRange, ok := payload["valueRange"].(map[string]any)
		if !ok {
			t.Fatalf("missing valueRange")
		}
		if valueRange["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range: %v", valueRange["range"])
		}
		if values, ok := valueRange["values"].([]any); !ok || len(values) != 2 {
			t.Fatalf("unexpected values: %#v", valueRange["values"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"revision":         12,
				"spreadsheetToken": "spreadsheet",
				"updatedRange":     "Sheet1!A1:B2",
				"updatedRows":      2,
				"updatedColumns":   2,
				"updatedCells":     4,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{
		"update",
		"--spreadsheet-id", "spreadsheet",
		"--range", "Sheet1!A1:B2",
		"--values", `[["Name","Amount"],["Ada",42]]`,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets update error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "Sheet1!A1:B2") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "updated_cells") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsAppendCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values_append" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("insertDataOption") != "INSERT_ROWS" {
			t.Fatalf("unexpected insertDataOption: %s", r.URL.RawQuery)
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected valueInputOption: %s", r.URL.RawQuery)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		valueRange, ok := payload["valueRange"].(map[string]any)
		if !ok {
			t.Fatalf("missing valueRange")
		}
		if valueRange["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range: %v", valueRange["range"])
		}
		if values, ok := valueRange["values"].([]any); !ok || len(values) != 2 {
			t.Fatalf("unexpected values: %#v", valueRange["values"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"revision":         12,
				"spreadsheetToken": "spreadsheet",
				"tableRange":       "Sheet1!A1:B2",
				"updates": map[string]any{
					"updatedRange":   "Sheet1!A1:B2",
					"updatedRows":    2,
					"updatedColumns": 2,
					"updatedCells":   4,
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{
		"append",
		"--spreadsheet-id", "spreadsheet",
		"--range", "Sheet1!A1:B2",
		"--values", `[["Name","Amount"],["Ada",42]]`,
		"--insert-data-option", "INSERT_ROWS",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets append error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Sheet1!A1:B2") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsMetadataCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v3/spreadsheets/spreadsheet" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"spreadsheet": map[string]any{
					"title": "Budget Q1",
					"token": "spreadsheet",
					"url":   "https://example.test/spreadsheet",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"metadata", "--spreadsheet-id", "spreadsheet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets metadata error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Budget Q1") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsMetadataCommandRequiresSpreadsheetID(t *testing.T) {
	requests := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		w.WriteHeader(http.StatusOK)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"metadata"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for missing spreadsheet-id")
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsClearCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values_clear" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("range") != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range query: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"clearedRange": "Sheet1!A1:B2",
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"clear", "--spreadsheet-id", "spreadsheet", "--range", "Sheet1!A1:B2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets clear error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "Sheet1!A1:B2") {
		t.Fatalf("unexpected output: %q", output)
	}
}
