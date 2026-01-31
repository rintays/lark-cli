package main

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"lark/internal/testutil"
)

func TestMsgSendIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	chatID := os.Getenv("LARK_TEST_CHAT_ID")
	if chatID == "" {
		t.Skip("missing LARK_TEST_CHAT_ID")
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "messages", "send", "--receive-id-type", "chat_id", "--receive-id", chatID, "--text", "ping"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg send failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	id, ok := payload["message_id"].(string)
	if !ok || id == "" {
		t.Fatalf("expected non-empty message_id in output, got: %v", payload)
	}
}
