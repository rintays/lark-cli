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

func TestLoadUsesEnvFallbackForKeyringBackendAutoWhenMissing(t *testing.T) {
	t.Setenv("LARK_KEYRING_BACKEND", "auto")

	path := filepath.Join(t.TempDir(), "missing.json")
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KeyringBackend != "auto" {
		t.Fatalf("expected env keyring backend auto, got %q", cfg.KeyringBackend)
	}
}

func TestLoadKeepsKeyringBackendAutoWhenSetInConfig(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	data := []byte(`{"keyring_backend":"auto"}`)
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if cfg.KeyringBackend != "auto" {
		t.Fatalf("expected keyring_backend auto, got %q", cfg.KeyringBackend)
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

func TestUserRefreshTokenPayloadRoundTrip(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{
		AppID: "app-id",
		UserAccounts: map[string]*UserAccount{
			"default": {
				RefreshToken: "legacy-refresh",
				UserRefreshTokenPayload: &UserRefreshTokenPayload{
					RefreshToken: "payload-refresh",
					Services:     []string{"drive", "docs"},
					Scopes:       "offline_access drive:drive",
					CreatedAt:    1700000100,
				},
			},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	acct := loaded.UserAccounts["default"]
	if acct == nil {
		t.Fatalf("expected account loaded")
	}
	if acct.UserRefreshTokenPayload == nil {
		t.Fatalf("expected refresh token payload")
	}
	if acct.UserRefreshTokenPayload.RefreshToken != "payload-refresh" {
		t.Fatalf("expected payload refresh token, got %q", acct.UserRefreshTokenPayload.RefreshToken)
	}
	if acct.UserRefreshTokenPayload.CreatedAt != 1700000100 {
		t.Fatalf("expected payload created_at, got %d", acct.UserRefreshTokenPayload.CreatedAt)
	}
	if acct.RefreshTokenValue() != "payload-refresh" {
		t.Fatalf("expected payload refresh token precedence, got %q", acct.RefreshTokenValue())
	}
}

func TestUserRefreshTokenFallbackToLegacy(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.json")
	cfg := &Config{
		AppID: "app-id",
		UserAccounts: map[string]*UserAccount{
			"default": {
				RefreshToken: "legacy-refresh",
			},
		},
	}
	if err := Save(path, cfg); err != nil {
		t.Fatalf("save config: %v", err)
	}
	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}
	acct := loaded.UserAccounts["default"]
	if acct == nil {
		t.Fatalf("expected account loaded")
	}
	if acct.UserRefreshTokenPayload != nil {
		t.Fatalf("expected no payload")
	}
	if acct.RefreshTokenValue() != "legacy-refresh" {
		t.Fatalf("expected legacy refresh token, got %q", acct.RefreshTokenValue())
	}
}
