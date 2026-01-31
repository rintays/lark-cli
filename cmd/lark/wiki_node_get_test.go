package main

import (
	"bytes"
	"testing"
)

func TestWikiNodeGetCommandRegistered(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"wiki", "node", "get", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
}
