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
	if !strings.Contains(msg, "drive:drive") {
		t.Fatalf("expected drive scope, got: %s", msg)
	}
	if !strings.Contains(msg, "search:docs:read") {
		t.Fatalf("expected search docs scope, got: %s", msg)
	}
	if !strings.Contains(msg, "lark auth user login --scopes \"offline_access drive:drive search:docs:read\" --force-consent") {
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
}

func TestWithUserScopeHintForCommand_ChatsListScopes(t *testing.T) {
	state := &appState{Command: "chats list"}
	err := errors.New("permission denied")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	msg := got.Error()
	if !strings.Contains(msg, "im:chat.group_info:readonly") {
		t.Fatalf("expected chat scope, got: %s", msg)
	}
	if !strings.Contains(msg, "lark auth user login --scopes \"offline_access im:chat.group_info:readonly\" --force-consent") {
		t.Fatalf("expected relogin command, got: %s", msg)
	}
}

func TestWithUserScopeHintForCommand_MessageSearchScopes(t *testing.T) {
	state := &appState{Command: "messages search"}
	err := errors.New("permission denied")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	msg := got.Error()
	if !strings.Contains(msg, "im:message:readonly") {
		t.Fatalf("expected message get scope, got: %s", msg)
	}
	if !strings.Contains(msg, "search:message") {
		t.Fatalf("expected search scope, got: %s", msg)
	}
}

func TestWithUserScopeHintForCommand_SkipsNonPermissionErrors(t *testing.T) {
	state := &appState{Command: "drive search"}
	err := errors.New("search drive files failed (code=99991400): request trigger frequency limit")
	got := withUserScopeHintForCommand(state, err)
	if got == nil {
		t.Fatalf("expected error")
	}
	if got.Error() != err.Error() {
		t.Fatalf("expected unchanged error, got: %s", got)
	}
}
