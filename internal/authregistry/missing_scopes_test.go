package authregistry

import (
	"reflect"
	"testing"
)

func TestServicesMissingRequiredUserScopes(t *testing.T) {
	missing, err := ServicesMissingRequiredUserScopes([]string{"drive-metadata", "mail", "wiki", "base"})
	if err != nil {
		t.Fatalf("ServicesMissingRequiredUserScopes() err=%v", err)
	}
	if len(missing) != 0 {
		t.Fatalf("ServicesMissingRequiredUserScopes()=%v, want empty", missing)
	}

	// Verify the detection behavior when a TokenUser service has undeclared (nil)
	// RequiredUserScopes.
	origMail := Registry["mail"]
	origWiki := Registry["wiki"]
	t.Cleanup(func() {
		Registry["mail"] = origMail
		Registry["wiki"] = origWiki
	})

	mail := origMail
	mail.RequiredUserScopes = nil
	Registry["mail"] = mail

	wiki := origWiki
	wiki.RequiredUserScopes = nil
	Registry["wiki"] = wiki

	missing, err = ServicesMissingRequiredUserScopes([]string{"drive-metadata", "mail", "wiki", "base"})
	if err != nil {
		t.Fatalf("ServicesMissingRequiredUserScopes() err=%v", err)
	}
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
