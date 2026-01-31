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

func TestBaseRecordBatchDeleteCommandSendsBatchDeleteRequest(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_delete" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
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

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "batch-delete", "--app-token", "app_1", "tbl_1", "rec_1", "rec_2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-delete error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}
