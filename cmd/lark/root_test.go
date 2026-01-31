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

	"github.com/spf13/cobra"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

func TestCanonicalCommandPath(t *testing.T) {
	if got := canonicalCommandPath(nil); got != "" {
		t.Fatalf("canonicalCommandPath(nil)=%q, want empty", got)
	}

	{
		root := &cobra.Command{Use: "lark"}
		if got := canonicalCommandPath(root); got != "" {
			t.Fatalf("canonicalCommandPath(root)=%q, want empty", got)
		}

		mail := &cobra.Command{Use: "mail"}
		send := &cobra.Command{Use: "send"}
		root.AddCommand(mail)
		mail.AddCommand(send)

		if got := canonicalCommandPath(send); got != "mail send" {
			t.Fatalf("canonicalCommandPath(send)=%q, want %q", got, "mail send")
		}
	}

	{
		root := &cobra.Command{Use: "foo"}
		mail := &cobra.Command{Use: "mail"}
		send := &cobra.Command{Use: "send"}
		root.AddCommand(mail)
		mail.AddCommand(send)

		if got := canonicalCommandPath(send); got != "mail send" {
			t.Fatalf("canonicalCommandPath(send with non-lark root)=%q, want %q", got, "mail send")
		}
	}
}

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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

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

func TestEnsureTenantTokenRefreshesWithSDK(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["app_id"] != "app" || payload["app_secret"] != "secret" {
			t.Fatalf("unexpected credentials: %v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	_, err = ensureTenantToken(context.Background(), state)
	if err != nil {
		t.Fatalf("ensureTenantToken error: %v", err)
	}
	if !called {
		t.Fatalf("expected SDK tenant token request")
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

func TestRootHelpShowsMinutesCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}

	foundMinutes := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "minutes" || strings.HasPrefix(trimmed, "minutes ") || strings.HasPrefix(trimmed, "minutes\t") {
			foundMinutes = true
		}
		if trimmed == "minute" || strings.HasPrefix(trimmed, "minute ") || strings.HasPrefix(trimmed, "minute\t") {
			t.Fatalf("unexpected minute command in help output: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundMinutes {
		t.Fatalf("expected minutes command in help output, got:\n%s", buf.String())
	}
}

func TestRootHelpShowsCalendarsCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}

	foundCalendars := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "calendars" || strings.HasPrefix(trimmed, "calendars ") || strings.HasPrefix(trimmed, "calendars\t") {
			foundCalendars = true
		}
		if trimmed == "calendar" || strings.HasPrefix(trimmed, "calendar ") || strings.HasPrefix(trimmed, "calendar\t") {
			t.Fatalf("unexpected calendar command in help output: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundCalendars {
		t.Fatalf("expected calendars command in help output, got:\n%s", buf.String())
	}
}

func TestRootHelpShowsMessagesCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}

	foundMessages := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "messages" || strings.HasPrefix(trimmed, "messages ") || strings.HasPrefix(trimmed, "messages\t") {
			foundMessages = true
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundMessages {
		t.Fatalf("expected messages command in help output, got:\n%s", buf.String())
	}
}

func TestRootMsgAliasWorks(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"msg", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}
	if !strings.Contains(buf.String(), "Send chat messages") {
		t.Fatalf("unexpected help output: %q", buf.String())
	}
}

func TestRootCalendarAliasWorks(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"calendar", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}
	if !strings.Contains(buf.String(), "Usage:\n  lark calendars [command]") {
		t.Fatalf("expected canonical calendars usage in help output, got: %q", buf.String())
	}
	// Ensure the canonical command name is used in the usage line.
	// Note: "lark calendars" contains the substring "lark calendar", so we must match a trailing space.
	if strings.Contains(buf.String(), "Usage:\n  lark calendar ") {
		t.Fatalf("unexpected singular calendar usage in help output: %q", buf.String())
	}
}

func TestRootHelpShowsBasesCommand(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}

	foundBases := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "bases" || strings.HasPrefix(trimmed, "bases ") || strings.HasPrefix(trimmed, "bases\t") {
			foundBases = true
		}
		if trimmed == "base" || strings.HasPrefix(trimmed, "base ") || strings.HasPrefix(trimmed, "base\t") {
			t.Fatalf("unexpected base command in help output: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundBases {
		t.Fatalf("expected bases command in help output, got:\n%s", buf.String())
	}
}

func TestRootBaseAliasWorks(t *testing.T) {
	cmd := newRootCmd()
	cmd.PersistentPreRunE = nil

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"base", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("help error: %v", err)
	}
	if !strings.Contains(buf.String(), "Manage Bitable bases") {
		t.Fatalf("unexpected help output: %q", buf.String())
	}
}
