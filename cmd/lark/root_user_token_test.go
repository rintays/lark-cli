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
		AppID:                    "app",
		AppSecret:                "secret",
		BaseURL:                  baseURL,
		UserAccessToken:          "cached-user",
		UserAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		RefreshToken:             "refresh-token",
	}
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
	sawAppToken := false
	sawRefresh := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/auth/v3/app_access_token/internal":
			sawAppToken = true
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["app_id"] != "app" || payload["app_secret"] != "secret" {
				t.Fatalf("unexpected credentials: %v", payload)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":             0,
				"msg":              "ok",
				"app_access_token": "app-token",
				"expire":           3600,
			})
		case "/open-apis/authen/v1/refresh_access_token":
			sawRefresh = true
			if authHeader := r.Header.Get("Authorization"); authHeader != "Bearer app-token" {
				t.Fatalf("unexpected authorization: %s", authHeader)
			}
			var payload map[string]string
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode refresh payload: %v", err)
			}
			if payload["grant_type"] != "refresh_token" {
				t.Fatalf("unexpected grant_type: %s", payload["grant_type"])
			}
			if payload["refresh_token"] != "refresh-me" {
				t.Fatalf("unexpected refresh_token: %s", payload["refresh_token"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"access_token":  "new-user-token",
					"expires_in":    3600,
					"refresh_token": "new-refresh-token",
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	configPath := filepath.Join(t.TempDir(), "config.json")
	initialCreatedAt := int64(1700000000)
	cfg := &config.Config{
		AppID:                    "app",
		AppSecret:                "secret",
		BaseURL:                  baseURL,
		UserAccessToken:          "stale",
		UserAccessTokenExpiresAt: time.Now().Add(10 * time.Second).Unix(),
		RefreshToken:             "legacy-refresh",
		UserRefreshTokenPayload: &config.UserRefreshTokenPayload{
			RefreshToken: "refresh-me",
			Services:     []string{"drive"},
			Scopes:       "offline_access drive:drive",
			CreatedAt:    initialCreatedAt,
		},
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
	if !sawAppToken || !sawRefresh {
		t.Fatalf("expected refresh flow, app token: %v refresh: %v", sawAppToken, sawRefresh)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.UserAccessToken != "new-user-token" {
		t.Fatalf("expected token saved, got %s", saved.UserAccessToken)
	}
	if saved.RefreshToken != "new-refresh-token" {
		t.Fatalf("expected refresh token saved, got %s", saved.RefreshToken)
	}
	if saved.UserRefreshTokenPayload == nil {
		t.Fatalf("expected refresh token payload saved")
	}
	if saved.UserRefreshTokenPayload.RefreshToken != "new-refresh-token" {
		t.Fatalf("expected refresh token payload updated, got %s", saved.UserRefreshTokenPayload.RefreshToken)
	}
	if saved.UserRefreshTokenPayload.CreatedAt <= initialCreatedAt {
		t.Fatalf("expected refresh token payload created_at updated, got %d", saved.UserRefreshTokenPayload.CreatedAt)
	}
	if saved.UserAccessTokenExpiresAt == 0 {
		t.Fatalf("expected expiry saved")
	}
}

func TestEnsureUserTokenMissingRefreshTokenClears(t *testing.T) {
	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &config.Config{
		AppID:                    "app",
		AppSecret:                "secret",
		BaseURL:                  "https://open.feishu.cn",
		UserAccessToken:          "stale",
		UserAccessTokenExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
		RefreshToken:             "",
	}
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
	if saved.UserAccessToken != "" {
		t.Fatalf("expected user access token cleared, got %s", saved.UserAccessToken)
	}
	if saved.RefreshToken != "" {
		t.Fatalf("expected refresh token cleared, got %s", saved.RefreshToken)
	}
	if saved.UserRefreshTokenPayload != nil {
		t.Fatalf("expected refresh token payload cleared")
	}
	if saved.UserAccessTokenExpiresAt != 0 {
		t.Fatalf("expected expiry cleared, got %d", saved.UserAccessTokenExpiresAt)
	}
}

func TestEnsureUserTokenRefreshFailureClears(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/auth/v3/app_access_token/internal":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code":             0,
				"msg":              "ok",
				"app_access_token": "app-token",
				"expire":           3600,
			})
		case "/open-apis/authen/v1/refresh_access_token":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 999,
				"msg":  "invalid refresh_token",
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	configPath := filepath.Join(t.TempDir(), "config.json")
	cfg := &config.Config{
		AppID:                    "app",
		AppSecret:                "secret",
		BaseURL:                  baseURL,
		UserAccessToken:          "stale",
		UserAccessTokenExpiresAt: time.Now().Add(-1 * time.Minute).Unix(),
		RefreshToken:             "refresh-me",
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
	if saved.UserAccessToken != "" {
		t.Fatalf("expected user access token cleared, got %s", saved.UserAccessToken)
	}
	if saved.RefreshToken != "" {
		t.Fatalf("expected refresh token cleared, got %s", saved.RefreshToken)
	}
	if saved.UserRefreshTokenPayload != nil {
		t.Fatalf("expected refresh token payload cleared")
	}
	if saved.UserAccessTokenExpiresAt != 0 {
		t.Fatalf("expected expiry cleared, got %d", saved.UserAccessTokenExpiresAt)
	}
}
