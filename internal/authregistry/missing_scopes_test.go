package authregistry

import (
	"reflect"
	"testing"
)

func TestServicesMissingRequiredUserScopes(t *testing.T) {
	missing, err := ServicesMissingRequiredUserScopes([]string{"drive", "mail", "wiki", "base"})
	if err != nil {
		t.Fatalf("ServicesMissingRequiredUserScopes() err=%v", err)
	}
	// drive declares RequiredUserScopes; mail/wiki intentionally do not yet.
	want := []string{"mail", "wiki"}
	if !reflect.DeepEqual(missing, want) {
		t.Fatalf("ServicesMissingRequiredUserScopes()=%v, want %v", missing, want)
	}
}

func TestServicesMissingRequiredUserScopesUnknownService(t *testing.T) {
	_, err := ServicesMissingRequiredUserScopes([]string{"nope"})
	if err == nil {
		t.Fatalf("expected error")
	}
}
