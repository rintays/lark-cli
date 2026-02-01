package main

import (
	"bytes"
	"testing"
)

func TestWikiNodeInfoCommandRegistered(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"wiki", "node", "info", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
}

func TestWikiNodeInfoCommandRequiresNodeToken(t *testing.T) {
	cmd := newWikiCmd(&appState{})
	cmd.SetArgs([]string{"node", "info", "", "docx"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "node-token is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWikiNodeInfoCommandRequiresObjType(t *testing.T) {
	cmd := newWikiCmd(&appState{})
	cmd.SetArgs([]string{"node", "info", "node_1", ""})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "obj-type is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
