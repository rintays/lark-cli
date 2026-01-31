package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"lark/internal/testutil"
)

func TestMailSendIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	to := testutil.RequireEnv(t, "LARK_TEST_MAIL_TO")
	fx := getIntegrationFixtures(t)

	subject := integrationFixturePrefix + "mail-" + time.Now().Format("20060102-150405")

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "mail", "send", "--to", to, "--subject", subject, "--text", "ping"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("mail send failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	id, _ := payload["message_id"].(string)
	if id == "" {
		t.Fatalf("expected non-empty message_id, got: %v", payload)
	}
}
