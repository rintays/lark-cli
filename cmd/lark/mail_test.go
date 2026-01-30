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

func TestMailFoldersCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/folders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
						"folder_id":   "fld_1",
						"name":        "Inbox",
						"folder_type": "INBOX",
					},
				},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"folders", "--mailbox-id", "mbx_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail folders error: %v", err)
	}

	if !strings.Contains(buf.String(), "fld_1\tInbox\tINBOX") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailFoldersCommandRequiresSDK(t *testing.T) {
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
		Client:  &larkapi.Client{},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"folders", "--mailbox-id", "mbx_1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailListCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		if r.URL.Query().Get("folder_id") != "fld_1" {
			t.Fatalf("unexpected folder_id: %s", r.URL.Query().Get("folder_id"))
		}
		if r.URL.Query().Get("only_unread") != "true" {
			t.Fatalf("unexpected only_unread: %s", r.URL.Query().Get("only_unread"))
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
						"message_id": "msg_1",
						"subject":    "Hello",
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"list",
		"--mailbox-id", "mbx_1",
		"--folder-id", "fld_1",
		"--limit", "2",
		"--only-unread",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail list error: %v", err)
	}

	if !strings.Contains(buf.String(), "msg_1\tHello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailGetCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/msg_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"message": map[string]any{
					"message_id": "msg_1",
					"subject":    "Hello",
				},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"get", "msg_1", "--mailbox-id", "mbx_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail get error: %v", err)
	}

	if !strings.Contains(buf.String(), "msg_1\tHello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailListCommandRequiresSDK(t *testing.T) {
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
		Client:  &larkapi.Client{},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"list", "--mailbox-id", "mbx_1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailGetCommandRequiresSDK(t *testing.T) {
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
		Client:  &larkapi.Client{},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"get", "msg_1", "--mailbox-id", "mbx_1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailSendCommandRequiresSDK(t *testing.T) {
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
		Client:  &larkapi.Client{},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
		"--user-access-token", "user-token",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailSendCommandRequiresUserAccessToken(t *testing.T) {
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
		Client:  &larkapi.Client{},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "mail send requires a user access token; pass --user-access-token or set LARK_USER_ACCESS_TOKEN" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailSendCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["subject"] != "Hello" {
			t.Fatalf("unexpected subject: %#v", payload["subject"])
		}
		if payload["body_plain_text"] != "hi" {
			t.Fatalf("unexpected body_plain_text: %#v", payload["body_plain_text"])
		}
		to, ok := payload["to"].([]any)
		if !ok || len(to) != 1 {
			t.Fatalf("unexpected to: %#v", payload["to"])
		}
		addr, ok := to[0].(map[string]any)
		if !ok || addr["mail_address"] != "a@example.com" {
			t.Fatalf("unexpected to address: %#v", to[0])
		}
		cc, ok := payload["cc"].([]any)
		if !ok || len(cc) != 1 {
			t.Fatalf("unexpected cc: %#v", payload["cc"])
		}
		headFrom, ok := payload["head_from"].(map[string]any)
		if !ok || headFrom["name"] != "Lark Bot" {
			t.Fatalf("unexpected head_from: %#v", payload["head_from"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"message_id": "msg_1",
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
			TenantAccessToken:          "tenant-token",
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

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--cc", "cc@example.com",
		"--text", "hi",
		"--from-name", "Lark Bot",
		"--user-access-token", "user-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
	if !strings.Contains(buf.String(), "message_id: msg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
