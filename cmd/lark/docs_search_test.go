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

func TestDocsSearchCommandUsesDocsSearchEndpoint(t *testing.T) {
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
		if payload["search_key"] != "spec" {
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
		if docTypes[0].(string) != "doc" {
			t.Fatalf("unexpected docs_types: %+v", docTypes)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"docs_entities": []map[string]any{
					{
						"docs_token": "d1",
						"docs_type":  "doc",
						"title":      "Specs",
						"owner_id":   "ou_docs",
						"open_url":   "https://example.com/docx",
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
			UserAccessToken:            "user-token",
			UserAccessTokenExpiresAt:   time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"search", "--query", "spec", "--limit", "10"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs search error: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "d1\tSpecs\tdocx\thttps://example.com/docx") {
		t.Fatalf("unexpected output: %q", out)
	}
}
