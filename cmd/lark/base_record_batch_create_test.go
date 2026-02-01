package main

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestBaseRecordBatchCreateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/batch_create" {
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
	cmd.SetArgs([]string{"record", "batch-create", "--app-token", "app_1", "--table-id", "tbl_1", "--records", `[{"Title":"Task"},{"Title":"Task2"}]`})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record batch-create error: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "rec_1") || !strings.Contains(out, "rec_2") {
		t.Fatalf("unexpected output: %q", out)
	}
}
