package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"lark/internal/testutil"
)

func TestMsgSendIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	chatID := fx.EnsureChatID(t)

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "messages", "send", "--receive-id-type", "chat_id", "--receive-id", chatID, "--text", "ping"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("msg send failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	id := messageIDFromPayload(payload)
	if id == "" {
		t.Fatalf("expected non-empty message id in output, got: %v", payload)
	}
}

func messageIDFromPayload(payload map[string]any) string {
	if id, ok := payload["message_id"].(string); ok && strings.TrimSpace(id) != "" {
		return id
	}
	message, ok := payload["message"]
	if !ok {
		return ""
	}
	switch value := message.(type) {
	case string:
		if strings.TrimSpace(value) != "" {
			return value
		}
	case map[string]any:
		if id, ok := value["message_id"].(string); ok && strings.TrimSpace(id) != "" {
			return id
		}
		if id, ok := value["id"].(string); ok && strings.TrimSpace(id) != "" {
			return id
		}
	}
	return ""
}
