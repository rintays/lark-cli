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

func TestConfigInfoJSONOutputValid(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "https://open.feishu.cn",
			DefaultMailboxID:           "mbx_1",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: 1700000000,
			DefaultUserAccount:         defaultUserAccountName,
		},
		Printer: output.Printer{Writer: &buf, JSON: true},
	}
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "refresh-token", 1700000100, "")

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"info"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config info error: %v", err)
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
	account := got.UserAccounts[defaultUserAccountName]
	if account == nil {
		t.Fatalf("expected user account in config output")
	}
	if account.UserAccessToken != "user-token" {
		t.Fatalf("expected user_access_token %s, got %s", "user-token", account.UserAccessToken)
	}
	if account.RefreshToken != "refresh-token" {
		t.Fatalf("expected refresh_token %s, got %s", "refresh-token", account.RefreshToken)
	}
}

func TestConfigInfoHumanOutputRedactsTokens(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    "https://open.feishu.cn",
			DefaultMailboxID:           "mbx_1",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: 1700000000,
			DefaultUserAccount:         defaultUserAccountName,
		},
		Printer: output.Printer{Writer: &buf},
	}
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "refresh-token", 1700000100, "")

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"info"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config info error: %v", err)
	}

	output := buf.String()
	for _, token := range []string{
		state.Config.TenantAccessToken,
		"user-token",
		"refresh-token",
	} {
		if token == "" {
			continue
		}
		if strings.Contains(output, token) {
			t.Fatalf("expected token %q to be redacted, got %q", token, output)
		}
	}
}

func TestConfigListKeysJSONOutput(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config:  config.Default(),
		Printer: output.Printer{Writer: &buf, JSON: true},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"list-keys"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("config list-keys error: %v", err)
	}

	if !json.Valid(buf.Bytes()) {
		t.Fatalf("expected valid JSON, got %q", buf.String())
	}

	var got []configKeyInfo
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal JSON: %v", err)
	}

	keys := make(map[string]struct{}, len(got))
	for _, item := range got {
		keys[item.Key] = struct{}{}
	}

	for _, key := range []string{
		"base-url",
		"platform",
		"default-mailbox-id",
		"default-user-account",
		"user-tokens",
		"app-id",
		"app-secret",
	} {
		if _, ok := keys[key]; !ok {
			t.Fatalf("expected key %q in output", key)
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
	if saved.BaseURL != "https://open.larksuite.com" {
		t.Fatalf("expected base_url saved, got %s", saved.BaseURL)
	}
}

func TestConfigSetAppCredentialsPersistsConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config:     config.Default(),
		Printer:    output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--app-id", "app", "--app-secret", "secret"})
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
	if saved.AppID != "app" {
		t.Fatalf("expected app_id saved, got %s", saved.AppID)
	}
	if saved.AppSecret != "secret" {
		t.Fatalf("expected app_secret saved")
	}
}

func TestConfigSetAppCredentialsMutuallyExclusiveWithBaseURL(t *testing.T) {
	state := &appState{
		Config:  config.Default(),
		Printer: output.Printer{Writer: io.Discard},
	}

	cmd := newConfigCmd(state)
	cmd.SetArgs([]string{"set", "--app-id", "app", "--base-url", "https://open.feishu.cn"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error")
	}
}
