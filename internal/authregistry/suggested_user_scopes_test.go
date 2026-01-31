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
	want := []string{"drive:drive", "mail:readonly"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("SuggestedUserOAuthScopesFromServices()=%v, want %v", got, want)
	}
}

func TestSuggestedUserOAuthScopesFromServicesReadonlyUsesVariantsAndFallback(t *testing.T) {
	got, err := SuggestedUserOAuthScopesFromServices([]string{"drive", "mail"}, true)
	if err != nil {
		t.Fatalf("SuggestedUserOAuthScopesFromServices(readonly) err=%v", err)
	}
	want := []string{"drive:drive:readonly", "mail:readonly"}
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
