package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"lark/internal/testutil"
)

func TestUsersSearchIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	email := testutil.RequireEnv(t, "LARK_TEST_USER_EMAIL")
	fx := getIntegrationFixtures(t)

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "users", "search", "--email", email})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("users search failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	users, ok := payload["users"]
	if !ok {
		t.Fatalf("expected users in output, got: %v", payload)
	}
	if _, ok := users.([]any); !ok {
		t.Fatalf("expected users to be an array, got: %T (%v)", users, users)
	}
}
