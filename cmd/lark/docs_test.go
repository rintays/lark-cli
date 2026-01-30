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
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestDocsCreateCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/docx/v1/documents" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["title"] != "Specs" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if payload["folder_token"] != "fld" {
			t.Fatalf("unexpected folder_token: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"document": map[string]any{
					"document_id": "doc1",
					"title":       "Specs",
					"url":         "https://example.com/doc",
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"create", "--title", "Specs", "--folder-id", "fld"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs create error: %v", err)
	}

	if !strings.Contains(buf.String(), "doc1\tSpecs\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsGetCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/docx/v1/documents/doc1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"document": map[string]any{
					"document_id": "doc1",
					"title":       "Specs",
					"url":         "https://example.com/doc",
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"get", "--doc-id", "doc1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs get error: %v", err)
	}

	if !strings.Contains(buf.String(), "doc1\tSpecs\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
