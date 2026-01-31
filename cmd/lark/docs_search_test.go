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

func TestDocsSearchCommandUsesDriveSearchEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["query"] != "spec" {
			t.Fatalf("unexpected query: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files": []map[string]any{
					{
						"token":       "d1",
						"name":        "Specs",
						"type":        "docx",
						"url":         "https://example.com/docx",
						"created_at":  0,
						"modified_at": 0,
					},
					{
						"token":       "s1",
						"name":        "Sheet",
						"type":        "sheet",
						"url":         "https://example.com/sheet",
						"created_at":  0,
						"modified_at": 0,
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"search", "--query", "spec", "--limit", "10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs search error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "d1\tSpecs\tdocx\thttps://example.com/docx") {
		t.Fatalf("unexpected output: %q", out)
	}
	if strings.Contains(out, "s1\tSheet\tsheet") {
		t.Fatalf("expected sheet file to be filtered out, got: %q", out)
	}
}
