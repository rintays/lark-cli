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
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestDocsCreateCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/docx/v1/documents" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("unexpected query: %q", r.URL.RawQuery)
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
		w.Header().Set("Content-Type", "application/json")
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"create", "Specs", "--folder-id", "fld"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs create error: %v", err)
	}

	if !strings.Contains(buf.String(), "doc1\tSpecs\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsCreateCommandFetchesURL(t *testing.T) {
	createCalled := 0
	getCalled := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/docx/v1/documents":
			createCalled++
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"document": map[string]any{
						"document_id": "doc1",
						"title":       "Specs",
						"url":         "",
					},
				},
			})
			return
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/docx/v1/documents/doc1":
			getCalled++
			w.Header().Set("Content-Type", "application/json")
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
			return
		default:
			t.Fatalf("unexpected request: %s %s", r.Method, r.URL.Path)
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"create", "Specs"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs create error: %v", err)
	}

	if createCalled != 1 || getCalled != 1 {
		t.Fatalf("unexpected call counts: create=%d get=%d", createCalled, getCalled)
	}
	if !strings.Contains(buf.String(), "doc1\tSpecs\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsCreateCommandMissingTitleDoesNotCallHTTP(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("unexpected http call: %s %s", r.Method, r.URL.Path)
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"create"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Fatal("unexpected http call")
	}
}

func TestDocsInfoCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/docx/v1/documents/doc1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("unexpected query: %q", r.URL.RawQuery)
		}
		w.Header().Set("Content-Type", "application/json")
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"info", "doc1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs info error: %v", err)
	}

	if !strings.Contains(buf.String(), "document_id\tdoc1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "title\tSpecs") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "url\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsInfoCommandMissingDocIDDoesNotCallHTTP(t *testing.T) {
	called := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		called = true
		t.Fatalf("unexpected http call: %s %s", r.Method, r.URL.Path)
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"info"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if called {
		t.Fatal("unexpected http call")
	}
}

func TestDocsInfoCommandFallbackURL(t *testing.T) {
	docxCalled := false
	driveCalled := false
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		switch r.URL.Path {
		case "/open-apis/docx/v1/documents/doc1":
			docxCalled = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"document": map[string]any{
						"document_id": "doc1",
						"title":       "Specs",
					},
				},
			})
		case "/open-apis/drive/v1/files/doc1":
			driveCalled = true
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"file": map[string]any{
						"token": "doc1",
						"name":  "Specs",
						"type":  "docx",
						"url":   "https://example.com/doc",
					},
				},
			})
		default:
			t.Fatalf("unexpected path: %s", r.URL.Path)
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"info", "--doc-id", "doc1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs info error: %v", err)
	}
	if !docxCalled {
		t.Fatal("docx endpoint not called")
	}
	if !driveCalled {
		t.Fatal("drive metadata endpoint not called")
	}
	if !strings.Contains(buf.String(), "document_id\tdoc1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "title\tSpecs") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "url\thttps://example.com/doc") {
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
		if r.URL.RawQuery != "" {
			t.Fatalf("unexpected query: %q", r.URL.RawQuery)
		}
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/drive/v1/export_tasks":
			w.Header().Set("Content-Type", "application/json")
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
			w.Header().Set("Content-Type", "application/json")
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
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "export.pdf")
	prevInterval := exportTaskPollInterval
	exportTaskPollInterval = 0
	defer func() {
		exportTaskPollInterval = prevInterval
	}()
	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"export", "doc1", "--format", "pdf", "--out", outPath})
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

func TestDocsInfoCommandRequiresSDK(t *testing.T) {
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"info", "doc1"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDocsGetCommand(t *testing.T) {
	exported := []byte("Hello doc\nLine two\n")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.RawQuery != "" {
			t.Fatalf("unexpected query: %q", r.URL.RawQuery)
		}
		if r.Method != http.MethodGet || r.URL.Path != "/open-apis/docx/v1/documents/doc1/raw_content" {
			t.Fatalf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"content": string(exported),
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"get", "doc1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs get error: %v", err)
	}

	if got := buf.String(); got != string(exported) {
		t.Fatalf("unexpected output: %q", got)
	}
}

func TestDocsGetBlocksCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.Method != http.MethodGet || r.URL.Path != "/open-apis/docx/v1/documents/doc1/blocks" {
			t.Fatalf("unexpected path: %s %s", r.Method, r.URL.Path)
		}
		if got := r.URL.Query().Get("page_size"); got == "" {
			t.Fatalf("expected page_size query param")
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"block_id":   "block1",
						"block_type": 2,
						"text": map[string]any{
							"elements": []map[string]any{
								{"text_run": map[string]any{"content": "hello"}},
							},
						},
					},
				},
				"has_more":   false,
				"page_token": "",
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"get", "doc1", "--format", "blocks"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs get blocks error: %v", err)
	}

	if !strings.Contains(buf.String(), "block1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsExportCommandRequiresSDK(t *testing.T) {
	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "export.pdf")

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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"export", "doc1", "--format", "pdf", "--out", outPath})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "sdk client is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}
