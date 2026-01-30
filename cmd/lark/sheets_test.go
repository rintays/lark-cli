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
	"lark/internal/output"
	"lark/internal/testutil"
)

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
