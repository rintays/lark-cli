package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AppID              string `json:"app_id"`
	AppSecret          string `json:"app_secret"`
	BaseURL            string `json:"base_url"`
	DefaultMailboxID   string `json:"default_mailbox_id"`
	DefaultTokenType   string `json:"default_token_type"`
	DefaultUserAccount string `json:"default_user_account,omitempty"`

	// KeyringBackend controls where OAuth tokens are stored.
	// Supported values: auto|file|keychain.
	//
	// NOTE: only `file` is implemented today; other values are validated but may
	// error at runtime.
	KeyringBackend string `json:"keyring_backend,omitempty"`

	UserScopes                 []string `json:"user_scopes,omitempty"`
	TenantAccessToken          string   `json:"tenant_access_token"`
	TenantAccessTokenExpiresAt int64    `json:"tenant_access_token_expires_at"`
	UserAccessToken            string   `json:"user_access_token"`
	UserAccessTokenScope       string   `json:"user_access_token_scope"`
	RefreshToken               string   `json:"refresh_token"`
	UserAccessTokenExpiresAt   int64    `json:"user_access_token_expires_at"`

	UserAccounts map[string]*UserAccount `json:"user_accounts,omitempty"`
}

type UserAccount struct {
	UserAccessToken          string   `json:"user_access_token,omitempty"`
	UserAccessTokenScope     string   `json:"user_access_token_scope,omitempty"`
	RefreshToken             string   `json:"refresh_token,omitempty"`
	UserAccessTokenExpiresAt int64    `json:"user_access_token_expires_at,omitempty"`
	UserScopes               []string `json:"user_scopes,omitempty"`
}

func Default() *Config {
	return &Config{
		BaseURL:            "https://open.feishu.cn",
		DefaultTokenType:   "tenant",
		DefaultUserAccount: "default",
	}
}

func Load(path string) (*Config, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			applyEnvFallback(cfg)
			normalizeDefaults(cfg)
			return cfg, nil
		}
		return nil, err
	}
	if err := json.Unmarshal(data, cfg); err != nil {
		return nil, err
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://open.feishu.cn"
	}
	applyEnvFallback(cfg)
	normalizeDefaults(cfg)
	return cfg, nil
}

func Save(path string, cfg *Config) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0o600)
}

func DefaultPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lark", "config.json"), nil
}

func DefaultPathForProfile(profile string) (string, error) {
	name := strings.TrimSpace(profile)
	if name == "" {
		return "", errors.New("profile is required")
	}
	if strings.Contains(name, "..") || strings.Contains(name, "/") || strings.Contains(name, "\\") {
		return "", fmt.Errorf("invalid profile %q", profile)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "lark", "profiles", name, "config.json"), nil
}

func applyEnvFallback(cfg *Config) {
	if cfg.AppID == "" {
		if appID := os.Getenv("LARK_APP_ID"); appID != "" {
			cfg.AppID = appID
		}
	}
	if cfg.AppSecret == "" {
		if appSecret := os.Getenv("LARK_APP_SECRET"); appSecret != "" {
			cfg.AppSecret = appSecret
		}
	}
	if cfg.KeyringBackend == "" {
		if v := os.Getenv("LARK_KEYRING_BACKEND"); v != "" {
			cfg.KeyringBackend = v
		}
	}
}

func normalizeDefaults(cfg *Config) {
	if cfg == nil {
		return
	}
	cfg.DefaultUserAccount = strings.TrimSpace(cfg.DefaultUserAccount)
	if cfg.DefaultUserAccount == "" {
		cfg.DefaultUserAccount = "default"
	}
	switch strings.ToLower(strings.TrimSpace(cfg.DefaultTokenType)) {
	case "tenant", "user":
		cfg.DefaultTokenType = strings.ToLower(strings.TrimSpace(cfg.DefaultTokenType))
	default:
		cfg.DefaultTokenType = "tenant"
	}

	switch strings.ToLower(strings.TrimSpace(cfg.KeyringBackend)) {
	case "", "auto":
		// Keep behavior backward-compatible: default to file storage.
		cfg.KeyringBackend = "file"
	case "file", "keychain":
		cfg.KeyringBackend = strings.ToLower(strings.TrimSpace(cfg.KeyringBackend))
	default:
		// Invalid values should not silently change behavior.
		// Keep it as-is (so callers can surface an error if needed), but ensure we
		// don't end up with empty.
		if strings.TrimSpace(cfg.KeyringBackend) == "" {
			cfg.KeyringBackend = "file"
		}
	}
}
