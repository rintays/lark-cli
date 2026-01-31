package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestContactsUserInfoCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/ou_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"user": map[string]any{
					"user_id": "u_1",
					"open_id": "ou_1",
					"name":    "Ada",
					"email":   "ada@example.com",
					"mobile":  "+1-555-0100",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newContactsCmd(state)
	cmd.SetArgs([]string{"user", "info", "--open-id", "ou_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("contacts user info error: %v", err)
	}

	if !strings.Contains(buf.String(), "u_1\tAda\tada@example.com\t+1-555-0100") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestContactsUserInfoHelpCommand(t *testing.T) {
	cmd := newContactsCmd(&appState{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"user", "info", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("contacts user info help error: %v", err)
	}

	if !strings.Contains(buf.String(), "Show a contact user by ID") {
		t.Fatalf("unexpected help output: %q", buf.String())
	}
}

func TestContactsHelpPrefersUserCommand(t *testing.T) {
	cmd := newContactsCmd(&appState{})

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("contacts help error: %v", err)
	}

	foundUser := false
	scanner := bufio.NewScanner(strings.NewReader(buf.String()))
	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimLeft(line, " \t")
		if trimmed == "user" || strings.HasPrefix(trimmed, "user ") || strings.HasPrefix(trimmed, "user\t") {
			foundUser = true
		}
		if trimmed == "users" || strings.HasPrefix(trimmed, "users ") || strings.HasPrefix(trimmed, "users\t") {
			t.Fatalf("unexpected users command in help output: %q", line)
		}
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scan help output: %v", err)
	}
	if !foundUser {
		t.Fatalf("expected user command in help output, got:\n%s", buf.String())
	}
}
