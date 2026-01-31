package authregistry

import (
	"reflect"
	"testing"
)

func TestRequiredUserScopesFromServicesStableSortedUnique(t *testing.T) {
	got, err := RequiredUserScopesFromServices([]string{"sheets", "drive", "drive", "docs"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices() err=%v", err)
	}
	want := []string{"drive:drive"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("RequiredUserScopesFromServices()=%v, want %v", got, want)
	}
}

func TestRequiredUserScopesFromServicesOrderIndependence(t *testing.T) {
	a, err := RequiredUserScopesFromServices([]string{"drive", "mail", "wiki"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices(a) err=%v", err)
	}
	b, err := RequiredUserScopesFromServices([]string{"wiki", "mail", "drive"})
	if err != nil {
		t.Fatalf("RequiredUserScopesFromServices(b) err=%v", err)
	}
	if !reflect.DeepEqual(a, b) {
		t.Fatalf("union not deterministic: a=%v b=%v", a, b)
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
	offline, err = RequiresOfflineFromServices([]string{"drive", "base"})
	if err != nil {
		t.Fatalf("RequiresOfflineFromServices(drive,base) err=%v", err)
	}
	if !offline {
		t.Fatalf("RequiresOfflineFromServices(drive,base)=false, want true")
	}
}
