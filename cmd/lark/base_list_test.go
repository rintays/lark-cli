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

func TestBasesListCommandUsesDriveSearchAndExtractsAppToken(t *testing.T) {
	t.Setenv("LARK_USER_ACCESS_TOKEN", "")

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/files/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}

		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["search_key"] != "budget" {
			t.Fatalf("unexpected search_key: %+v", payload["search_key"])
		}
		fileTypes, ok := payload["file_types"].([]any)
		if !ok || len(fileTypes) != 1 || fileTypes[0].(string) != "bitable" {
			t.Fatalf("unexpected file_types: %+v", payload["file_types"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files": []map[string]any{{
					"token": "f_bitable",
					"name":  "My Base",
					"type":  "bitable",
					"url":   "https://example.com/base/bascnTOKEN123?tab=table",
				}},
				"has_more":   false,
				"page_token": "",
			},
		})
	})

	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:     "app",
			AppSecret: "secret",
			BaseURL:   baseURL,
			UserAccounts: map[string]*config.UserAccount{
				defaultUserAccountName: {
					UserAccessToken:          "user-token",
					UserAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
				},
			},
		},
		Printer: output.Printer{Writer: &buf},
	}
	client, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = client

	cmd := newBaseCmd(state)
	cmd.SetArgs([]string{"list", "--query", "budget", "--limit", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("bases list error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "bascnTOKEN123") {
		t.Fatalf("expected app_token in output, got: %q", out)
	}
	if !strings.Contains(out, "f_bitable") || !strings.Contains(out, "My Base") {
		t.Fatalf("unexpected output: %q", out)
	}
}
