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

func TestBaseRecordBatchDeleteCommandWithRecordIDFlags(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_delete" {
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
			t.Fatalf("expected records array, got %T", payload["records"])
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 record ids, got %d", len(records))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"records": []map[string]any{
					{"record_id": "rec_1", "deleted": true},
					{"record_id": "rec_2", "deleted": true},
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
		Force:   true,
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "batch-delete", "tbl_1", "--app-token", "app_1", "--record-id", "rec_1", "--record-id", "rec_2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-delete error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}

func TestBaseRecordBatchDeleteCommandWithRecordIDsJSONAtFile(t *testing.T) {
	idsPath := filepath.Join(t.TempDir(), "ids.json")
	if err := os.WriteFile(idsPath, []byte(`["rec_1","rec_2"]`), 0o600); err != nil {
		t.Fatalf("write ids file: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_delete" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		records, ok := payload["records"].([]any)
		if !ok {
			t.Fatalf("expected records array, got %T", payload["records"])
		}
		if len(records) != 2 {
			t.Fatalf("expected 2 record ids, got %d", len(records))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "msg": "ok", "data": map[string]any{"records": []map[string]any{{"record_id": "rec_1", "deleted": true}, {"record_id": "rec_2", "deleted": true}}}})
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
		Force:   true,
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "batch-delete", "tbl_1", "--app-token", "app_1", "--record-ids-json", "@" + idsPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-delete error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}
