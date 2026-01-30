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
