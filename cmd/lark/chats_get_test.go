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

func TestChatsGetCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/open-apis/im/v1/chats/oc_1":
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{"name": "Chat One"},
			})
		case "/open-apis/im/v1/chats/oc_1/members":
			if r.URL.Query().Get("member_id_type") != "open_id" {
				t.Fatalf("unexpected member_id_type: %s", r.URL.Query().Get("member_id_type"))
			}
			if r.URL.Query().Get("page_size") != "20" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{"member_id_type": "open_id", "member_id": "ou_1", "name": "User One"},
					},
					"has_more":     false,
					"member_total": 1,
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
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
	cmd.SetArgs([]string{"get", "oc_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("chats get error: %v", err)
	}
	if !strings.Contains(buf.String(), "oc_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
