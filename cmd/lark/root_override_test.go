package main

import (
	"os"
	"path/filepath"
	"testing"

	"lark/internal/config"
)

func TestApplyBaseURLOverridesPrecedence(t *testing.T) {
	tests := []struct {
		name            string
		baseURL         string
		platform        string
		configBaseURL   string
		expectedBaseURL string
	}{
		{
			name:            "base-url wins",
			baseURL:         "https://example.com",
			platform:        "lark",
			configBaseURL:   "https://open.feishu.cn",
			expectedBaseURL: "https://example.com",
		},
		{
			name:            "platform wins when no base-url",
			platform:        "lark",
			configBaseURL:   "https://open.feishu.cn",
			expectedBaseURL: "https://open.larksuite.com",
		},
		{
			name:            "config base-url used when no overrides",
			configBaseURL:   "https://custom.example.com",
			expectedBaseURL: "https://custom.example.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := &config.Config{BaseURL: tt.configBaseURL}
			state := &appState{
				BaseURL:  tt.baseURL,
				Platform: tt.platform,
			}
			if err := applyBaseURLOverrides(state, cfg); err != nil {
				t.Fatalf("applyBaseURLOverrides error: %v", err)
			}
			if cfg.BaseURL != tt.expectedBaseURL {
				t.Fatalf("expected base URL %s, got %s", tt.expectedBaseURL, cfg.BaseURL)
			}
		})
	}
}

func TestApplyBaseURLOverridesRejectsUnsupportedPlatform(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://open.feishu.cn"}
	state := &appState{Platform: "unknown"}
	if err := applyBaseURLOverrides(state, cfg); err == nil {
		t.Fatalf("expected error for unsupported platform")
	}
}

func TestRuntimeBaseURLOverrideDoesNotPersistConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   "https://open.feishu.cn",
	}
	if err := config.Save(configPath, original); err != nil {
		t.Fatalf("save config: %v", err)
	}
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
		BaseURL:    "https://open.larksuite.com",
	}
	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("applyBaseURLOverrides error: %v", err)
	}
	if cfg.BaseURL != "https://open.larksuite.com" {
		t.Fatalf("expected runtime base URL override applied, got %s", cfg.BaseURL)
	}
	if err := state.saveConfig(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if _, err := os.ReadFile(configPath); err != nil {
		t.Fatalf("read config after: %v", err)
	}
	saved, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config after: %v", err)
	}
	if normalizeBaseURL(saved.BaseURL) != normalizeBaseURL(original.BaseURL) {
		t.Fatalf("expected base URL not persisted (want %s, got %s)", original.BaseURL, saved.BaseURL)
	}
	_ = before // keep variable for potential future regression debugging
}

func TestRuntimePlatformOverrideDoesNotPersistConfig(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	original := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   "https://open.feishu.cn",
	}
	if err := config.Save(configPath, original); err != nil {
		t.Fatalf("save config: %v", err)
	}
	before, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config before: %v", err)
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
		Platform:   "lark",
	}
	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("applyBaseURLOverrides error: %v", err)
	}
	if cfg.BaseURL != "https://open.larksuite.com" {
		t.Fatalf("expected runtime platform override applied, got %s", cfg.BaseURL)
	}
	if err := state.saveConfig(); err != nil {
		t.Fatalf("save config: %v", err)
	}
	if _, err := os.ReadFile(configPath); err != nil {
		t.Fatalf("read config after: %v", err)
	}
	saved, err := config.Load(configPath)
	if err != nil {
		t.Fatalf("load config after: %v", err)
	}
	if normalizeBaseURL(saved.BaseURL) != normalizeBaseURL(original.BaseURL) {
		t.Fatalf("expected base URL not persisted (want %s, got %s)", original.BaseURL, saved.BaseURL)
	}
	_ = before // keep variable for potential future regression debugging
}
