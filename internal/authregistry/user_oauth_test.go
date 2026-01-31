package authregistry

import (
	"reflect"
	"testing"
)

func TestListUserOAuthServicesStableSorted(t *testing.T) {
	got := ListUserOAuthServices()
	want := []string{"docs", "docx", "drive", "sheets"}
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
