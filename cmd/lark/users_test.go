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

func TestUsersInfoCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/ou_1" {
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
				"user": map[string]any{
					"user_id": "u_1",
					"open_id": "ou_1",
					"name":    "Ada",
					"email":   "ada@example.com",
					"mobile":  "+1-555-0100",
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

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"info", "--open-id", "ou_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("users info error: %v", err)
	}

	if !strings.Contains(buf.String(), "u_1\tAda\tada@example.com\t+1-555-0100") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestUsersSearchByEmail(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/batch_get_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["emails"][0] != "dev@example.com" {
			t.Fatalf("unexpected emails: %+v", payload["emails"])
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"user_list": []map[string]any{{"user_id": "u1", "email": "dev@example.com"}},
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

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"search", "--email", "dev@example.com"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("users search error: %v", err)
	}

	if !strings.Contains(buf.String(), "u1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestUsersSearchByName(t *testing.T) {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/find_by_department" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("department_id") != "0" {
			t.Fatalf("unexpected department_id: %s", r.URL.Query().Get("department_id"))
		}
		if r.URL.Query().Get("page_size") != "50" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch calls {
		case 0:
			if r.URL.Query().Get("page_token") != "" {
				t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
			}
		case 1:
			if r.URL.Query().Get("page_token") != "next" {
				t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
			}
		default:
			t.Fatalf("unexpected request count: %d", calls)
		}
		w.Header().Set("Content-Type", "application/json")
		if calls == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{"user_id": "u1", "name": "Ada Lovelace"},
					},
					"has_more":   true,
					"page_token": "next",
				},
			})
			calls++
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{"user_id": "u2", "name": "Grace Hopper"},
				},
				"has_more": false,
			},
		})
		calls++
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

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"search", "--name", "Ada"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("users search error: %v", err)
	}

	if !strings.Contains(buf.String(), "Ada Lovelace") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if strings.Contains(buf.String(), "Grace Hopper") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestUsersSearchRequiresSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s", r.URL.Path)
	})
	_, baseURL := testutil.NewTestClient(handler)

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

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"list", "--email", "dev@example.com"})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected error")
	}
	if !strings.Contains(err.Error(), "sdk client is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsersSearchRequiresCriteria(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s", r.URL.Path)
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"search"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "email") && !strings.Contains(err.Error(), "mobile") && !strings.Contains(err.Error(), "name") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestUsersSearchRejectsMultipleCriteria(t *testing.T) {
	cmd := newUsersCmd(&appState{})
	cmd.SetArgs([]string{"search", "--email", "dev@example.com", "--name", "Ada"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "group") && !strings.Contains(err.Error(), "were all set") {
		t.Fatalf("unexpected error: %v", err)
	}
}
