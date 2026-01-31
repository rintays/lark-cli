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

func TestMsgSearchCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/search/v2/message":
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.Header.Get("Authorization") != "Bearer u-token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["query"] != "hello" {
				t.Fatalf("unexpected query: %v", payload["query"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":    []string{"m1", "m2"},
					"has_more": false,
				},
			})
			return
		case r.Method == http.MethodGet && strings.HasPrefix(r.URL.Path, "/open-apis/im/v1/messages/"):
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			if r.Header.Get("Authorization") != "Bearer u-token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			messageID := strings.TrimPrefix(r.URL.Path, "/open-apis/im/v1/messages/")
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"message_id":  messageID,
							"msg_type":    "text",
							"create_time": "123",
							"sender": map[string]any{
								"id":          "ou_1",
								"id_type":     "open_id",
								"sender_type": "user",
							},
							"body": map[string]any{
								"content": "{\"text\":\"hello " + messageID + "\"}",
							},
						},
					},
				},
			})
			return
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
		}
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                    "app",
			AppSecret:                "secret",
			BaseURL:                  baseURL,
			UserAccessToken:          "u-token",
			UserAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMsgCmd(state)
	cmd.SetArgs([]string{"search", "hello", "--limit", "2", "--page-size", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg search error: %v", err)
	}
	outputText := buf.String()
	if !strings.Contains(outputText, "id: m1") {
		t.Fatalf("unexpected output: %q", outputText)
	}
	if !strings.Contains(outputText, "from: user:open_id:ou_1") {
		t.Fatalf("unexpected output: %q", outputText)
	}
	if !strings.Contains(outputText, "hello m1") {
		t.Fatalf("unexpected output: %q", outputText)
	}
}
