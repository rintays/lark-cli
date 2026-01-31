package main

import (
	"bytes"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
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

func TestConfigSetBaseURLPersistsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--base-url", "https://open.feishu.cn"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.BaseURL != "https://open.feishu.cn" {
		t.Fatalf("expected base_url saved, got %s", saved.BaseURL)
	}
}

func TestConfigSetPlatformFeishuPersistsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--platform", "feishu"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.BaseURL != "https://open.feishu.cn" {
		t.Fatalf("expected base_url saved, got %s", saved.BaseURL)
	}
}

func TestConfigSetPlatformLarkPersistsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--platform", "lark"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config set error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.BaseURL != "https://open.larkoffice.com" {
		t.Fatalf("expected base_url saved, got %s", saved.BaseURL)
	}
}

func TestConfigSetPlatformAndBaseURLErrors(t *testing.T) {
	state := &appState{
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--platform", "feishu", "--base-url", "https://open.feishu.cn"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected config set error for platform and base-url")
	}
}

func TestConfigUnsetBaseURLPersistsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := config.Default()
	cfg.BaseURL = "https://open.larkoffice.com"
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"unset", "--base-url"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config unset error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.BaseURL != "" {
		t.Fatalf("expected base_url cleared, got %s", saved.BaseURL)
	}
}

func TestConfigUnsetBaseURLClearsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}
	state.Config.BaseURL = "https://open.feishu.cn"

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"unset", "--base-url"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config unset error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.BaseURL != "" {
		t.Fatalf("expected base_url cleared, got %q", saved.BaseURL)
	}
}

func TestConfigUnsetDefaultMailboxIDClearsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}
	state.Config.DefaultMailboxID = "mbx_123"

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"unset", "--default-mailbox-id"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("config unset error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.DefaultMailboxID != "" {
		t.Fatalf("expected default_mailbox_id cleared, got %q", saved.DefaultMailboxID)
	}
}

func TestConfigUnsetBaseURLAndDefaultMailboxIDErrors(t *testing.T) {
	state := &appState{
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"unset", "--base-url", "--default-mailbox-id"})

	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected config unset error for base-url and default-mailbox-id")
	}
}
