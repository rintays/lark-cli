package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestBaseAppGetCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer tenant-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"app": map[string]any{
					"app_token": "app_1",
					"name":      "MyApp",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"app", "info", "--app-token", "app_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base app get error: %v", err)
	}
	if !strings.Contains(buf.String(), "app_1\tMyApp") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestBaseAppGetCommandAcceptsPositionalAppToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/bitable/v1/apps/app_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"app": map[string]any{
					"app_token": "app_1",
					"name":      "MyApp",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			BaseURL:                    baseURL,
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"app", "info", "app_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("base app info error: %v", err)
	}
	if !strings.Contains(buf.String(), "app_1\tMyApp") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestBaseAppGetCommandRejectsAppTokenProvidedTwice(t *testing.T) {
	cmd := newBaseCmd(&appState{})
	cmd.SetArgs([]string{"app", "info", "--app-token", "app_1", "app_2"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "app-token provided twice" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestBaseAppGetCommandRequiresAppToken(t *testing.T) {
	cmd := newBaseCmd(&appState{})
	cmd.SetArgs([]string{"app", "info"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"app-token\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}
