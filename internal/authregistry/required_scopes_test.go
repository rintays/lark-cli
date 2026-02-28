package authregistry

import (
	"reflect"
	"testing"
)

func TestRequiredUserScopesFromServicesStableSortedUnique(t *testing.T) {
	got, err := RequiredUserScopesFromServices([]string{"sheets", "drive-read", "drive-read", "docs"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices() err=%v", err)
	}
	want := []string{
		"docs:document.comment:read",
		"docx:document:readonly",
		"drive:drive.metadata:readonly",
		"drive:drive.search:readonly",
		"sheets:spreadsheet:read",
		"space:document:retrieve",
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RequiredUserScopesFromServices()=%v, want %v", got, want)
	}
}

func TestRequiredUserScopesFromServicesOrderIndependence(t *testing.T) {
	a, err := RequiredUserScopesFromServices([]string{"drive-read", "mail", "wiki"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices(a) err=%v", err)
	}
	b, err := RequiredUserScopesFromServices([]string{"wiki", "mail", "drive-read"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices(b) err=%v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("union not deterministic: a=%v b=%v", a, b)
	}
	want := []string{
		"docs:document.comment:read",
		"drive:drive.metadata:readonly",
		"drive:drive.search:readonly",
		"mail:user_mailbox.message.address:read",
		"mail:user_mailbox.message.body:read",
		"mail:user_mailbox.message.subject:read",
		"mail:user_mailbox.message:readonly",
		"space:document:retrieve",
		"wiki:wiki",
	}
	if !reflect.DeepEqual(a, want) {
		t.Fatalf("RequiredUserScopesFromServices()=%v, want %v", a, want)
	}
}

func TestRequiredUserScopesFromServicesReportUndeclaredVsEmpty(t *testing.T) {
	orig := Registry["mail"]
	t.Cleanup(func() {
		Registry["mail"] = orig
	})

	// nil RequiredUserScopes means "unknown/undeclared".
	tmp := orig
	tmp.RequiredUserScopes = nil
	Registry["mail"] = tmp

	scopes, missing, err := RequiredUserScopesFromServicesReport([]string{"mail"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServicesReport() err=%v", err)
	}
	if len(scopes) != 0 {
		t.Fatalf("scopes=%v, want empty when undeclared", scopes)
	}
	if want := []string{"mail"}; !reflect.DeepEqual(missing, want) {
		t.Fatalf("missing=%v, want %v", missing, want)
	}

	// An explicitly empty slice means "declared (but empty)".
	tmp.RequiredUserScopes = []string{}
	Registry["mail"] = tmp

	scopes, missing, err = RequiredUserScopesFromServicesReport([]string{"mail"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServicesReport() err=%v", err)
	}
	if len(scopes) != 0 {
		t.Fatalf("scopes=%v, want empty", scopes)
	}
	if len(missing) != 0 {
		t.Fatalf("missing=%v, want empty", missing)
	}
}

func TestRequiredUserScopesFromServicesUnknownService(t *testing.T) {
	_, err := RequiredUserScopesFromServices([]string{"nope"})
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestRequiresOfflineFromServices(t *testing.T) {
	offline, err := RequiresOfflineFromServices([]string{"base"})
	if err != nil {
		t.Fatalf("RequiresOfflineFromServices(base) err=%v", err)
	}
	if offline {
		t.Fatalf("RequiresOfflineFromServices(base)=true, want false")
	}
	offline, err = RequiresOfflineFromServices([]string{"drive-read", "base"})
	if err != nil {
		t.Fatalf("RequiresOfflineFromServices(drive,base) err=%v", err)
	}
	if !offline {
		t.Fatalf("RequiresOfflineFromServices(drive,base)=false, want true")
	}
}
