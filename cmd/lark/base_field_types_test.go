package main

import (
	"bytes"
	"strings"
	"testing"

	"lark/internal/output"
)

func TestBaseFieldTypesCommand(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{Printer: output.Printer{Writer: &buf}}
	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"field", "types"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base field types error: %v", err)
	}
	if !strings.Contains(buf.String(), "1\ttext") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
