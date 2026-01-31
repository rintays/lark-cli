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
		cmd.SetArgs([]string{"info", "m1", "--user-id-type", "open_id"})
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
		cmd.SetArgs([]string{"info", "m1"})
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
		if err.Error() != "minute-token is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMinutesListCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("unexpected method: %s", r.Method)
			}
			if r.URL.Path != "/open-apis/drive/v1/files" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("folder_token") != "root" {
				t.Fatalf("unexpected folder_token: %s", r.URL.Query().Get("folder_token"))
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"files": []map[string]any{
						{
							"token": "m1",
							"name":  "Weekly Sync",
							"type":  "minutes",
							"url":   "http://example.com/m1",
						},
						{
							"token": "d1",
							"name":  "Docs",
							"type":  "docx",
							"url":   "http://example.com/d1",
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
		cmd.SetArgs([]string{"list", "--limit", "2", "--folder-id", "root"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("minutes list error: %v", err)
		}

		out := buf.String()
		if !strings.Contains(out, "m1\tWeekly Sync\thttp://example.com/m1") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
		if strings.Contains(out, "d1\tDocs") {
			t.Fatalf("expected non-minutes files to be filtered out, got: %q", out)
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

func TestMinutesDeleteCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/files/m1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "minutes" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"task_id": "task1",
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
	cmd.SetArgs([]string{"delete", "m1", "--type", "minutes"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("minutes delete error: %v", err)
	}

	if !strings.Contains(buf.String(), "m1\tminutes\ttask1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMinutesUpdateCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/m1/public" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "minutes" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["link_share_entity"] != "tenant_readable" {
			t.Fatalf("unexpected link_share_entity: %+v", payload)
		}
		if payload["external_access"] != true {
			t.Fatalf("unexpected external_access: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"permission_public": map[string]any{
					"link_share_entity": "tenant_readable",
					"external_access":   true,
					"invite_external":   false,
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
	cmd.SetArgs([]string{"update", "m1", "--link-share", "tenant_readable", "--external-access"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("minutes update error: %v", err)
	}

	if !strings.Contains(buf.String(), "m1\tminutes\ttenant_readable\ttrue\tfalse") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
