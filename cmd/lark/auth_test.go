package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/output"
)

func TestAuthLoginWritesConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config: &config.Config{
			BaseURL:                    "https://open.feishu.cn",
			TenantAccessToken:          "cached",
			TenantAccessTokenExpiresAt: 123,
		},
		Printer: output.Printer{Writer: io.Discard},
	}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{
		"login",
		"--app-id", "app",
		"--app-secret", "secret",
		"--base-url", "https://example.com",
	})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth login error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.AppID != "app" {
		t.Fatalf("expected app_id saved, got %s", saved.AppID)
	}
	if saved.AppSecret != "secret" {
		t.Fatalf("expected app_secret saved, got %s", saved.AppSecret)
	}
	if saved.BaseURL != "https://example.com" {
		t.Fatalf("expected base_url saved, got %s", saved.BaseURL)
	}
	if saved.TenantAccessToken != "cached" {
		t.Fatalf("expected token preserved, got %s", saved.TenantAccessToken)
	}
	if saved.TenantAccessTokenExpiresAt != 123 {
		t.Fatalf("expected token expiry preserved, got %d", saved.TenantAccessTokenExpiresAt)
	}
}

func TestAuthPrintsTokenAndExpiry(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			TenantAccessToken:          "cached",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"tenant"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "tenant_access_token: cached") {
		t.Fatalf("expected token in output, got %q", got)
	}
	if !strings.Contains(got, "expires_at:") {
		t.Fatalf("expected expires_at in output, got %q", got)
	}
}
