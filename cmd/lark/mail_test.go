package main

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
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
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"id":          "fld_1",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
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

func TestMailFoldersCommandUsesDefaultMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_default/folders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"id":          "fld_1",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"folders"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail folders error: %v", err)
	}

	if !strings.Contains(buf.String(), "fld_1\tInbox\tINBOX") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailFoldersCommandDefaultsToMeMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/me/folders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"id":          "fld_1",
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
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"folders"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail folders error: %v", err)
	}

	if !strings.Contains(buf.String(), "fld_1\tInbox\tINBOX") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailPublicMailboxesListCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/public_mailboxes" {
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
						"public_mailbox_id": "mbx_1",
						"name":              "Public",
						"email":             "public@example.com",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"public-mailboxes", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail public-mailboxes list error: %v", err)
	}

	if !strings.Contains(buf.String(), "mbx_1\tPublic\tpublic@example.com") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailMailboxesListAliasCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/public_mailboxes" {
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
						"public_mailbox_id": "mbx_1",
						"name":              "Public",
						"email":             "public@example.com",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"mailboxes", "list"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail mailboxes list error: %v", err)
	}

	if !strings.Contains(buf.String(), "mbx_1\tPublic\tpublic@example.com") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailMailboxInfoCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"mailbox": map[string]any{
					"mailbox_id":    "mbx_1",
					"name":          "Primary",
					"primary_email": "dev@example.com",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"mailbox", "info", "--mailbox-id", "mbx_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail mailbox info error: %v", err)
	}

	if !strings.Contains(buf.String(), "mailbox_id\tmbx_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "name\tPrimary") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "primary_email\tdev@example.com") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailMailboxSetCommandPersistsDefaultMailbox(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: &bytes.Buffer{}},
	}

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"mailbox", "set", "--mailbox-id", "mbx_123"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail mailbox set error: %v", err)
	}

	loaded, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	if loaded.DefaultMailboxID != "mbx_123" {
		t.Fatalf("expected default mailbox id %q, got %q", "mbx_123", loaded.DefaultMailboxID)
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
	var listCalls int
	var getCalls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/open-apis/mail/v1/user_mailboxes/mbx_1/messages":
			listCalls++
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("folder_id") != "fld_1" {
				t.Fatalf("unexpected folder_id: %s", r.URL.Query().Get("folder_id"))
			}
			if r.URL.Query().Get("only_unread") != "true" {
				t.Fatalf("unexpected only_unread: %s", r.URL.Query().Get("only_unread"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":      []string{"msg_1", "msg_2"},
					"has_more":   false,
					"page_token": "",
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/msg_1":
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"message": map[string]any{
						"message_id":    "msg_1",
						"subject":       "Hello",
						"internal_date": "2026-01-30T12:00:00Z",
						"from": map[string]any{
							"mail_address": "alice@example.com",
							"name":         "Alice",
						},
					},
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/msg_2":
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"message": map[string]any{
						"message_id":    "msg_2",
						"subject":       "World",
						"internal_date": "2026-01-30T13:00:00Z",
						"from": map[string]any{
							"mail_address": "bob@example.com",
							"name":         "Bob",
						},
					},
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
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

	if !strings.Contains(buf.String(), "msg_1\tHello\tAlice <alice@example.com>\t2026-01-30T12:00:00Z") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "msg_2\tWorld\tBob <bob@example.com>\t2026-01-30T13:00:00Z") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if listCalls != 1 {
		t.Fatalf("expected 1 list call, got %d", listCalls)
	}
	if getCalls != 2 {
		t.Fatalf("expected 2 get calls, got %d", getCalls)
	}
}

func TestMailListCommandUsesDefaultMailboxID(t *testing.T) {
	var folderCalls int
	var listCalls int
	var getCalls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/open-apis/mail/v1/user_mailboxes/mbx_default/folders":
			folderCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"id":          "fld_inbox",
							"name":        "Inbox",
							"folder_type": "INBOX",
						},
					},
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/mbx_default/messages":
			listCalls++
			if r.URL.Query().Get("page_size") != "1" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("folder_id") != "fld_inbox" {
				t.Fatalf("unexpected folder_id: %s", r.URL.Query().Get("folder_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":      []string{"msg_1"},
					"has_more":   false,
					"page_token": "",
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/mbx_default/messages/msg_1":
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"message": map[string]any{
						"message_id":    "msg_1",
						"subject":       "Hello",
						"internal_date": "2026-01-30T12:00:00Z",
						"from": map[string]any{
							"mail_address": "alice@example.com",
							"name":         "Alice",
						},
					},
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"list", "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail list error: %v", err)
	}

	if !strings.Contains(buf.String(), "msg_1\tHello\tAlice <alice@example.com>\t2026-01-30T12:00:00Z") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if listCalls != 1 {
		t.Fatalf("expected 1 list call, got %d", listCalls)
	}
	if getCalls != 1 {
		t.Fatalf("expected 1 get call, got %d", getCalls)
	}
	if folderCalls != 1 {
		t.Fatalf("expected 1 folder call, got %d", folderCalls)
	}
}

func TestMailListCommandDefaultsToMeMailboxID(t *testing.T) {
	var folderCalls int
	var listCalls int
	var getCalls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch r.URL.Path {
		case "/open-apis/mail/v1/user_mailboxes/me/folders":
			folderCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"id":          "fld_inbox",
							"name":        "Inbox",
							"folder_type": "INBOX",
						},
					},
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/me/messages":
			listCalls++
			if r.URL.Query().Get("page_size") != "1" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("folder_id") != "fld_inbox" {
				t.Fatalf("unexpected folder_id: %s", r.URL.Query().Get("folder_id"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":      []string{"msg_1"},
					"has_more":   false,
					"page_token": "",
				},
			})
		case "/open-apis/mail/v1/user_mailboxes/me/messages/msg_1":
			getCalls++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"message": map[string]any{
						"message_id":    "msg_1",
						"subject":       "Hello",
						"internal_date": "2026-01-30T12:00:00Z",
						"from": map[string]any{
							"mail_address": "alice@example.com",
							"name":         "Alice",
						},
					},
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
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"list", "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail list error: %v", err)
	}

	if !strings.Contains(buf.String(), "msg_1\tHello\tAlice <alice@example.com>\t2026-01-30T12:00:00Z") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if listCalls != 1 {
		t.Fatalf("expected 1 list call, got %d", listCalls)
	}
	if getCalls != 1 {
		t.Fatalf("expected 1 get call, got %d", getCalls)
	}
	if folderCalls != 1 {
		t.Fatalf("expected 1 folder call, got %d", folderCalls)
	}
}

