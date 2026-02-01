package main

import (
	"bytes"
	"testing"
)

func TestWikiNodeExtraCommandsRegistered(t *testing.T) {
	commands := [][]string{
		{"wiki", "node", "create", "--help"},
		{"wiki", "node", "move", "--help"},
		{"wiki", "node", "update-title", "--help"},
		{"wiki", "node", "attach", "--help"},
	}
	for _, args := range commands {
		cmd := newRootCmd()
		var buf bytes.Buffer
		cmd.SetOut(&buf)
		cmd.SetErr(&buf)
		cmd.SetArgs(args)
		if err := cmd.Execute(); err != nil {
			t.Fatalf("help failed for %v: %v", args, err)
		}
	}
}

func TestWikiSpaceUpdateSettingCommandRegistered(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"wiki", "space", "update-setting", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
}
