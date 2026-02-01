package main

import (
	"bytes"
	"net/http"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestWikiMemberAddCommandRequiresMemberID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call")
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "add", "--space-id", "spc1", "userid"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"member-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWikiMemberAddCommandRequiresMemberType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call")
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &bytes.Buffer{}},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "add", "--space-id", "spc1", "--member-id", "u1"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"member-type\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}
