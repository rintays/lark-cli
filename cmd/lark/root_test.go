package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larkapi"
	"lark/internal/testutil"
)

func TestEnsureTenantTokenUsesCache(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		w.WriteHeader(http.StatusInternalServerError)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		ConfigPath: filepath.Join(t.TempDir(), "config.json"),
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "cached",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Client: &larkapi.Client{BaseURL: baseURL, AppID: "app", AppSecret: "secret", HTTPClient: httpClient},
	}

	token, err := ensureTenantToken(context.Background(), state)
	if err != nil {
		t.Fatalf("ensureTenantToken error: %v", err)
	}
	if token != "cached" {
		t.Fatalf("expected cached token, got %s", token)
	}
	if called {
		t.Fatalf("expected cached token without API call")
	}
}

func TestEnsureTenantTokenRefreshesAndSaves(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":                0,
			"msg":                 "ok",
			"tenant_access_token": "fresh",
			"expire":              3600,
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	configPath := filepath.Join(t.TempDir(), "config.json")
	state := &appState{
		ConfigPath: configPath,
		Config: &config.Config{
			AppID:     "app",
			AppSecret: "secret",
			BaseURL:   baseURL,
		},
		Client: &larkapi.Client{BaseURL: baseURL, AppID: "app", AppSecret: "secret", HTTPClient: httpClient},
	}

	_, err := ensureTenantToken(context.Background(), state)
	if err != nil {
		t.Fatalf("ensureTenantToken error: %v", err)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("read config: %v", err)
	}
	var saved config.Config
	if err := json.Unmarshal(data, &saved); err != nil {
		t.Fatalf("unmarshal config: %v", err)
	}
	if saved.TenantAccessToken != "fresh" {
		t.Fatalf("expected token saved, got %s", saved.TenantAccessToken)
	}
	if saved.TenantAccessTokenExpiresAt == 0 {
		t.Fatalf("expected expiry saved")
	}
}

func TestRootHelpShowsMeetingsCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}

	foundMeetings := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "meetings" || strings.HasPrefix(trimmed, "meetings ") || strings.HasPrefix(trimmed, "meetings\t") {
			foundMeetings = true
		}
		if trimmed == "meeting" || strings.HasPrefix(trimmed, "meeting ") || strings.HasPrefix(trimmed, "meeting\t") {
			t.Fatalf("unexpected meeting command in help output: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundMeetings {
		t.Fatalf("expected meetings command in help output, got:\n%s", buf.String())
	}
}
