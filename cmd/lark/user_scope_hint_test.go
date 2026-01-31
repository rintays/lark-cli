package main

import (
	"errors"
	"strings"
	"testing"
)

func TestWithUserScopeHintForCommand_InferredScopes(t *testing.T) {
	state := &appState{Command: "drive search"}
	err := errors.New("permission denied")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	msg := got.Error()
	if !strings.Contains(msg, "permission denied") {
		t.Fatalf("unexpected message: %s", msg)
	}
	if !strings.Contains(msg, "Missing user OAuth scopes") {
		t.Fatalf("expected hint, got: %s", msg)
	}
	if !strings.Contains(msg, "based on command \"drive search\"") {
		t.Fatalf("expected derivation, got: %s", msg)
	}
	if !strings.Contains(msg, "drive:drive") {
		t.Fatalf("expected drive scope, got: %s", msg)
	}
	if !strings.Contains(msg, "lark auth user login --scopes \"offline_access drive:drive\" --force-consent") {
		t.Fatalf("expected relogin command, got: %s", msg)
	}
}

func TestWithUserScopeHintForCommand_ExtractedScopesWin(t *testing.T) {
	state := &appState{Command: "drive search"}
	err := errors.New("permission denied [drive:drive:readonly]")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	msg := got.Error()
	if !strings.Contains(msg, "drive:drive:readonly") {
		t.Fatalf("expected extracted scope, got: %s", msg)
	}
	if strings.Contains(msg, "based on command") {
		t.Fatalf("did not expect inferred note when extracted scopes exist, got: %s", msg)
	}
}

func TestWithUserScopeHintForCommand_UnchangedWhenNoScopes(t *testing.T) {
	state := &appState{Command: "chats list"}
	err := errors.New("permission denied")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	if got.Error() != err.Error() {
		t.Fatalf("expected unchanged error, got: %s", got)
	}
}
