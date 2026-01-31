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

func TestSheetsSearchCommandUsesDocsSearchEndpoint(t *testing.T) {
	t.Setenv("LARK_USER_ACCESS_TOKEN", "")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/suite/docs-api/search/object" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["search_key"] != "budget" {
			t.Fatalf("unexpected search_key: %+v", payload)
		}
		if payload["count"].(float64) != 10 {
			t.Fatalf("unexpected count: %+v", payload["count"])
		}
		if payload["offset"].(float64) != 0 {
			t.Fatalf("unexpected offset: %+v", payload["offset"])
		}
		docTypes, ok := payload["docs_types"].([]any)
		if !ok || len(docTypes) != 1 {
			t.Fatalf("unexpected docs_types: %+v", payload["docs_types"])
		}
		if docTypes[0].(string) != "sheet" {
			t.Fatalf("unexpected docs_types: %+v", docTypes)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"docs_entities": []map[string]any{
					{
						"docs_token": "s1",
						"docs_type":  "sheet",
						"title":      "Budget",
						"owner_id":   "ou_sheet",
						"open_url":   "https://example.com/sheet",
					},
				},
				"has_more": false,
				"total":    1,
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
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "", time.Now().Add(2*time.Hour).Unix(), "")
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newSheetsCmd(state)
	cmd.SetArgs([]string{"search", "budget", "--limit", "10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("sheets search error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "s1\tBudget\tsheet\thttps://example.com/sheet") {
		t.Fatalf("unexpected output: %q", out)
	}
}
