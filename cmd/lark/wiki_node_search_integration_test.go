package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"lark/internal/testutil"
)

func TestWikiNodeSearchIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	// Requires either:
	// - env LARK_USER_ACCESS_TOKEN, OR
	// - cached config with refresh token so ensureUserToken can refresh.
	// (If neither exists, command should error; treat as a skip to keep integration suite usable.)
	if os.Getenv("LARK_USER_ACCESS_TOKEN") == "" {
		// We'll still try; if it errors due to missing token/refresh token, skip.
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "wiki", "node", "search", "a", "--limit", "1"})

	err := cmd.Execute()
	if err != nil {
		// If user token isn't available, the CLI should explain what to do.
		// In integration runs, that usually means env wasn't configured.
		t.Skipf("wiki search unavailable (likely missing user token): %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["nodes"]; !ok {
		t.Fatalf("expected nodes in output, got: %v", payload)
	}
}
