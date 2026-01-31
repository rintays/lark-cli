package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"lark/internal/config"

	"github.com/zalando/go-keyring"
)

const keyringServiceName = "lark-cli"

type userToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    int64  `json:"expires_at"`
	Scope        string `json:"scope"`
}

func userTokenBackend(cfg *config.Config) string {
	if cfg == nil {
		return "file"
	}
	backend := strings.ToLower(strings.TrimSpace(cfg.KeyringBackend))
	if backend == "" || backend == "auto" {
		return "file"
	}
	return backend
}

func loadUserToken(state *appState, account string) (userToken, bool, error) {
	if state == nil || state.Config == nil {
		return userToken{}, false, errors.New("config is required")
	}
	backend := userTokenBackend(state.Config)
	switch backend {
	case "keychain":
		return loadUserTokenFromKeyring(state, account)
	case "file":
		return loadUserTokenFromConfig(state.Config, account)
	default:
		return userToken{}, false, fmt.Errorf("unsupported keyring backend %q", backend)
	}
}

func storeUserToken(state *appState, account string, token userToken) error {
	if state == nil || state.Config == nil {
		return errors.New("config is required")
	}
	backend := userTokenBackend(state.Config)
	switch backend {
	case "keychain":
		if err := saveKeyringToken(state, account, token); err != nil {
			return err
		}
		acct := ensureUserAccount(state.Config, account)
		acct.UserAccessToken = ""
		acct.RefreshToken = ""
		acct.UserAccessTokenExpiresAt = token.ExpiresAt
		if token.Scope != "" {
			acct.UserAccessTokenScope = token.Scope
		}
		saveUserAccount(state.Config, account, acct)
		return nil
	case "file":
		saveUserTokenToConfig(state.Config, account, token)
		return nil
	default:
		return fmt.Errorf("unsupported keyring backend %q", backend)
	}
}

func clearUserToken(state *appState, account string) error {
	if state == nil || state.Config == nil {
		return errors.New("config is required")
	}
	backend := userTokenBackend(state.Config)
	switch backend {
	case "keychain":
		if err := deleteKeyringToken(state, account); err != nil {
			return err
		}
		clearUserAccountTokens(state.Config, account)
		return nil
	case "file":
		clearUserAccountTokens(state.Config, account)
		return nil
	default:
		return fmt.Errorf("unsupported keyring backend %q", backend)
	}
}

func loadUserTokenFromConfig(cfg *config.Config, account string) (userToken, bool, error) {
	acct, ok := loadUserAccount(cfg, account)
	if !ok {
		return userToken{}, false, nil
	}
	if acct.UserAccessToken == "" && acct.RefreshToken == "" && acct.UserAccessTokenExpiresAt == 0 && acct.UserAccessTokenScope == "" {
		return userToken{}, false, nil
	}
	return userToken{
		AccessToken:  acct.UserAccessToken,
		RefreshToken: acct.RefreshToken,
		ExpiresAt:    acct.UserAccessTokenExpiresAt,
		Scope:        acct.UserAccessTokenScope,
	}, true, nil
}

func saveUserTokenToConfig(cfg *config.Config, account string, token userToken) {
	acct := ensureUserAccount(cfg, account)
	acct.UserAccessToken = token.AccessToken
	acct.RefreshToken = token.RefreshToken
	acct.UserAccessTokenExpiresAt = token.ExpiresAt
	if token.Scope != "" {
		acct.UserAccessTokenScope = token.Scope
	}
	saveUserAccount(cfg, account, acct)
}

func loadUserTokenFromKeyring(state *appState, account string) (userToken, bool, error) {
	stored, ok, err := getKeyringToken(state, account)
	if err != nil {
		return userToken{}, false, err
	}
	if ok {
		return stored, true, nil
	}

	// Migrate any file-stored token to keyring on first read.
	fileToken, fileOK, err := loadUserTokenFromConfig(state.Config, account)
	if err != nil {
		return userToken{}, false, err
	}
	if !fileOK {
		return userToken{}, false, nil
	}
	if err := saveKeyringToken(state, account, fileToken); err != nil {
		return userToken{}, false, err
	}
	clearUserAccountTokens(state.Config, account)
	acct := ensureUserAccount(state.Config, account)
	acct.UserAccessTokenExpiresAt = fileToken.ExpiresAt
	if fileToken.Scope != "" {
		acct.UserAccessTokenScope = fileToken.Scope
	}
	saveUserAccount(state.Config, account, acct)
	if err := state.saveConfig(); err != nil {
		return userToken{}, false, err
	}
	return fileToken, true, nil
}

