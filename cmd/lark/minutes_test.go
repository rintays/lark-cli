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

func TestMinutesInfoCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/open-apis/minutes/v1/minutes/m1" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"minute": map[string]any{
						"token": "m1",
						"title": "Weekly Sync",
						"url":   "http://example.com/m1",
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
				TenantAccessToken:          "token",
				TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			},
			Printer: output.Printer{Writer: &buf},
		}
		sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("sdk client error: %v", err)
		}
		state.SDK = sdkClient

		cmd := newMinutesCmd(state)
		cmd.SetArgs([]string{"info", "--minute-token", "m1", "--user-id-type", "open_id"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("minutes info error: %v", err)
		}

		if !strings.Contains(buf.String(), "m1\tWeekly Sync\thttp://example.com/m1") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires sdk client", func(t *testing.T) {
		state := &appState{
			Config: &config.Config{
				AppID:                      "app",
				AppSecret:                  "secret",
				BaseURL:                    "http://example.com",
				TenantAccessToken:          "token",
				TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			},
			Printer: output.Printer{Writer: &bytes.Buffer{}},
		}

		cmd := newMinutesCmd(state)
		cmd.SetArgs([]string{"info", "--minute-token", "m1"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "sdk client is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("requires minute token", func(t *testing.T) {
		cmd := newMinutesCmd(&appState{})
		cmd.SetArgs([]string{"info"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "required flag(s) \"minute-token\" not set" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMinutesListCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/open-apis/minutes/v1/minutes" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("user_id_type") != "open_id" {
				t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"token": "m1",
							"title": "Weekly Sync",
							"url":   "http://example.com/m1",
						},
					},
					"has_more": false,
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
				TenantAccessToken:          "token",
				TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			},
			Printer: output.Printer{Writer: &buf},
		}
		sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("sdk client error: %v", err)
		}
		state.SDK = sdkClient

		cmd := newMinutesCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2", "--user-id-type", "open_id"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("minutes list error: %v", err)
		}

		if !strings.Contains(buf.String(), "m1\tWeekly Sync\thttp://example.com/m1") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires sdk client", func(t *testing.T) {
		state := &appState{
			Config: &config.Config{
				AppID:                      "app",
				AppSecret:                  "secret",
				BaseURL:                    "http://example.com",
				TenantAccessToken:          "token",
				TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
			},
			Printer: output.Printer{Writer: &bytes.Buffer{}},
		}

		cmd := newMinutesCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "sdk client is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
