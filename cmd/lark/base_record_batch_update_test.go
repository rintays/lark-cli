package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestBaseRecordBatchUpdateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_update" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tenant-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read body: %v", err)
		}
		var payload map[string]any
		if err := json.Unmarshal(body, &payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		records, ok := payload["records"].([]any)
		if !ok {
			t.Fatalf("missing records payload: %#v", payload["records"])
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 records, got %#v", records)
		}
		first, ok := records[0].(map[string]any)
		if !ok {
			t.Fatalf("first record not object: %#v", records[0])
		}
		if first["record_id"] != "rec_1" {
			t.Fatalf("unexpected record_id: %#v", first["record_id"])
		}
		fields, ok := first["fields"].(map[string]any)
		if !ok {
			t.Fatalf("missing fields: %#v", first["fields"])
		}
		if fields["Title"] != "Task" {
			t.Fatalf("unexpected fields: %#v", fields)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"records": []map[string]any{
					{
						"record_id":          "rec_1",
						"created_time":       1700000000,
						"last_modified_time": 1700000001,
					},
					{
						"record_id":          "rec_2",
						"created_time":       1700000002,
						"last_modified_time": 1700000003,
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
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "batch-update", "--app-token", "app_1", "--table-id", "tbl_1", "--records", `[{"record_id":"rec_1","fields":{"Title":"Task"}},{"record_id":"rec_2","fields":{"Title":"Task2"}}]`})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-update error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestBaseRecordBatchUpdateCommandWithRecordsFileAndClientToken(t *testing.T) {
	recordsPath := filepath.Join(t.TempDir(), "records.json")
	if err := os.WriteFile(recordsPath, []byte(`[
  {"record_id":"rec_1","fields":{"Title":"Task"}},
  {"record_id":"rec_2","fields":{"Done":true}}
]`), 0o600); err != nil {
		t.Fatalf("write records file: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_update" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("client_token") != "token_1" {
			t.Fatalf("unexpected client_token: %s", r.URL.Query().Get("client_token"))
		}
		if r.URL.Query().Get("ignore_consistency_check") != "true" {
			t.Fatalf("unexpected ignore_consistency_check: %s", r.URL.Query().Get("ignore_consistency_check"))
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		records, ok := payload["records"].([]any)
		if !ok || len(records) != 2 {
			t.Fatalf("expected records array length 2, got %T len=%d", payload["records"], len(records))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"records": []map[string]any{
					{"record_id": "rec_1", "created_time": "1700000000", "last_modified_time": "1700000001"},
					{"record_id": "rec_2", "created_time": "1700000002", "last_modified_time": "1700000003"},
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
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "batch-update", "--app-token", "app_1", "--table-id", "tbl_1", "--records", "@" + recordsPath, "--client-token", "token_1", "--ignore-consistency-check"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-update error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}
