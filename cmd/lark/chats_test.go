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
		callCount := 0
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			callCount++
			if r.URL.Path != "/open-apis/im/v1/chats" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			query := r.URL.Query()
			switch callCount {
			case 1:
				if query.Get("page_size") != "3" {
					t.Fatalf("unexpected page_size: %s", query.Get("page_size"))
				}
				if _, ok := query["page_token"]; ok {
					t.Fatalf("unexpected page_token on first request: %q", query.Get("page_token"))
				}
			case 2:
				if query.Get("page_size") != "1" {
					t.Fatalf("unexpected page_size: %s", query.Get("page_size"))
				}
				if query.Get("page_token") != "next" {
					t.Fatalf("unexpected page_token: %s", query.Get("page_token"))
				}
			default:
				t.Fatalf("unexpected call count: %d", callCount)
			}
			w.Header().Set("Content-Type", "application/json")
			if callCount == 1 {
				_ = json.NewEncoder(w).Encode(map[string]any{
					"code": 0,
					"msg":  "ok",
					"data": map[string]any{
						"items":      []map[string]any{{"chat_id": "c1", "name": "Chat One"}, {"chat_id": "c2", "name": "Chat Two"}},
						"has_more":   true,
						"page_token": "next",
					},
				})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":    []map[string]any{{"chat_id": "c3", "name": "Chat Three"}},
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
		cmd.SetArgs([]string{"list", "--limit", "3"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("chats list error: %v", err)
		}

		if callCount != 2 {
			t.Fatalf("expected 2 requests, got %d", callCount)
		}
		if !strings.Contains(buf.String(), "c1\tChat One") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
		if !strings.Contains(buf.String(), "c2\tChat Two") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
		if !strings.Contains(buf.String(), "c3\tChat Three") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires credentials", func(t *testing.T) {
		state := &appState{
			Config: &config.Config{
				AppID:                      "app",
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
		if !strings.Contains(err.Error(), "app_id and app_secret") && !strings.Contains(err.Error(), "missing app credentials") {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
