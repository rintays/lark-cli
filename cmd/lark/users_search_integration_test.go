package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"lark/internal/testutil"
)

func TestUsersSearchIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	email := os.Getenv("LARK_TEST_USER_EMAIL")
	if email == "" {
		t.Skip("LARK_TEST_USER_EMAIL not set")
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "users", "search", "--email", email})

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
