package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestMailMailboxInfoDefaultsMailboxID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"code": 0, "msg": "ok", "data": map[string]any{"mailbox": map[string]any{"mailbox_id": "mbx_1"}}})
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
			DefaultMailboxID:           "mbx_1",
		},
		Printer: output.Printer{Writer: &buf, JSON: true},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newMailMailboxCmd(state)
	cmd.SetArgs([]string{"info"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("mailbox info error: %v", err)
	}
}
