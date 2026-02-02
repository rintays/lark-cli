package main

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

func TestEnsureUserTokenUsesCache(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	cfg := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   baseURL,
	}
	withUserAccount(cfg, defaultUserAccountName, "cached-user", "refresh-token", time.Now().Add(2*time.Hour).Unix(), "")
	sdkClient, err := larksdk.New(cfg, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state := &appState{
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config:     cfg,
		SDK:        sdkClient,
	}

	token, err := ensureUserToken(context.Background(), state)
	if err != nil {
		t.Fatalf("ensureUserToken error: %v", err)
	}
	if token != "cached-user" {
		t.Fatalf("expected cached token, got %s", token)
	}
	if called {
		t.Fatalf("expected cached token without API call")
	}
}

func TestEnsureUserTokenRefreshesAndSaves(t *testing.T) {
	sawRefresh := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/authen/v2/oauth/token":
			sawRefresh = true
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode refresh payload: %v", err)
			}
			if payload["grant_type"] != "refresh_token" {
				t.Fatalf("unexpected grant_type: %s", payload["grant_type"])
			}
			if payload["client_id"] != "app" || payload["client_secret"] != "secret" {
				t.Fatalf("unexpected credentials: %v", payload)
			}
			if payload["refresh_token"] != "refresh-me" {
				t.Fatalf("unexpected refresh_token: %s", payload["refresh_token"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":          0,
				"access_token":  "new-user-token",
				"expires_in":    3600,
				"refresh_token": "new-refresh-token",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	configPath := filepath.Join(t.TempDir(), "config.json")
	initialCreatedAt := int64(1700000000)
	cfg := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   baseURL,
	}
	withUserAccount(cfg, defaultUserAccountName, "stale", "refresh-me", time.Now().Add(10*time.Second).Unix(), "")
	if acct, ok := cfg.UserAccounts[defaultUserAccountName]; ok {
		acct.UserRefreshTokenPayload = &config.UserRefreshTokenPayload{
			RefreshToken: "refresh-me",
			Services:     []string{"drive"},
			Scopes:       "offline_access drive:drive",
			CreatedAt:    initialCreatedAt,
		}
	}
	sdkClient, err := larksdk.New(cfg, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
		SDK:        sdkClient,
	}

	token, err := ensureUserToken(context.Background(), state)
	if err != nil {
		t.Fatalf("ensureUserToken error: %v", err)
	}
	if token != "new-user-token" {
		t.Fatalf("expected refreshed token, got %s", token)
	}
	if !sawRefresh {
		t.Fatalf("expected refresh flow")
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	account, ok := loadUserAccount(&saved, defaultUserAccountName)
	if !ok {
		t.Fatalf("expected account saved")
	}
	if account.UserAccessToken != "new-user-token" {
		t.Fatalf("expected token saved, got %s", account.UserAccessToken)
	}
	if account.RefreshToken != "new-refresh-token" {
		t.Fatalf("expected refresh token saved, got %s", account.RefreshToken)
	}
	if account.UserAccessTokenExpiresAt == 0 {
		t.Fatalf("expected expiry saved")
	}
	if account.UserRefreshTokenPayload == nil {
		t.Fatalf("expected refresh token payload saved")
	}
	if account.UserRefreshTokenPayload.RefreshToken != "new-refresh-token" {
		t.Fatalf("expected refresh token payload updated, got %s", account.UserRefreshTokenPayload.RefreshToken)
	}
	if account.UserRefreshTokenPayload.CreatedAt <= initialCreatedAt {
		t.Fatalf("expected refresh token payload created_at updated, got %d", account.UserRefreshTokenPayload.CreatedAt)
	}
}

func TestEnsureUserTokenMissingRefreshTokenClears(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   "https://open.feishu.cn",
	}
	withUserAccount(cfg, defaultUserAccountName, "stale", "", time.Now().Add(-1*time.Minute).Unix(), "")
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
	}

	_, err := ensureUserToken(context.Background(), state)
	if err == nil {
		t.Fatalf("expected ensureUserToken error")
	}
	if !strings.Contains(err.Error(), userOAuthReloginCommand) {
		t.Fatalf("expected login instruction, got %v", err)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	account, ok := loadUserAccount(&saved, defaultUserAccountName)
	if ok {
		if account.UserAccessToken != "" {
			t.Fatalf("expected user access token cleared, got %s", account.UserAccessToken)
		}
		if account.RefreshToken != "" {
			t.Fatalf("expected refresh token cleared, got %s", account.RefreshToken)
		}
		if account.UserRefreshTokenPayload != nil {
			t.Fatalf("expected refresh token payload cleared")
		}
		if account.UserAccessTokenExpiresAt != 0 {
			t.Fatalf("expected expiry cleared, got %d", account.UserAccessTokenExpiresAt)
		}
	}
}

func TestEnsureUserTokenRefreshFailureClears(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/authen/v2/oauth/token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":              999,
				"error":             "invalid_request",
				"error_description": "invalid refresh_token",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &config.Config{
		AppID:     "app",
		AppSecret: "secret",
		BaseURL:   baseURL,
	}
	withUserAccount(cfg, defaultUserAccountName, "stale", "refresh-me", time.Now().Add(-1*time.Minute).Unix(), "")
	sdkClient, err := larksdk.New(cfg, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state := &appState{
		ConfigPath: configPath,
		Config:     cfg,
		SDK:        sdkClient,
	}

	_, err = ensureUserToken(context.Background(), state)
	if err == nil {
		t.Fatalf("expected ensureUserToken error")
	}
	if !strings.Contains(err.Error(), "refresh token revoked") {
		t.Fatalf("expected refresh token revoked/expired hint, got %v", err)
	}
	if !strings.Contains(err.Error(), userOAuthReloginCommand) {
		t.Fatalf("expected login instruction, got %v", err)
	}

	data, readErr := os.ReadFile(configPath)
	if readErr != nil {
		t.Fatalf("read config: %v", readErr)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	account, ok := loadUserAccount(&saved, defaultUserAccountName)
	if ok {
		if account.UserAccessToken != "" {
			t.Fatalf("expected user access token cleared, got %s", account.UserAccessToken)
		}
		if account.RefreshToken != "" {
			t.Fatalf("expected refresh token cleared, got %s", account.RefreshToken)
		}
		if account.UserRefreshTokenPayload != nil {
			t.Fatalf("expected refresh token payload cleared")
		}
		if account.UserAccessTokenExpiresAt != 0 {
			t.Fatalf("expected expiry cleared, got %d", account.UserAccessTokenExpiresAt)
		}
	}
}
