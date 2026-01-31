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

func TestBaseRecordUpdateCommandRequiresFieldsJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call")
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"record", "update", "--app-token", "app_1", "--table-id", "tbl_1", "--record-id", "rec_1"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "fields-json or field is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBaseRecordUpdateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/tables/tbl_1/records/rec_1" {
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
		fields, ok := payload["fields"].(map[string]any)
		if !ok {
			t.Fatalf("missing fields payload: %#v", payload["fields"])
		}
		if fields["Title"] != "Updated" {
			t.Fatalf("unexpected fields: %#v", fields)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"record": map[string]any{
					"record_id":          "rec_1",
					"fields":             map[string]any{"Title": "Updated"},
					"created_time":       1700000000,
					"last_modified_time": 1700000002,
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
	cmd.SetArgs([]string{"record", "update", "--app-token", "app_1", "--table-id", "tbl_1", "--record-id", "rec_1", "--fields-json", `{"Title":"Updated"}`})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base record update error: %v", err)
	}
	if !strings.Contains(buf.String(), "rec_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
