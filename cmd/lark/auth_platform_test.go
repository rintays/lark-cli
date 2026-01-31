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

func TestAuthPlatformSetFeishuPersistsBaseURL(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config: &config.Config{
			BaseURL: "https://open.feishu.cn",
		},
		Printer: output.Printer{Writer: io.Discard},
	}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"platform", "set", "feishu"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth platform set error: %v", err)
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

func TestAuthPlatformSetLarkPersistsBaseURL(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config: &config.Config{
			BaseURL: "https://open.feishu.cn",
		},
		Printer: output.Printer{Writer: io.Discard},
	}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"platform", "set", "lark"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth platform set error: %v", err)
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

func TestAuthPlatformInfoKnownPlatform(t *testing.T) {
	tests := []struct {
		name         string
		baseURL      string
		wantPlatform string
	}{
		{
			name:         "feishu",
			baseURL:      "https://open.feishu.cn",
			wantPlatform: "feishu",
		},
		{
			name:         "lark",
			baseURL:      "https://open.larksuite.com",
			wantPlatform: "lark",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			state := &appState{
				Config:         &config.Config{BaseURL: tt.baseURL},
				baseURLPersist: tt.baseURL,
				Printer:        output.Printer{Writer: &buf},
			}

			cmd := newAuthCmd(state)
			cmd.SetArgs([]string{"platform", "info"})

			if err := cmd.Execute(); err != nil {
				t.Fatalf("auth platform info error: %v", err)
			}

			got := buf.String()
			if !strings.Contains(got, "platform: "+tt.wantPlatform) {
				t.Fatalf("expected platform %s in output, got %q", tt.wantPlatform, got)
			}
			if !strings.Contains(got, "base_url: "+tt.baseURL) {
				t.Fatalf("expected base_url %s in output, got %q", tt.baseURL, got)
			}
		})
	}
}

func TestAuthPlatformInfoCustomPlatform(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config:         &config.Config{BaseURL: "https://example.com"},
		baseURLPersist: "https://example.com",
		Printer:        output.Printer{Writer: &buf},
	}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"platform", "info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth platform info error: %v", err)
	}

	got := buf.String()
	if !strings.Contains(got, "platform: custom") {
		t.Fatalf("expected custom platform in output, got %q", got)
	}
	if !strings.Contains(got, "base_url: https://example.com") {
		t.Fatalf("expected base_url in output, got %q", got)
	}
}
