package config

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	AppID                      string `json:"app_id"`
	AppSecret                  string `json:"app_secret"`
	BaseURL                    string `json:"base_url"`
	DefaultMailboxID           string `json:"default_mailbox_id"`
	DefaultTokenType           string `json:"default_token_type"`
	TenantAccessToken          string `json:"tenant_access_token"`
	TenantAccessTokenExpiresAt int64  `json:"tenant_access_token_expires_at"`
	UserAccessToken            string `json:"user_access_token"`
	UserAccessTokenScope       string `json:"user_access_token_scope"`
	RefreshToken               string `json:"refresh_token"`
	UserAccessTokenExpiresAt   int64  `json:"user_access_token_expires_at"`
}

func Default() *Config {
	return &Config{
		BaseURL:          "https://open.feishu.cn",
		DefaultTokenType: "tenant",
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
}

func normalizeDefaults(cfg *Config) {
	if cfg == nil {
		return
	}
	switch strings.ToLower(strings.TrimSpace(cfg.DefaultTokenType)) {
	case "tenant", "user":
		cfg.DefaultTokenType = strings.ToLower(strings.TrimSpace(cfg.DefaultTokenType))
	default:
		cfg.DefaultTokenType = "tenant"
	}
}
