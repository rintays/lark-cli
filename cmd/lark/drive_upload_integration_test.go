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

	folderToken := os.Getenv("LARK_TEST_FOLDER_TOKEN")
	if folderToken == "" {
		t.Skip("missing LARK_TEST_FOLDER_TOKEN")
	}

	tmpDir := t.TempDir()
	filePath := filepath.Join(tmpDir, "hello.txt")
	if err := os.WriteFile(filePath, []byte("hello"), 0o644); err != nil {
		t.Fatalf("write temp file: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--json", "drive", "upload", "--folder-token", folderToken, "--file", filePath})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive upload failed: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	tok, ok := payload["file_token"].(string)
	if !ok || tok == "" {
		t.Fatalf("expected non-empty file_token, got: %v", payload)
	}
}
