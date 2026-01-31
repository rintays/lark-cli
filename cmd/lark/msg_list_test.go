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

func TestMsgListCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/im/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("container_id_type") != "chat" {
			t.Fatalf("unexpected container_id_type: %s", r.URL.Query().Get("container_id_type"))
		}
		if r.URL.Query().Get("container_id") != "oc_123" {
			t.Fatalf("unexpected container_id: %s", r.URL.Query().Get("container_id"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"message_id":  "m1",
						"msg_type":    "text",
						"chat_id":     "oc_123",
						"create_time": "123",
						"body":        map[string]any{"content": "{\"text\":\"hi\"}"},
					},
				},
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

	cmd := newMsgCmd(state)
	cmd.SetArgs([]string{"list", "--container-id", "oc_123", "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg list error: %v", err)
	}
	if !strings.Contains(buf.String(), "m1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
