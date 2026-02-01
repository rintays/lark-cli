package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lark/internal/config"
)

func TestRootProfileFlagSelectsProfileConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	defaultPath, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}
	devPath, err := config.DefaultPathForProfile("dev")
	if err != nil {
		t.Fatalf("profile path: %v", err)
	}

	if err := os.MkdirAll(filepath.Dir(defaultPath), 0o700); err != nil {
		t.Fatalf("mkdir default dir: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(devPath), 0o700); err != nil {
		t.Fatalf("mkdir dev dir: %v", err)
	}

	if err := os.WriteFile(defaultPath, []byte(`{"base_url":"https://default.example.com"}`), 0o600); err != nil {
		t.Fatalf("write default config: %v", err)
	}
	if err := os.WriteFile(devPath, []byte(`{"base_url":"https://dev.example.com"}`), 0o600); err != nil {
		t.Fatalf("write dev config: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "--profile", "dev", "config", "info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var got config.Config
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v (out=%q)", err, buf.String())
	}
	if got.BaseURL != "https://dev.example.com" {
		t.Fatalf("expected dev base_url, got %q", got.BaseURL)
	}
}

func TestRootDefaultProfileAliasUsesLegacyDefaultConfigPath(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	defaultPath, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("default path: %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(defaultPath), 0o700); err != nil {
		t.Fatalf("mkdir default dir: %v", err)
	}
	if err := os.WriteFile(defaultPath, []byte(`{"base_url":"https://default.example.com"}`), 0o600); err != nil {
		t.Fatalf("write default config: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "--profile", "default", "config", "info"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute: %v", err)
	}

	var got config.Config
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v (out=%q)", err, buf.String())
	}
	if got.BaseURL != "https://default.example.com" {
		t.Fatalf("expected legacy default base_url, got %q", got.BaseURL)
	}
}
