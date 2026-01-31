package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"lark/internal/testutil"
)

func TestDriveUploadIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	folderToken := fx.DriveFolderToken

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "drive", "upload", "--folder-token", folderToken, "--file", filePath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive upload failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	tok := fileTokenFromPayload(payload)
	if tok == "" {
		t.Fatalf("expected non-empty file token, got: %v", payload)
	}
}

func fileTokenFromPayload(payload map[string]any) string {
	if tok, ok := payload["file_token"].(string); ok && tok != "" {
		return tok
	}
	if tok, ok := payload["token"].(string); ok && tok != "" {
		return tok
	}
	file, ok := payload["file"]
	if !ok {
		return ""
	}
	if fileMap, ok := file.(map[string]any); ok {
		if tok, ok := fileMap["file_token"].(string); ok && tok != "" {
			return tok
		}
		if tok, ok := fileMap["token"].(string); ok && tok != "" {
			return tok
		}
	}
	return ""
}
