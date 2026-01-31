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

func TestBaseAppCopyCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1/copy" {
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
		if payload["name"] != "CopyApp" {
			t.Fatalf("unexpected app name: %#v", payload["name"])
		}
		if payload["folder_token"] != "fld_1" {
			t.Fatalf("unexpected folder token: %#v", payload["folder_token"])
		}
		if payload["without_content"] != true {
			t.Fatalf("unexpected without_content: %#v", payload["without_content"])
		}
		if payload["time_zone"] != "Asia/Shanghai" {
			t.Fatalf("unexpected time zone: %#v", payload["time_zone"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"app": map[string]any{
					"app_token": "app_copy",
					"name":      "CopyApp",
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
	cmd.SetArgs([]string{
		"app", "copy",
		"--app-token", "app_1",
		"--name", "CopyApp",
		"--folder-token", "fld_1",
		"--without-content",
		"--time-zone", "Asia/Shanghai",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base app copy error: %v", err)
	}
	if !strings.Contains(buf.String(), "app_copy\tCopyApp") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
