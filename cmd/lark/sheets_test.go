package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larkapi"
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

	legacyClient := &http.Client{Transport: testutil.HandlerRoundTripper{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("legacy client used for sheets read")
	})}}

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
		Client:  &larkapi.Client{BaseURL: "http://legacy.test", HTTPClient: legacyClient},
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

func TestSheetsReadCommandFallbackToAPI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values/Sheet1%21A1:B2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
		SDK:     &larksdk.Client{},
	}

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

func TestSheetsReadCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values/Sheet1%21A1:B2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

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

func TestSheetsUpdateCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected valueInputOption: %s", r.URL.Query().Get("valueInputOption"))
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

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

	legacyClient := &http.Client{Transport: testutil.HandlerRoundTripper{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("legacy client used for sheets update")
	})}}

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
		Client:  &larkapi.Client{BaseURL: "http://legacy.test", HTTPClient: legacyClient},
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

func TestSheetsUpdateCommandFallbackToAPI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPut {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected valueInputOption: %s", r.URL.Query().Get("valueInputOption"))
		}
		w.Header().Set("Content-Type", "application/json")
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
		SDK:     &larksdk.Client{},
	}

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

func TestSheetsAppendCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values_append" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Query().Get("insertDataOption") != "INSERT_ROWS" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

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
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/metainfo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"properties": map[string]any{
					"title": "Budget Q1",
				},
				"sheets": []map[string]any{
					{"sheetId": "s1", "title": "Summary", "index": 0},
					{"sheetId": "s2", "title": "Details", "index": 1},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"metadata", "--spreadsheet-id", "spreadsheet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets metadata error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "Budget Q1") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "Summary") || !strings.Contains(output, "Details") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsClearCommand(t *testing.T) {
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
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

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
