package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
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

func TestDocsExportCommand(t *testing.T) {
	exported := []byte("pdf bytes")
	pollCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/drive/v1/export_tasks":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["token"] != "doc1" {
				t.Fatalf("unexpected token: %+v", payload)
			}
			if payload["type"] != "docx" {
				t.Fatalf("unexpected type: %+v", payload)
			}
			if payload["file_extension"] != "pdf" {
				t.Fatalf("unexpected file_extension: %+v", payload)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"ticket": "ticket1",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/drive/v1/export_tasks/ticket1":
			pollCount++
			result := map[string]any{
				"file_extension": "pdf",
				"type":           "docx",
				"file_name":      "Doc.pdf",
				"file_token":     "",
				"file_size":      0,
				"job_error_msg":  "running",
				"job_status":     1,
			}
			if pollCount > 1 {
				result["file_token"] = "file1"
				result["file_size"] = int64(len(exported))
				result["job_error_msg"] = "success"
				result["job_status"] = 0
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"result": result,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/drive/v1/export_tasks/file/file1/download":
			_, _ = w.Write(exported)
		default:
			t.Fatalf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
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

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "export.pdf")
	prevInterval := exportTaskPollInterval
	exportTaskPollInterval = 0
	defer func() {
		exportTaskPollInterval = prevInterval
	}()
	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"export", "--doc-id", "doc1", "--format", "pdf", "--out", outPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs export error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read exported file: %v", err)
	}
	if !bytes.Equal(data, exported) {
		t.Fatalf("unexpected export content: %q", string(data))
	}
}

func TestDocsCatCommand(t *testing.T) {
	exported := []byte("Hello doc\nLine two\n")
	pollCount := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/drive/v1/export_tasks":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["token"] != "doc1" {
				t.Fatalf("unexpected token: %+v", payload)
			}
			if payload["type"] != "docx" {
				t.Fatalf("unexpected type: %+v", payload)
			}
			if payload["file_extension"] != "txt" {
				t.Fatalf("unexpected file_extension: %+v", payload)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"ticket": "ticket1",
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/drive/v1/export_tasks/ticket1":
			pollCount++
			result := map[string]any{
				"file_extension": "txt",
				"type":           "docx",
				"file_name":      "Doc.txt",
				"file_token":     "",
				"file_size":      0,
				"job_error_msg":  "running",
				"job_status":     1,
			}
			if pollCount > 1 {
				result["file_token"] = "file1"
				result["file_size"] = int64(len(exported))
				result["job_error_msg"] = "success"
				result["job_status"] = 0
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"result": result,
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/drive/v1/export_tasks/file/file1/download":
			_, _ = w.Write(exported)
		default:
			t.Fatalf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
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

	prevInterval := exportTaskPollInterval
	exportTaskPollInterval = 0
	defer func() {
		exportTaskPollInterval = prevInterval
	}()
	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"cat", "--doc-id", "doc1", "--format", "txt"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs cat error: %v", err)
	}

	if got := buf.String(); got != string(exported) {
		t.Fatalf("unexpected output: %q", got)
	}
}
