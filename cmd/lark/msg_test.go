package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larkapi"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestMsgSendCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/im/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("receive_id_type") != "open_id" {
			t.Fatalf("unexpected receive_id_type: %s", r.URL.Query().Get("receive_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["receive_id"] != "ou_123" {
			t.Fatalf("unexpected receive_id: %s", payload["receive_id"])
		}
		if payload["msg_type"] != "text" {
			t.Fatalf("unexpected msg_type: %s", payload["msg_type"])
		}
		if payload["content"] != "{\"text\":\"hello\"}" {
			t.Fatalf("unexpected content: %s", payload["content"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{"message_id": "m1"},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMsgCmd(state)
	cmd.SetArgs([]string{"send", "--receive-id", "ou_123", "--receive-id-type", "open_id", "--text", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg send error: %v", err)
	}

	if !strings.Contains(buf.String(), "message_id: m1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMsgSendCommandFallbackToAPI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/im/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("receive_id_type") != "chat_id" {
			t.Fatalf("unexpected receive_id_type: %s", r.URL.Query().Get("receive_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["receive_id"] != "chat123" {
			t.Fatalf("unexpected receive_id: %s", payload["receive_id"])
		}
		if payload["msg_type"] != "text" {
			t.Fatalf("unexpected msg_type: %s", payload["msg_type"])
		}
		if payload["content"] != "{\"text\":\"hello\"}" {
			t.Fatalf("unexpected content: %s", payload["content"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{"message_id": "m2"},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
		SDK:     &larksdk.Client{},
	}

	cmd := newMsgCmd(state)
	cmd.SetArgs([]string{"send", "--receive-id", "chat123", "--text", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg send error: %v", err)
	}

	if !strings.Contains(buf.String(), "message_id: m2") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
