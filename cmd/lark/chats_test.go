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

func TestChatsListCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/open-apis/im/v1/chats" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":    []map[string]any{{"chat_id": "c1", "name": "Chat One"}},
					"has_more": false,
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

		cmd := newChatsCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("chats list error: %v", err)
		}

		if !strings.Contains(buf.String(), "c1\tChat One") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires sdk client", func(t *testing.T) {
		state := &appState{
			Config: &config.Config{
				AppID:                      "app",
				AppSecret:                  "secret",
				BaseURL:                    "http://example.com",
				TenantAccessToken:          "token",
				TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			},
			Printer: output.Printer{Writer: &bytes.Buffer{}},
		}

		cmd := newChatsCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "sdk client is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
