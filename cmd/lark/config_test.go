package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"lark/internal/config"
	"lark/internal/output"
)

func TestConfigGetJSONOutputValid(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "https://open.feishu.cn",
			DefaultMailboxID:           "mbx_1",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: 1700000000,
			UserAccessToken:            "user-token",
			RefreshToken:               "refresh-token",
			UserAccessTokenExpiresAt:   1700000100,
		},
		Printer: output.Printer{Writer: &buf, JSON: true},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"get"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config get error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("expected valid JSON, got %q", buf.String())
	}

	var got config.Config
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}
	if got.AppID != state.Config.AppID {
		t.Fatalf("expected app_id %s, got %s", state.Config.AppID, got.AppID)
	}
	if got.BaseURL != state.Config.BaseURL {
		t.Fatalf("expected base_url %s, got %s", state.Config.BaseURL, got.BaseURL)
	}
	if got.TenantAccessToken != state.Config.TenantAccessToken {
		t.Fatalf("expected tenant_access_token %s, got %s", state.Config.TenantAccessToken, got.TenantAccessToken)
	}
	if got.UserAccessToken != state.Config.UserAccessToken {
		t.Fatalf("expected user_access_token %s, got %s", state.Config.UserAccessToken, got.UserAccessToken)
	}
	if got.RefreshToken != state.Config.RefreshToken {
		t.Fatalf("expected refresh_token %s, got %s", state.Config.RefreshToken, got.RefreshToken)
	}
}

func TestConfigGetHumanOutputRedactsTokens(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "https://open.feishu.cn",
			DefaultMailboxID:           "mbx_1",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: 1700000000,
			UserAccessToken:            "user-token",
			RefreshToken:               "refresh-token",
			UserAccessTokenExpiresAt:   1700000100,
		},
		Printer: output.Printer{Writer: &buf},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"get"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config get error: %v", err)
	}

	output := buf.String()
	for _, token := range []string{
		state.Config.TenantAccessToken,
		state.Config.UserAccessToken,
		state.Config.RefreshToken,
	} {
		if token == "" {
			continue
		}
		if strings.Contains(output, token) {
			t.Fatalf("expected token %q to be redacted, got %q", token, output)
		}
	}
}
