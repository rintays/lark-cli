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

func TestCalendarListRejectsPositionalArgsDoesNotCallHTTP(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("unexpected HTTP call: %s %s", r.Method, r.URL.Path)
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"list", "--start", "2020-01-01T00:00:00Z", "--end", "2020-01-02T00:00:00Z", "extra"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Fatal("unexpected http call")
	}
}