func getKeyringToken(state *appState, account string) (userToken, bool, error) {
	if state == nil {
		return userToken{}, false, errors.New("state is required")
	}
	name := keyringUsername(state, account)
	value, err := keyring.Get(keyringServiceName, name)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			// Backward compat: tokens used to be keyed by config path only.
			legacy := legacyKeyringUsername(state.ConfigPath, account)
			legacyValue, legacyErr := keyring.Get(keyringServiceName, legacy)
			if legacyErr != nil {
				if errors.Is(legacyErr, keyring.ErrNotFound) {
					return userToken{}, false, nil
				}
				if errors.Is(legacyErr, keyring.ErrUnsupportedPlatform) {
					return userToken{}, false, errors.New("keychain backend is not supported on this platform; use keyring_backend=file")
				}
				return userToken{}, false, legacyErr
			}
			var token userToken
			if err := json.Unmarshal([]byte(legacyValue), &token); err != nil {
				return userToken{}, false, fmt.Errorf("invalid keyring token data: %w", err)
			}
			// Best-effort migration to the new bucketed key.
			_ = saveKeyringToken(state, account, token)
			_ = deleteLegacyKeyringToken(state.ConfigPath, account)
			return token, true, nil
		}
		if errors.Is(err, keyring.ErrUnsupportedPlatform) {
			return userToken{}, false, errors.New("keychain backend is not supported on this platform; use keyring_backend=file")
		}
		return userToken{}, false, err
	}
	var token userToken
	if err := json.Unmarshal([]byte(value), &token); err != nil {
		return userToken{}, false, fmt.Errorf("invalid keyring token data: %w", err)
	}
	return token, true, nil
}

func saveKeyringToken(state *appState, account string, token userToken) error {
	if state == nil {
		return errors.New("state is required")
	}
	payload, err := json.Marshal(token)
	if err != nil {
		return err
	}
	name := keyringUsername(state, account)
	if err := keyring.Set(keyringServiceName, name, string(payload)); err != nil {
		if errors.Is(err, keyring.ErrUnsupportedPlatform) {
			return errors.New("keychain backend is not supported on this platform; use keyring_backend=file")
		}
		return err
	}
	return nil
}

func deleteKeyringToken(state *appState, account string) error {
	if state == nil {
		return errors.New("state is required")
	}
	// Delete both new and legacy keys so `auth user accounts remove` does not
	// leave stale credentials behind.
	name := keyringUsername(state, account)
	err := keyring.Delete(keyringServiceName, name)
	if err != nil {
		if !errors.Is(err, keyring.ErrNotFound) {
			if errors.Is(err, keyring.ErrUnsupportedPlatform) {
				return errors.New("keychain backend is not supported on this platform; use keyring_backend=file")
			}
			return err
		}
	}
	legacyErr := deleteLegacyKeyringToken(state.ConfigPath, account)
	if legacyErr != nil {
		if errors.Is(legacyErr, keyring.ErrUnsupportedPlatform) {
			return errors.New("keychain backend is not supported on this platform; use keyring_backend=file")
		}
		return legacyErr
	}
	return nil
}

func deleteLegacyKeyringToken(configPath, account string) error {
	legacy := legacyKeyringUsername(configPath, account)
	err := keyring.Delete(keyringServiceName, legacy)
	if err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		return err
	}
	return nil
}

// keyringUsername returns a stable keyring username for the user token.
//
// It includes a "client bucket" derived from (profile=config path, base_url,
// app_id) so that refresh tokens do not get mixed across apps or platforms.
func keyringUsername(state *appState, account string) string {
	if account == "" {
		account = defaultUserAccountName
	}
	bucket := userTokenBucketID(state)
	return fmt.Sprintf("%s:%s", bucket, account)
}

func legacyKeyringUsername(configPath, account string) string {
	if account == "" {
		account = defaultUserAccountName
	}
	hash := sha256.Sum256([]byte(configPath))
	return fmt.Sprintf("%s:%s", hex.EncodeToString(hash[:]), account)
}

func userTokenBucketID(state *appState) string {
	if state == nil {
		return ""
	}
	baseURL := ""
	appID := ""
	if state.Config != nil {
		baseURL = strings.ToLower(normalizeBaseURL(state.Config.BaseURL))
		appID = strings.TrimSpace(state.Config.AppID)
	}
	seed := strings.Join([]string{state.ConfigPath, baseURL, appID}, "\n")
	hash := sha256.Sum256([]byte(seed))
	return hex.EncodeToString(hash[:])
}
