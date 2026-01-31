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
	cmd.SetArgs([]string{"list", "oc_123", "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg list error: %v", err)
	}
	if !strings.Contains(buf.String(), "m1") {
		t.Fatalf("unexpected output (missing id): %q", buf.String())
	}
	if !strings.Contains(buf.String(), "hi") {
		t.Fatalf("unexpected output (missing content): %q", buf.String())
	}
	metaIndex := strings.Index(buf.String(), "id: m1")
	contentIndex := strings.Index(buf.String(), "hi")
	if metaIndex == -1 || contentIndex == -1 || metaIndex > contentIndex {
		t.Fatalf("unexpected output (meta should appear before content): %q", buf.String())
	}
}

func TestMessageContentForDisplayMentions(t *testing.T) {
	message := larksdk.Message{
		MsgType: "text",
		Body:    larksdk.MessageBody{Content: `{"text":"@_user_1 hello"}`},
		Mentions: []larksdk.MessageMention{
			{Key: "@_user_1", ID: "ou_123", IDType: "user_id"},
		},
	}
	got := messageContentForDisplay(message)
	want := "[@_user_1](user_id://ou_123) hello"
	if got != want {
		t.Fatalf("unexpected mention render: %q", got)
	}
}

func TestMessageContentForDisplayTemplate(t *testing.T) {
	message := larksdk.Message{
		MsgType: "system",
		Body:    larksdk.MessageBody{Content: `{"template":"{from_user} started the group chat.","from_user":["Fred Liang"],"divider_text":{}}`},
	}
	got := messageContentForDisplay(message)
	want := "Fred Liang started the group chat."
	if got != want {
		t.Fatalf("unexpected template render: %q", got)
	}
}
