package larksdk

import (
	"errors"
	"strings"
	"testing"
)

func TestWithInsufficientScopeRemediation_AppendsHint(t *testing.T) {
	base := errors.New("wiki node search failed: insufficient_scope")
	err := withInsufficientScopeRemediation(base, "insufficient_scope")
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "missing permission/scope") {
		t.Fatalf("expected remediation hint, got %q", err.Error())
	}
	if !strings.Contains(err.Error(), userOAuthReloginCommand) {
		t.Fatalf("expected relogin command, got %q", err.Error())
	}
}

func TestWithInsufficientScopeRemediation_NoChange(t *testing.T) {
	base := errors.New("wiki node search failed: something else")
	err := withInsufficientScopeRemediation(base, "rate limit")
	if err == nil {
		t.Fatalf("expected error")
	}
	if err.Error() != base.Error() {
		t.Fatalf("expected unchanged error, got %q", err.Error())
	}
}
