package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/zalando/go-keyring"
)

func hydrateAppSecretFromKeyring(state *appState) error {
	if state == nil || state.Config == nil {
		return nil
	}
	if !state.Config.AppSecretInKeyring {
		return nil
	}
	if strings.TrimSpace(state.Config.AppSecret) != "" {
		return nil
	}
	secret, ok, err := loadAppSecretFromKeyring(state)
	if err != nil {
		return err
	}
	if ok {
		state.Config.AppSecret = secret
	}
	return nil
}

func ensureAppSecret(state *appState) error {
	if state == nil || state.Config == nil {
		return errors.New("config is required")
	}
	if strings.TrimSpace(state.Config.AppSecret) != "" {
		return nil
	}
	if !state.Config.AppSecretInKeyring {
		return errors.New("app_id and app_secret must be set in config")
	}
	secret, ok, err := loadAppSecretFromKeyring(state)
	if err != nil {
		return err
	}
	if !ok || strings.TrimSpace(secret) == "" {
		return errors.New("app_id and app_secret must be set in config")
	}
	state.Config.AppSecret = secret
	return nil
}

func persistAppSecret(state *appState, secret string, storeInKeyring bool) error {
	if state == nil || state.Config == nil {
		return errors.New("config is required")
	}
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("app-secret must not be empty")
	}
	if storeInKeyring {
		if err := saveAppSecretToKeyring(state, secret); err != nil {
			return err
		}
		state.Config.AppSecret = ""
		state.Config.AppSecretInKeyring = true
		return nil
	}
	state.Config.AppSecret = secret
	state.Config.AppSecretInKeyring = false
	_ = deleteAppSecretFromKeyring(state)
	return nil
}

func resolveAppSecretStorage(state *appState, storeInKeyringFlag, storeInConfigFlag bool) (bool, error) {
	if storeInKeyringFlag && storeInConfigFlag {
		return false, errors.New("store-secret-in-keyring and store-secret-in-config are mutually exclusive")
	}
	if storeInKeyringFlag {
		return true, nil
	}
	if storeInConfigFlag {
		return false, nil
	}
	if state != nil && state.Config != nil && state.Config.AppSecretInKeyring {
		return true, nil
	}
	return false, nil
}

func loadAppSecretFromKeyring(state *appState) (string, bool, error) {
	name, err := appSecretKeyringUsername(state)
	if err != nil {
		return "", false, err
	}
	value, err := keyring.Get(keyringServiceName, name)
	if err == nil {
		return value, true, nil
	}
	if errors.Is(err, keyring.ErrNotFound) {
		return "", false, nil
	}
	if errors.Is(err, keyring.ErrUnsupportedPlatform) {
		return "", false, errors.New("keychain backend is not supported on this platform; use keyring_backend=file or store app secret in config")
	}
	return "", false, err
}

func saveAppSecretToKeyring(state *appState, secret string) error {
	name, err := appSecretKeyringUsername(state)
	if err != nil {
		return err
	}
	if err := keyring.Set(keyringServiceName, name, secret); err != nil {
		if errors.Is(err, keyring.ErrUnsupportedPlatform) {
			return errors.New("keychain backend is not supported on this platform; use keyring_backend=file or store app secret in config")
		}
		return err
	}
	return nil
}

func deleteAppSecretFromKeyring(state *appState) error {
	name, err := appSecretKeyringUsername(state)
	if err != nil {
		return err
	}
	if err := keyring.Delete(keyringServiceName, name); err != nil {
		if errors.Is(err, keyring.ErrNotFound) {
			return nil
		}
		if errors.Is(err, keyring.ErrUnsupportedPlatform) {
			return nil
		}
		return err
	}
	return nil
}

func appSecretKeyringUsername(state *appState) (string, error) {
	if state == nil || state.Config == nil {
		return "", errors.New("config is required")
	}
	if strings.TrimSpace(state.Config.AppID) == "" {
		return "", errors.New("app_id is required to store app secret in keychain")
	}
	bucket := userTokenBucketID(state)
	if strings.TrimSpace(bucket) == "" {
		return "", errors.New("app_id is required to store app secret in keychain")
	}
	return fmt.Sprintf("%s:app-secret", bucket), nil
}
