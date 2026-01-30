package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larkapi"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestUsersSearchByEmail(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/batch_get_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
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

func TestUsersSearchByEmailFallbackToAPI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/batch_get_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
		SDK:     &larksdk.Client{},
	}

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
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{"user_id": "u1", "name": "Ada Lovelace"},
					{"user_id": "u2", "name": "Grace Hopper"},
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
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
