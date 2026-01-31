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

func TestChatsAnnouncementGetCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/im/v1/chats/oc_1/announcement" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{"revision": "12", "content": "{\"text\":\"hi\"}"},
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
	cmd.SetArgs([]string{"announcement", "get", "oc_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("announcement get error: %v", err)
	}
	if !strings.Contains(buf.String(), "12") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestChatsAnnouncementGetCommandDocxFallbackWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/im/v1/chats/oc_1/announcement":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 125,
				"msg":  "Unable to operate docx type chat announcement.",
			})
		case "/open-apis/docx/v1/chats/oc_1/announcement":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"revision_id":       7,
					"announcement_type": "docx",
				},
			})
		case "/open-apis/docx/v1/chats/oc_1/announcement/blocks":
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if r.URL.Query().Get("revision_id") != "7" {
				t.Fatalf("unexpected revision_id: %s", r.URL.Query().Get("revision_id"))
			}
			if r.URL.Query().Get("page_size") != "200" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
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
						{"block_id": "b1", "block_type": 2},
					},
					"has_more": false,
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
	cmd.SetArgs([]string{"announcement", "get", "oc_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("announcement get error: %v", err)
	}
	if !strings.Contains(buf.String(), "docx") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestChatsAnnouncementUpdateCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/im/v1/chats/oc_1/announcement" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["revision"] != "12" {
			t.Fatalf("unexpected revision: %v", payload["revision"])
		}
		requests := payload["requests"].([]any)
		if len(requests) != 1 {
			t.Fatalf("unexpected requests: %v", payload["requests"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{},
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
	cmd.SetArgs([]string{"announcement", "update", "--chat-id", "oc_1", "--revision", "12", "--request", "{\"requestType\":\"InsertBlocksRequestType\"}"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("announcement update error: %v", err)
	}
	if !strings.Contains(buf.String(), "oc_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
