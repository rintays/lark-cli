package authregistry

import (
	"reflect"
	"testing"
)

func TestListUserOAuthServicesStableSorted(t *testing.T) {
	got := ListUserOAuthServices()
	want := []string{"calendar", "docs", "docx", "drive", "mail", "search-message", "sheets", "wiki"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ListUserOAuthServices()=%v, want %v", got, want)
	}
}

func TestExpandUserOAuthServiceAliases(t *testing.T) {
	got := ExpandUserOAuthServiceAliases([]string{"all"})
	want := []string{"drive", "docs", "docx", "sheets"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ExpandUserOAuthServiceAliases(all)=%v, want %v", got, want)
	}
}

func TestUserOAuthScopesFromServicesStableUnion(t *testing.T) {
	got1, err := UserOAuthScopesFromServices([]string{"drive", "docs"}, false, "")
	if err != nil {
		t.Fatalf("scopes(drive,docs): %v", err)
	}
	got2, err := UserOAuthScopesFromServices([]string{"docs", "drive"}, false, "")
	if err != nil {
		t.Fatalf("scopes(docs,drive): %v", err)
	}
	want := []string{"drive:drive"}
	if !reflect.DeepEqual(got1, want) {
		t.Fatalf("scopes(drive,docs)=%v, want %v", got1, want)
	}
	if !reflect.DeepEqual(got2, want) {
		t.Fatalf("scopes(docs,drive)=%v, want %v", got2, want)
	}
}

func TestUserOAuthScopesFromServicesReadonly(t *testing.T) {
	got, err := UserOAuthScopesFromServices([]string{"drive"}, true, "")
	if err != nil {
		t.Fatalf("scopes(drive readonly): %v", err)
	}
	want := []string{"drive:drive:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes=%v, want %v", got, want)
	}
}

func TestUserOAuthScopesFromServicesReadonlyUsesVariantsAndFallback(t *testing.T) {
	got, err := UserOAuthScopesFromServices([]string{"drive", "mail"}, true, "")
	if err != nil {
		t.Fatalf("scopes(drive,mail readonly): %v", err)
	}
	want := []string{"drive:drive:readonly", "mail:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes=%v, want %v", got, want)
	}

	// Ensure aggregation is deterministic w.r.t input ordering.
	got2, err := UserOAuthScopesFromServices([]string{"mail", "drive"}, true, "")
	if err != nil {
		t.Fatalf("scopes(mail,drive readonly): %v", err)
	}
	if !reflect.DeepEqual(got2, want) {
		t.Fatalf("scopes=%v, want %v", got2, want)
	}
}

func TestUserOAuthScopesFromServicesReadonlyFallsBackToFullWhenReadonlyVariantMissing(t *testing.T) {
	Registry["test-full-only"] = ServiceDef{
		Name:       "test-full-only",
		TokenTypes: []TokenType{TokenUser},
		UserScopes: ServiceScopeSet{Full: []string{"test:full"}},
	}
	t.Cleanup(func() {
		delete(Registry, "test-full-only")
	})

	got, err := UserOAuthScopesFromServices([]string{"test-full-only"}, true, "")
	if err != nil {
		t.Fatalf("scopes(test-full-only readonly): %v", err)
	}
	want := []string{"test:full"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes=%v, want %v", got, want)
	}
}

func TestUserOAuthScopesFromServicesFullFallsBackToReadonlyWhenFullVariantMissing(t *testing.T) {
	Registry["test-readonly-only"] = ServiceDef{
		Name:       "test-readonly-only",
		TokenTypes: []TokenType{TokenUser},
		UserScopes: ServiceScopeSet{Readonly: []string{"test:readonly"}},
	}
	t.Cleanup(func() {
		delete(Registry, "test-readonly-only")
	})

	got, err := UserOAuthScopesFromServices([]string{"test-readonly-only"}, false, "")
	if err != nil {
		t.Fatalf("scopes(test-readonly-only full): %v", err)
	}
	want := []string{"test:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("scopes=%v, want %v", got, want)
	}
}
