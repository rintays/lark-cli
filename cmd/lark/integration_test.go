package main

import (
	"testing"

	"lark/internal/testutil"
)

func TestIntegrationRootCommandConstructs(t *testing.T) {
	testutil.RequireIntegration(t)

	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("newRootCmd panicked: %v", r)
		}
	}()

	cmd := newRootCmd()
	if cmd == nil {
		t.Fatal("expected root command")
	}
}
