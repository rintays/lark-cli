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

func TestWikiNodeSearchCommandUsesUserTokenAndEndpoint(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/wiki/v1/nodes/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer usertoken" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["query"] != "hello" {
			t.Fatalf("unexpected query: %+v", payload)
		}
		if payload["space_id"] != "spc1" {
			t.Fatalf("unexpected space_id: %+v", payload)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{"node_token": "n1", "obj_type": "docx", "title": "A", "url": "https://example.com/a"},
					{"node_token": "n2", "obj_type": "sheet", "title": "B", "url": "https://example.com/b"},
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

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"node", "search", "hello", "--space-id", "spc1", "--user-access-token", "usertoken"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("wiki node search error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "n1\tdocx\tA\thttps://example.com/a") {
		t.Fatalf("unexpected output: %q", out)
	}
	if !strings.Contains(out, "n2\tsheet\tB\thttps://example.com/b") {
		t.Fatalf("unexpected output: %q", out)
	}
}
