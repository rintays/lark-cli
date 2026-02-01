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

func TestDrivePermissionsAddCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/members" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("need_notification") != "true" {
			t.Fatalf("unexpected need_notification: %s", r.URL.Query().Get("need_notification"))
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["member_type"] != "openid" {
			t.Fatalf("unexpected member_type: %+v", payload)
		}
		if payload["member_id"] != "ou_1" {
			t.Fatalf("unexpected member_id: %+v", payload)
		}
		if payload["perm"] != "view" {
			t.Fatalf("unexpected perm: %+v", payload)
		}
		if payload["perm_type"] != "container" {
			t.Fatalf("unexpected perm_type: %+v", payload)
		}
		if payload["type"] != "user" {
			t.Fatalf("unexpected type: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"member": map[string]any{
					"member_type": "openid",
					"member_id":   "ou_1",
					"perm":        "view",
					"perm_type":   "container",
					"type":        "user",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		TokenType: "user",
		Config: &config.Config{
			AppID:             "app",
			AppSecret:         "secret",
			BaseURL:           baseURL,
			TenantAccessToken: "tenant-token",
			UserAccounts: map[string]*config.UserAccount{
				defaultUserAccountName: {
					UserAccessToken:          "user-token",
					UserAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
				},
			},
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{
		"permissions", "add", "f1", "openid", "ou_1",
		"--type", "docx",
		"--perm", "view",
		"--perm-type", "container",
		"--member-kind", "user",
		"--need-notification",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive permissions add error: %v", err)
	}

	if !strings.Contains(buf.String(), "openid\tou_1\tview\tcontainer\tuser") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDrivePermissionsAddMissingFileTokenDoesNotCallHTTP(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{
		"permissions", "add",
		"--type", "docx",
		"--member-type", "openid",
		"--member-id", "ou_1",
		"--perm", "view",
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "file-token is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDrivePermissionsListCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/members" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("perm_type") != "container" {
			t.Fatalf("unexpected perm_type: %s", r.URL.Query().Get("perm_type"))
		}
		if r.URL.Query().Get("fields") != "name,avatar" {
			t.Fatalf("unexpected fields: %s", r.URL.Query().Get("fields"))
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
						"member_type": "openid",
						"member_id":   "ou_1",
						"perm":        "view",
						"perm_type":   "container",
						"type":        "user",
						"name":        "User One",
					},
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{
		"permissions", "list", "f1",
		"--type", "docx",
		"--perm-type", "container",
		"--fields", "name,avatar",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive permissions list error: %v", err)
	}

	if !strings.Contains(buf.String(), "openid\tou_1\tview\tcontainer\tuser\tUser One") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDrivePermissionsUpdateCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/members/ou_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("need_notification") != "true" {
			t.Fatalf("unexpected need_notification: %s", r.URL.Query().Get("need_notification"))
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["member_type"] != "openid" {
			t.Fatalf("unexpected member_type: %+v", payload)
		}
		if payload["member_id"] != "ou_1" {
			t.Fatalf("unexpected member_id: %+v", payload)
		}
		if payload["perm"] != "edit" {
			t.Fatalf("unexpected perm: %+v", payload)
		}
		if payload["type"] != "user" {
			t.Fatalf("unexpected type: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"member": map[string]any{
					"member_type": "openid",
					"member_id":   "ou_1",
					"perm":        "edit",
					"perm_type":   "container",
					"type":        "user",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		TokenType: "user",
		Config: &config.Config{
			AppID:             "app",
			AppSecret:         "secret",
			BaseURL:           baseURL,
			TenantAccessToken: "tenant-token",
			UserAccounts: map[string]*config.UserAccount{
				defaultUserAccountName: {
					UserAccessToken:          "user-token",
					UserAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
				},
			},
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{
		"permissions", "update", "f1", "openid", "ou_1",
		"--type", "docx",
		"--perm", "edit",
		"--member-kind", "user",
		"--need-notification",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive permissions update error: %v", err)
	}

	if !strings.Contains(buf.String(), "openid\tou_1\tedit\tcontainer\tuser") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDrivePermissionsDeleteCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/members/ou_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		if r.URL.Query().Get("member_type") != "openid" {
			t.Fatalf("unexpected member_type: %s", r.URL.Query().Get("member_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["perm_type"] != "container" {
			t.Fatalf("unexpected perm_type: %+v", payload)
		}
		if payload["type"] != "user" {
			t.Fatalf("unexpected type: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{
		"permissions", "delete", "f1", "openid", "ou_1",
		"--type", "docx",
		"--perm-type", "container",
		"--member-kind", "user",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive permissions delete error: %v", err)
	}

	if !strings.Contains(buf.String(), "ou_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
