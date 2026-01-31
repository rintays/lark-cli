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

func TestLoadConfigTakesPrecedenceOverEnvForAppIDSecret(t *testing.T) {
	t.Setenv("LARK_APP_ID", "env-app-id")
	t.Setenv("LARK_APP_SECRET", "env-app-secret")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"app_id":"file-app-id","app_secret":"file-app-secret","keyring_backend":"file"}`)
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
	if cfg.KeyringBackend != "file" {
		t.Fatalf("expected keyring_backend to be file, got %q", cfg.KeyringBackend)
	}
}

func TestLoadEnvOverridesKeyringBackendWhenConfigPresent(t *testing.T) {
	t.Setenv("LARK_KEYRING_BACKEND", "keychain")

	dir := t.TempDir()
	path := filepath.Join(dir, "config.json")
	data := []byte(`{"keyring_backend":"file"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KeyringBackend != "keychain" {
		t.Fatalf("expected env keyring backend, got %q", cfg.KeyringBackend)
	}
}

func TestDefaultPathForProfile(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	path, err := DefaultPathForProfile("dev")
	if err != nil {
		t.Fatalf("DefaultPathForProfile returned error: %v", err)
	}
	expected := filepath.Join(home, ".config", "lark", "profiles", "dev", "config.json")
	if path != expected {
		t.Fatalf("expected %q, got %q", expected, path)
	}

	invalid := []string{"", "   ", "dev/one", "dev\\one", "..", "dev..", "../dev"}
	for _, profile := range invalid {
		if _, err := DefaultPathForProfile(profile); err == nil {
			t.Fatalf("expected error for profile %q", profile)
		}
	}
}

func TestLoadUsesEnvFallbackForKeyringBackendWhenMissing(t *testing.T) {
	t.Setenv("LARK_KEYRING_BACKEND", "keychain")

	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KeyringBackend != "keychain" {
		t.Fatalf("expected env keyring backend, got %q", cfg.KeyringBackend)
	}
}

func TestLoadDefaultsKeyringBackendToFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KeyringBackend != "file" {
		t.Fatalf("expected default keyring backend file, got %q", cfg.KeyringBackend)
	}
}
