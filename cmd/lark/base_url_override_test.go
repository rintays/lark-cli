package main

import (
	"testing"

	"lark/internal/config"
)

func TestApplyBaseURLOverrides_BaseURLWins(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://open.feishu.cn"}
	state := &appState{Platform: "lark", BaseURL: "https://example.com/open-apis/"}

	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
	if cfg.BaseURL != "https://example.com" {
		t.Fatalf("expected base_url=https://example.com, got %q", cfg.BaseURL)
	}
}

func TestApplyBaseURLOverrides_PlatformWinsOverConfig(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://open.feishu.cn"}
	state := &appState{Platform: "lark"}

	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
	if cfg.BaseURL != "https://open.larkoffice.com" {
		t.Fatalf("expected base_url=https://open.larkoffice.com, got %q", cfg.BaseURL)
	}
}

func TestApplyBaseURLOverrides_NormalizeConfig(t *testing.T) {
	cfg := &config.Config{BaseURL: "https://open.feishu.cn/open-apis/"}
	state := &appState{}

	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("apply overrides: %v", err)
	}
	if cfg.BaseURL != "https://open.feishu.cn" {
		t.Fatalf("expected base_url=https://open.feishu.cn, got %q", cfg.BaseURL)
	}
}