func TestMailListCommandLimitMustBePositiveDoesNotCallHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call: %s", r.URL.Path)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"list", "--limit", "0"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "limit must be greater than 0" {
		t.Fatalf("unexpected error: %v", err)
	}
}
func TestMailInfoCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/msg_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
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
					"raw":        "RAW_CONTENT",
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"info", "msg_1", "--mailbox-id", "mbx_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail info error: %v", err)
	}

	if !strings.Contains(buf.String(), "message_id\tmsg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "subject\tHello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if strings.Contains(buf.String(), "RAW_CONTENT") {
		t.Fatalf("expected metadata-only output, got: %q", buf.String())
	}
}

func TestMailInfoCommandUsesDefaultMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_default/messages/msg_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"info", "msg_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail info error: %v", err)
	}

	if !strings.Contains(buf.String(), "message_id\tmsg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "subject\tHello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailInfoCommandDefaultsToMeMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/me/messages/msg_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
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
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"info", "msg_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail info error: %v", err)
	}

	if !strings.Contains(buf.String(), "message_id\tmsg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "subject\tHello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
func TestMailInfoCommandRequiresMessageID(t *testing.T) {
	cmd := newMailCmd(&appState{})
	cmd.SetArgs([]string{"info", "--mailbox-id", "mbx_1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"message-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
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
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"message": map[string]any{
					"message_id":      "msg_1",
					"subject":         "Hello",
					"raw":             "RAW_CONTENT",
					"body_plain_text": "Hello world",
					"body_html":       "<p>Hello</p>",
					"smtp_message_id": "smtp_1",
					"internal_date":   "2026-01-31T12:00:00Z",
					"message_state":   1,
					"folder_id":       "fld_1",
					"snippet":         "Hello world",
					"thread_id":       "th_1",
					"head_from":       map[string]any{"mail_address": "alice@example.com", "name": "Alice"},
					"to":              []map[string]any{{"mail_address": "bob@example.com", "name": "Bob"}},
					"attachments":     []map[string]any{{"id": "att_1", "filename": "a.txt", "body": "YXQ=", "attachment_type": 1}},
					"bcc":             []map[string]any{{"mail_address": "bcc@example.com", "name": "Bcc"}},
					"cc":              []map[string]any{{"mail_address": "cc@example.com", "name": "Cc"}},
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
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

	if !strings.Contains(buf.String(), "message_id\tmsg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "raw\tRAW_CONTENT") {
		t.Fatalf("expected raw output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "body_plain_text\tHello world") {
		t.Fatalf("expected body output: %q", buf.String())
	}
}

func TestMailGetCommandRequiresMessageID(t *testing.T) {
	cmd := newMailCmd(&appState{})
	cmd.SetArgs([]string{"get", "--mailbox-id", "mbx_1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"message-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailGetCommandMissingMessageIDDoesNotCallHTTP(t *testing.T) {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"get", "--mailbox-id", "mbx_1"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 0 {
		t.Fatalf("unexpected request count: %d", calls)
	}
}

func TestMailInfoCommandMissingMessageIDDoesNotCallHTTP(t *testing.T) {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"info", "--mailbox-id", "mbx_1"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if calls != 0 {
		t.Fatalf("unexpected request count: %d", calls)
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

func TestMailInfoCommandRequiresSDK(t *testing.T) {
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

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{"info", "msg_1", "--mailbox-id", "mbx_1"})
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
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "http://example.com",
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	t.Setenv("LARK_USER_ACCESS_TOKEN", "")

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
	if !strings.Contains(err.Error(), userOAuthReloginCommand) {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailSendMissingSubjectDoesNotCallHTTP(t *testing.T) {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
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
		"--to", "a@example.com",
		"--text", "hi",
		"--user-access-token", "user-token",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "subject is required" {
		t.Fatalf("unexpected error: %v", err)
	}
	if calls != 0 {
		t.Fatalf("unexpected request count: %d", calls)
	}
}

func TestMailSendCommandRawRejectsText(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
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
		"--raw", "Zm9v",
		"--text", "hi",
		"--user-access-token", "user-token",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "raw is mutually exclusive with subject/to/cc/bcc/text/html" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMailSendCommandWithRawPayload(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["raw"] != "Zm9v" {
			t.Fatalf("unexpected raw: %#v", payload["raw"])
		}
		if _, ok := payload["subject"]; ok {
			t.Fatalf("unexpected subject: %#v", payload["subject"])
		}
		if _, ok := payload["to"]; ok {
			t.Fatalf("unexpected to: %#v", payload["to"])
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
		"--raw", "Zm9v",
		"--user-access-token", "user-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
	if !strings.Contains(buf.String(), "message_id: msg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailSendCommandWithRawFile(t *testing.T) {
	rawData := []byte("Subject: Hello\r\n\r\nBody")
	rawPath := filepath.Join(t.TempDir(), "hello.eml")
	if err := os.WriteFile(rawPath, rawData, 0o644); err != nil {
		t.Fatalf("write raw file: %v", err)
	}
	expected := base64.URLEncoding.EncodeToString(rawData)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["raw"] != expected {
			t.Fatalf("unexpected raw: %#v", payload["raw"])
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
		"--raw-file", rawPath,
		"--user-access-token", "user-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
	if !strings.Contains(buf.String(), "message_id: msg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func setupMailSendState(t *testing.T, expectedToken string) (*appState, *bytes.Buffer) {
	t.Helper()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != fmt.Sprintf("Bearer %s", expectedToken) {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	return state, &buf
}

func TestMailSendCommandUsesFlagToken(t *testing.T) {
	state, _ := setupMailSendState(t, "flag-token")
	state.Config.DefaultMailboxID = "mbx_default"
	state.Config.UserAccessToken = "cached-token"
	state.Config.UserAccessTokenExpiresAt = time.Now().Add(2 * time.Hour).Unix()
	t.Setenv("LARK_USER_ACCESS_TOKEN", "env-token")

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
		"--user-access-token", "flag-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
}

func TestMailSendCommandUsesEnvToken(t *testing.T) {
	state, _ := setupMailSendState(t, "env-token")
	state.Config.UserAccessToken = "cached-token"
	state.Config.UserAccessTokenExpiresAt = time.Now().Add(2 * time.Hour).Unix()
	t.Setenv("LARK_USER_ACCESS_TOKEN", "env-token")

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
}

func TestMailSendCommandUsesCachedUserToken(t *testing.T) {
	state, _ := setupMailSendState(t, "cached-token")
	state.Config.UserAccessToken = "cached-token"
	state.Config.UserAccessTokenExpiresAt = time.Now().Add(2 * time.Hour).Unix()
	t.Setenv("LARK_USER_ACCESS_TOKEN", "")

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--mailbox-id", "mbx_1",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
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

func TestMailSendDefaultsToConfigMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_default/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
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
			DefaultMailboxID:           "mbx_default",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
		"--user-access-token", "user-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
	if !strings.Contains(buf.String(), "message_id: msg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMailSendDefaultsToMeMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/me/messages/send" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailCmd(state)
	cmd.SetArgs([]string{
		"send",
		"--subject", "Hello",
		"--to", "a@example.com",
		"--text", "hi",
		"--user-access-token", "user-token",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send error: %v", err)
	}
	if !strings.Contains(buf.String(), "message_id: msg_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
