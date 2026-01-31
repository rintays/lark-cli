package authregistry

import (
	"reflect"
	"testing"
)

func TestSuggestedUserOAuthScopesFromServicesFullUsesVariantsAndFallback(t *testing.T) {
	got, err := SuggestedUserOAuthScopesFromServices([]string{"drive", "mail"}, false)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices() err=%v", err)
	}
	want := []string{"drive:drive", "mail:user_mailbox.message:readonly", "mail:user_mailbox.message:send"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedUserOAuthScopesFromServices()=%v, want %v", got, want)
	}
}

func TestSuggestedUserOAuthScopesFromServicesReadonlyUsesVariantsAndFallback(t *testing.T) {
	got, err := SuggestedUserOAuthScopesFromServices([]string{"drive", "mail"}, true)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(readonly) err=%v", err)
	}
	want := []string{"drive:drive:readonly", "mail:user_mailbox.message:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(readonly)=%v, want %v", got, want)
	}
}

func TestSuggestedUserOAuthScopesFromServicesUnknownService(t *testing.T) {
	_, err := SuggestedUserOAuthScopesFromServices([]string{"nope"}, false)
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestSuggestedUserOAuthScopesFromServicesReadonlyFallsBackToFullWhenReadonlyVariantMissing(t *testing.T) {
	Registry["test-full-only"] = ServiceDef{
		Name:       "test-full-only",
		TokenTypes: []TokenType{TokenUser},
		UserScopes: ServiceScopeSet{Full: []string{"test:full"}},
	}
	t.Cleanup(func() {
		delete(Registry, "test-full-only")
	})

	got, err := SuggestedUserOAuthScopesFromServices([]string{"test-full-only"}, true)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(test-full-only readonly) err=%v", err)
	}
	want := []string{"test:full"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedUserOAuthScopesFromServices()=%v, want %v", got, want)
	}
}

func TestSuggestedUserOAuthScopesFromServicesFullFallsBackToReadonlyWhenFullVariantMissing(t *testing.T) {
	Registry["test-readonly-only"] = ServiceDef{
		Name:       "test-readonly-only",
		TokenTypes: []TokenType{TokenUser},
		UserScopes: ServiceScopeSet{Readonly: []string{"test:readonly"}},
	}
	t.Cleanup(func() {
		delete(Registry, "test-readonly-only")
	})

	got, err := SuggestedUserOAuthScopesFromServices([]string{"test-readonly-only"}, false)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(test-readonly-only full) err=%v", err)
	}
	want := []string{"test:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedUserOAuthScopesFromServices()=%v, want %v", got, want)
	}
}

func TestSuggestedUserOAuthScopesFromServicesDeterministicUnion(t *testing.T) {
	a, err := SuggestedUserOAuthScopesFromServices([]string{"drive", "mail"}, true)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(a) err=%v", err)
	}
	b, err := SuggestedUserOAuthScopesFromServices([]string{"mail", "drive"}, true)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(b) err=%v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("union not deterministic: a=%v b=%v", a, b)
	}
}
