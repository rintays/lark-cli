package main

import (
	"bytes"
	"encoding/json"
	"testing"

	"lark/internal/testutil"
)

func TestWhoamiIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "whoami"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("whoami failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if _, ok := payload["tenant_key"]; !ok {
		// whoami payload may evolve; require at least one stable key.
		// If tenant_key is absent, keep the test strict by failing.
		t.Fatalf("expected tenant_key in output, got: %v", payload)
	}
}
