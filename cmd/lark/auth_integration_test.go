package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"lark/internal/testutil"
)

func TestAuthIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "auth"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	tok, ok := payload["tenant_access_token"].(string)
	if !ok || tok == "" {
		t.Fatalf("expected non-empty tenant_access_token, got: %v", payload)
	}
}
