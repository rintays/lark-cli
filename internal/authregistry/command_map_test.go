package authregistry

import (
	"reflect"
	"testing"
)

func TestServicesForCommandPathMapping(t *testing.T) {
	tests := []struct {
		path []string
		want []string
	}{
		{path: []string{"drive"}, want: []string{"drive"}},
		{path: []string{"docs"}, want: []string{"docs"}},
		{path: []string{"sheets"}, want: []string{"sheets"}},
		{path: []string{"mail"}, want: []string{"mail"}},
		{path: []string{"wiki"}, want: []string{"wiki"}},
		{path: []string{"base"}, want: []string{"base"}},
		{path: []string{"calendar"}, want: []string{"calendar"}},
		{path: []string{"chats"}, want: []string{"im"}},
		{path: []string{"msg"}, want: []string{"im"}},
		{path: []string{"im"}, want: []string{"im"}},
	}

	for _, tt := range tests {
		got, ok := ServicesForCommandPath(tt.path)
		if !ok {
			t.Fatalf("ServicesForCommandPath(%v)=ok false, want true", tt.path)
		}
		if !reflect.DeepEqual(got, tt.want) {
			t.Fatalf("ServicesForCommandPath(%v)=%v, want %v", tt.path, got, tt.want)
		}
	}
}

func TestServicesForCommandPathPrefixMatch(t *testing.T) {
	got, ok := ServicesForCommandPath([]string{"drive", "list"})
	if !ok {
		t.Fatalf("ServicesForCommandPath(drive list)=ok false, want true")
	}
	want := []string{"drive"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ServicesForCommandPath(drive list)=%v, want %v", got, want)
	}
}

func TestServicesForCommandNormalization(t *testing.T) {
	got, ok := ServicesForCommand("  DRIVE\t  LiSt  ")
	if !ok {
		t.Fatalf("ServicesForCommand(DRIVE LiSt)=ok false, want true")
	}
	want := []string{"drive"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ServicesForCommand(DRIVE LiSt)=%v, want %v", got, want)
	}
}

func TestServicesForCommandPathLongestPrefixWins(t *testing.T) {
	orig, had := commandServiceMap["drive list"]
	commandServiceMap["drive list"] = []string{"docs"}
	t.Cleanup(func() {
		if had {
			commandServiceMap["drive list"] = orig
			return
		}
		delete(commandServiceMap, "drive list")
	})

	got, ok := ServicesForCommandPath([]string{"drive", "list"})
	if !ok {
		t.Fatalf("ServicesForCommandPath(drive list)=ok false, want true")
	}
	want := []string{"docs"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ServicesForCommandPath(drive list)=%v, want %v", got, want)
	}
}

func TestServicesForCommandPathUnknown(t *testing.T) {
	if _, ok := ServicesForCommandPath([]string{"unknown", "cmd"}); ok {
		t.Fatalf("ServicesForCommandPath(unknown cmd)=ok true, want false")
	}
}

func TestServicesForCommandDeterministicSortedUnique(t *testing.T) {
	orig := commandServiceMap["drive"]
	commandServiceMap["drive"] = []string{"docs", "drive", "docs"}
	t.Cleanup(func() {
		commandServiceMap["drive"] = orig
	})

	got, ok := ServicesForCommand("drive")
	if !ok {
		t.Fatalf("ServicesForCommand(drive)=ok false, want true")
	}
	want := []string{"docs", "drive"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("ServicesForCommand(drive)=%v, want %v", got, want)
	}
}
