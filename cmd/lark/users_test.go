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

func TestUsersSearch(t *testing.T) {
	var calls int
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/search/v1/user" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Query().Get("query") != "Ada" {
			t.Fatalf("unexpected query: %s", r.URL.Query().Get("query"))
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		switch calls {
		case 0:
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.URL.Query().Get("page_token") != "" {
				t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
			}
		case 1:
			if r.URL.Query().Get("page_size") != "1" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
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
					"users": []map[string]any{
						{"user_id": "u1", "open_id": "ou1", "name": "Ada Lovelace", "department_ids": []string{"d1"}},
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
				"users": []map[string]any{
					{"user_id": "u2", "open_id": "ou2", "name": "Grace Hopper", "department_ids": []string{"d2", "d3"}},
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
			AppID:     "app",
			AppSecret: "secret",
			BaseURL:   baseURL,
		},
		Printer: output.Printer{Writer: &buf},
	}
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "", time.Now().Add(2*time.Hour).Unix(), "")
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newUsersCmd(state)
	cmd.SetArgs([]string{"search", "--limit", "2", "--pages", "2", "Ada"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("users search error: %v", err)
	}

	if !strings.Contains(buf.String(), "Ada Lovelace") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "Grace Hopper") {
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
	cmd.SetArgs([]string{"search", "Ada"})
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
			AppID:     "app",
			AppSecret: "secret",
			BaseURL:   baseURL,
		},
	}
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "", time.Now().Add(2*time.Hour).Unix(), "")
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
	if !strings.Contains(err.Error(), "search_query") {
		t.Fatalf("unexpected error: %v", err)
	}
}
