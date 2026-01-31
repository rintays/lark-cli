package main

import (
	"reflect"
	"testing"
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
