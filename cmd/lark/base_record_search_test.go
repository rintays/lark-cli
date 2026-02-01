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

func TestBaseRecordSearchCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/search" {
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
		if payload["app_token"] != "app_1" {
			t.Fatalf("unexpected app_token: %#v", payload["app_token"])
		}
		if payload["table_id"] != "tbl_1" {
			t.Fatalf("unexpected table_id: %#v", payload["table_id"])
		}
		if payload["view_id"] != "viw_1" {
			t.Fatalf("unexpected view_id: %#v", payload["view_id"])
		}
		if payload["automatic_fields"] != true {
			t.Fatalf("unexpected automatic_fields: %#v", payload["automatic_fields"])
		}
		fieldNames, ok := payload["field_names"].([]any)
		if !ok || len(fieldNames) != 2 {
			t.Fatalf("unexpected field_names: %#v", payload["field_names"])
		}
		if fieldNames[0] != "City" || fieldNames[1] != "TempC" {
			t.Fatalf("unexpected field_names order: %#v", fieldNames)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"record_id":          "rec_1",
						"created_time":       "1700000000",
						"last_modified_time": "1700000001",
						"fields": map[string]any{
							"City":  "Seattle",
							"TempC": 6.0,
						},
					},
				},
				"has_more":   false,
				"page_token": "",
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
	cmd.SetArgs([]string{"record", "search", "tbl_1", "--app-token", "app_1", "--view-id", "viw_1", "--fields", "City,TempC"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record search error: %v", err)
	}
	if !strings.Contains(buf.String(), "rec_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Seattle") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
