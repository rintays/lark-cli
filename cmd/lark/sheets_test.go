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

func TestSheetsCreateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch r.URL.Path {
		case "/open-apis/sheets/v3/spreadsheets":
			if r.Method != http.MethodPost {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["title"] != "Budget Q1" {
				t.Fatalf("unexpected title: %v", payload["title"])
			}
			if payload["folder_token"] != "fld" {
				t.Fatalf("unexpected folder_token: %v", payload["folder_token"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"spreadsheet": map[string]any{
						"spreadsheet_token": "spreadsheet",
					},
				},
			})
		case "/open-apis/sheets/v3/spreadsheets/spreadsheet":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"spreadsheet": map[string]any{
						"title":    "Budget Q1",
						"token":    "spreadsheet",
						"owner_id": "ou_1",
					},
				},
			})
		case "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/query":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"sheets": []map[string]any{
						{
							"sheet_id": "sheet_1",
							"title":    "Sheet1",
							"index":    0,
							"hidden":   false,
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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
	cmd.SetArgs([]string{"create", "--title", "Budget Q1", "--folder-id", "fld"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets create error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "spreadsheet") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "sheet_1") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsCreateCommandRequiresTitle(t *testing.T) {
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
	cmd.SetArgs([]string{"create"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for missing title")
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsInfoCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch r.URL.Path {
		case "/open-apis/sheets/v3/spreadsheets/spreadsheet":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"spreadsheet": map[string]any{
						"title":         "Budget Q1",
						"token":         "spreadsheet",
						"owner_id":      "ou_1",
						"url":           "https://example.test/spreadsheet",
						"folder_token":  "fld_1",
						"without_mount": false,
					},
				},
			})
		case "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/query":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"sheets": []map[string]any{
						{
							"sheet_id":      "sheet_1",
							"title":         "Sheet1",
							"index":         0,
							"hidden":        false,
							"resource_type": "sheet",
							"grid_properties": map[string]any{
								"frozen_row_count":    1,
								"frozen_column_count": 0,
								"row_count":           100,
								"column_count":        10,
							},
							"merges": []map[string]any{
								{
									"start_row_index":    0,
									"end_row_index":      1,
									"start_column_index": 0,
									"end_column_index":   2,
								},
							},
						},
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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
	cmd.SetArgs([]string{"info", "--spreadsheet-id", "spreadsheet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets info error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "title\tBudget Q1") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "token\tspreadsheet") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "owner_id\tou_1") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "sheets[0].title\tSheet1") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsInfoCommandRequiresSpreadsheetID(t *testing.T) {
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
	cmd.SetArgs([]string{"info"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for missing spreadsheet-id")
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsDeleteCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/spreadsheet" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "sheet" {
			t.Fatalf("unexpected type: %s", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"task_id": "task_1",
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
	cmd.SetArgs([]string{"delete", "--spreadsheet-id", "spreadsheet"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets delete error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "spreadsheet") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "task_1") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsDeleteCommandRequiresSpreadsheetID(t *testing.T) {
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
	cmd.SetArgs([]string{"delete"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error for missing spreadsheet-id")
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsClearCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		// Clear is implemented as update with empty strings.
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		vr, ok := payload["valueRange"].(map[string]any)
		if !ok {
			t.Fatalf("missing valueRange: %v", payload)
		}
		if vr["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range: %v", vr["range"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"updatedCells": 4,
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

func TestSheetsColsInsertCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/sheet/insert_dimension" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		dimensionRange, ok := payload["dimension_range"].(map[string]any)
		if !ok {
			t.Fatalf("missing dimension_range")
		}
		if dimensionRange["major_dimension"] != "COLUMNS" {
			t.Fatalf("unexpected major_dimension: %#v", dimensionRange["major_dimension"])
		}
		if startIndex, ok := dimensionRange["start_index"].(float64); !ok || int(startIndex) != 4 {
			t.Fatalf("unexpected start_index: %#v", dimensionRange["start_index"])
		}
		if endIndex, ok := dimensionRange["end_index"].(float64); !ok || int(endIndex) != 7 {
			t.Fatalf("unexpected end_index: %#v", dimensionRange["end_index"])
		}
		if _, ok := payload["dimensionRange"]; !ok {
			t.Fatalf("missing dimensionRange")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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
		"cols",
		"insert",
		"--spreadsheet-id", "spreadsheet",
		"--sheet-id", "sheet",
		"--start-index", "4",
		"--count", "3",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets cols insert error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok: inserted cols") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsColsInsertRequiresSheetID(t *testing.T) {
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
	cmd.SetArgs([]string{
		"cols",
		"insert",
		"--spreadsheet-id", "spreadsheet",
		"--start-index", "1",
		"--count", "2",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing sheet-id")
	}
	if err.Error() != "required flag(s) \"sheet-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsColsDeleteCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/sheet/delete_dimension" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		dimensionRange, ok := payload["dimension_range"].(map[string]any)
		if !ok {
			t.Fatalf("missing dimension_range")
		}
		if dimensionRange["major_dimension"] != "COLUMNS" {
			t.Fatalf("unexpected major_dimension: %#v", dimensionRange["major_dimension"])
		}
		if startIndex, ok := dimensionRange["start_index"].(float64); !ok || int(startIndex) != 3 {
			t.Fatalf("unexpected start_index: %#v", dimensionRange["start_index"])
		}
		if endIndex, ok := dimensionRange["end_index"].(float64); !ok || int(endIndex) != 5 {
			t.Fatalf("unexpected end_index: %#v", dimensionRange["end_index"])
		}
		if _, ok := payload["dimensionRange"]; !ok {
			t.Fatalf("missing dimensionRange")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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
		"cols",
		"delete",
		"--spreadsheet-id", "spreadsheet",
		"--sheet-id", "sheet",
		"--start-index", "3",
		"--count", "2",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets cols delete error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok: deleted cols") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsColsDeleteRequiresSheetID(t *testing.T) {
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
	cmd.SetArgs([]string{
		"cols",
		"delete",
		"--spreadsheet-id", "spreadsheet",
		"--start-index", "3",
		"--count", "2",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing sheet-id")
	}
	if err.Error() != "required flag(s) \"sheet-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsRowsInsertCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/sheet/insert_dimension" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		dimensionRange, ok := payload["dimension_range"].(map[string]any)
		if !ok {
			t.Fatalf("missing dimension_range")
		}
		if dimensionRange["major_dimension"] != "ROWS" {
			t.Fatalf("unexpected major_dimension: %#v", dimensionRange["major_dimension"])
		}
		if startIndex, ok := dimensionRange["start_index"].(float64); !ok || int(startIndex) != 1 {
			t.Fatalf("unexpected start_index: %#v", dimensionRange["start_index"])
		}
		if endIndex, ok := dimensionRange["end_index"].(float64); !ok || int(endIndex) != 3 {
			t.Fatalf("unexpected end_index: %#v", dimensionRange["end_index"])
		}
		if _, ok := payload["dimensionRange"]; !ok {
			t.Fatalf("missing dimensionRange")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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
		"rows",
		"insert",
		"--spreadsheet-id", "spreadsheet",
		"--sheet-id", "sheet",
		"--start-index", "1",
		"--count", "2",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets rows insert error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok: inserted rows") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsRowsDeleteCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v3/spreadsheets/spreadsheet/sheets/sheet/delete_dimension" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		dimensionRange, ok := payload["dimension_range"].(map[string]any)
		if !ok {
			t.Fatalf("missing dimension_range")
		}
		if dimensionRange["major_dimension"] != "ROWS" {
			t.Fatalf("unexpected major_dimension: %#v", dimensionRange["major_dimension"])
		}
		if startIndex, ok := dimensionRange["start_index"].(float64); !ok || int(startIndex) != 2 {
			t.Fatalf("unexpected start_index: %#v", dimensionRange["start_index"])
		}
		if endIndex, ok := dimensionRange["end_index"].(float64); !ok || int(endIndex) != 6 {
			t.Fatalf("unexpected end_index: %#v", dimensionRange["end_index"])
		}
		if _, ok := payload["dimensionRange"]; !ok {
			t.Fatalf("missing dimensionRange")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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
		"rows",
		"delete",
		"--spreadsheet-id", "spreadsheet",
		"--sheet-id", "sheet",
		"--start-index", "2",
		"--count", "4",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets rows delete error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "ok: deleted rows") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestSheetsRowsInsertRequiresSheetID(t *testing.T) {
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
	cmd.SetArgs([]string{
		"rows",
		"insert",
		"--spreadsheet-id", "spreadsheet",
		"--start-index", "1",
		"--count", "2",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing sheet-id")
	}
	if err.Error() != "required flag(s) \"sheet-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsRowsInsertRequiresStartIndex(t *testing.T) {
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
	cmd.SetArgs([]string{
		"rows",
		"insert",
		"--spreadsheet-id", "spreadsheet",
		"--sheet-id", "sheet",
		"--count", "2",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing start-index")
	}
	if err.Error() != "required flag(s) \"start-index\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}

func TestSheetsRowsDeleteRequiresSheetID(t *testing.T) {
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
	cmd.SetArgs([]string{
		"rows",
		"delete",
		"--spreadsheet-id", "spreadsheet",
		"--start-index", "1",
		"--count", "2",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatalf("expected error for missing sheet-id")
	}
	if err.Error() != "required flag(s) \"sheet-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
	if requests != 0 {
		t.Fatalf("expected no HTTP requests, got %d", requests)
	}
}
