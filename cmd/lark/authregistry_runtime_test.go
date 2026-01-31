package main

import (
	"reflect"
	"strings"
	"testing"

	"lark/internal/authregistry"
)

func TestUserOAuthScopesForCommand(t *testing.T) {
	services, scopes, undeclared, ok, err := userOAuthScopesForCommand("mail send")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok")
	}
	if len(undeclared) != 0 {
		t.Fatalf("expected no undeclared services, got %v", undeclared)
	}
	if want := []string{"mail"}; !reflect.DeepEqual(services, want) {
		t.Fatalf("services=%v, want %v", services, want)
	}
	if want := []string{"offline_access", "mail:readonly"}; !reflect.DeepEqual(scopes, want) {
		t.Fatalf("scopes=%v, want %v", scopes, want)
	}

	services, scopes, undeclared, ok, err = userOAuthScopesForCommand("drive search")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok")
	}
	if len(undeclared) != 0 {
		t.Fatalf("expected no undeclared services, got %v", undeclared)
	}
	if want := []string{"drive"}; !reflect.DeepEqual(services, want) {
		t.Fatalf("services=%v, want %v", services, want)
	}
	if want := []string{"offline_access", "drive:drive"}; !reflect.DeepEqual(scopes, want) {
		t.Fatalf("scopes=%v, want %v", scopes, want)
	}

	_, _, _, ok, err = userOAuthScopesForCommand("chats list")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false for tenant-only service")
	}

	_, _, _, ok, err = userOAuthScopesForCommand("unknown cmd")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if ok {
		t.Fatalf("expected ok=false for unmapped command")
	}
}

func TestUserOAuthReloginCommandForCommand(t *testing.T) {
	{
		cmd, note, ok, err := userOAuthReloginCommandForCommand("mail send")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !ok {
			t.Fatalf("expected ok")
		}
		if cmd != "lark auth user login --scopes \"offline_access mail:readonly\" --force-consent" {
			t.Fatalf("cmd=%q", cmd)
		}
		if note == "" {
			t.Fatalf("expected note")
		}
	}

	{
		cmd, note, ok, err := userOAuthReloginCommandForCommand("chats list")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatalf("expected ok=false")
		}
		if cmd != "" || note != "" {
			t.Fatalf("expected empty cmd/note for tenant-only command; got cmd=%q note=%q", cmd, note)
		}
	}

	{
		cmd, note, ok, err := userOAuthReloginCommandForCommand("unknown cmd")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if ok {
			t.Fatalf("expected ok=false")
		}
		if cmd != "" || note != "" {
			t.Fatalf("expected empty cmd/note for unmapped command; got cmd=%q note=%q", cmd, note)
		}
	}
}

func TestUserOAuthReloginCommandForCommand_MissingScopeDeclarations(t *testing.T) {
	orig := authregistry.Registry["mail"]
	origScopes := orig.RequiredUserScopes
	orig.RequiredUserScopes = nil
	authregistry.Registry["mail"] = orig
	t.Cleanup(func() {
		orig.RequiredUserScopes = origScopes
		authregistry.Registry["mail"] = orig
	})

	cmd, note, ok, err := userOAuthReloginCommandForCommand("mail send")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok")
	}
	if cmd != "lark auth user login --scopes \"offline_access\" --force-consent" {
		t.Fatalf("cmd=%q", cmd)
	}
	if !strings.Contains(note, "missing scope declarations") {
		t.Fatalf("note=%q", note)
	}
}
