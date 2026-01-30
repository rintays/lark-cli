package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadUsesEnvFallbackWhenConfigMissing(t *testing.T) {
	t.Setenv("LARK_APP_ID", "env-app-id")
	t.Setenv("LARK_APP_SECRET", "env-app-secret")

	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AppID != "env-app-id" {
		t.Fatalf("expected env app id, got %q", cfg.AppID)
	}
	if cfg.AppSecret != "env-app-secret" {
		t.Fatalf("expected env app secret, got %q", cfg.AppSecret)
	}
}

func TestLoadConfigTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("LARK_APP_ID", "env-app-id")
	t.Setenv("LARK_APP_SECRET", "env-app-secret")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"app_id":"file-app-id","app_secret":"file-app-secret"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.AppID != "file-app-id" {
		t.Fatalf("expected file app id, got %q", cfg.AppID)
	}
	if cfg.AppSecret != "file-app-secret" {
		t.Fatalf("expected file app secret, got %q", cfg.AppSecret)
	}
}
