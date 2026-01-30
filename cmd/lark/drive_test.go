package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larkapi"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestDriveListCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/drive/v1/files" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		if r.URL.Query().Get("folder_token") != "root" {
			t.Fatalf("unexpected folder_token: %s", r.URL.Query().Get("folder_token"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files":           []map[string]any{{"token": "f1", "name": "Doc", "type": "docx", "url": "https://example.com/doc"}},
				"has_more":        false,
				"next_page_token": "",
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"list", "--folder-id", "root", "--limit", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive list error: %v", err)
	}

	if !strings.Contains(buf.String(), "f1\tDoc\tdocx\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveSearchCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/drive/v1/files/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["query"] != "budget" {
			t.Fatalf("unexpected query: %+v", payload)
		}
		if payload["page_size"].(float64) != 2 {
			t.Fatalf("unexpected page_size: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files":    []map[string]any{{"token": "f2", "name": "Budget", "type": "sheet", "url": "https://example.com/sheet"}},
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"search", "--query", "budget", "--limit", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive search error: %v", err)
	}

	if !strings.Contains(buf.String(), "f2\tBudget\tsheet\thttps://example.com/sheet") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveGetCommandWithSDK(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/drive/v1/files/f1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"file": map[string]any{"token": "f1", "name": "Doc", "type": "docx", "url": "https://example.com/doc"},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	legacyClient := &http.Client{Transport: testutil.HandlerRoundTripper{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("legacy client used for drive get")
	})}}

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
		Client:  &larkapi.Client{BaseURL: "http://legacy.test", HTTPClient: legacyClient},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"get", "f1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive get error: %v", err)
	}

	if !strings.Contains(buf.String(), "f1\tDoc\tdocx\thttps://example.com/doc") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveURLsCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/open-apis/drive/v1/files/f1":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"file": map[string]any{"token": "f1", "name": "Doc", "url": "https://example.com/doc"},
				},
			})
		case "/open-apis/drive/v1/files/f2":
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"file": map[string]any{"token": "f2", "name": "Sheet", "url": "https://example.com/sheet"},
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"urls", "f1", "f2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive urls error: %v", err)
	}

	output := buf.String()
	if !strings.Contains(output, "f1\thttps://example.com/doc\tDoc") {
		t.Fatalf("unexpected output: %q", output)
	}
	if !strings.Contains(output, "f2\thttps://example.com/sheet\tSheet") {
		t.Fatalf("unexpected output: %q", output)
	}
}

func TestDriveShareCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/public" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
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
		Client:  &larkapi.Client{BaseURL: baseURL, HTTPClient: httpClient},
	}

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"share", "f1", "--type", "docx", "--link-share", "tenant_readable", "--external-access"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive share error: %v", err)
	}

	if !strings.Contains(buf.String(), "f1\tdocx\ttenant_readable\ttrue\tfalse") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveUploadCommand(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "lark-upload-*.txt")
	if err != nil {
		t.Fatalf("create temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())
	content := []byte("hello upload")
	if _, err := tmpFile.Write(content); err != nil {
		t.Fatalf("write temp file: %v", err)
	}
	if err := tmpFile.Close(); err != nil {
		t.Fatalf("close temp file: %v", err)
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/drive/v1/files/upload_all":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("missing auth header")
			}
			if !strings.HasPrefix(r.Header.Get("Content-Type"), "multipart/form-data; boundary=") {
				t.Fatalf("unexpected content type: %s", r.Header.Get("Content-Type"))
			}
			if err := r.ParseMultipartForm(4 << 20); err != nil {
				t.Fatalf("parse multipart: %v", err)
			}
			if r.FormValue("file_name") != "report.txt" {
				t.Fatalf("unexpected file_name: %s", r.FormValue("file_name"))
			}
			if r.FormValue("parent_type") != "explorer" {
				t.Fatalf("unexpected parent_type: %s", r.FormValue("parent_type"))
			}
			if r.FormValue("parent_node") != "fld_123" {
				t.Fatalf("unexpected parent_node: %s", r.FormValue("parent_node"))
			}
			if r.FormValue("size") != fmt.Sprintf("%d", len(content)) {
				t.Fatalf("unexpected size: %s", r.FormValue("size"))
			}
			files := r.MultipartForm.File["file"]
			if len(files) != 1 {
				t.Fatalf("expected 1 file part, got %d", len(files))
			}
			part, err := files[0].Open()
			if err != nil {
				t.Fatalf("open file part: %v", err)
			}
			defer part.Close()
			data, err := io.ReadAll(part)
			if err != nil {
				t.Fatalf("read file part: %v", err)
			}
			if string(data) != string(content) {
				t.Fatalf("unexpected file content: %q", string(data))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"file_token": "file_123",
				},
			})
		case "/open-apis/drive/v1/files/file_123":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"file": map[string]any{
						"token": "file_123",
						"name":  "report.txt",
						"type":  "file",
						"url":   "https://example.com/file",
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

	cmd := newDriveCmd(state)
	cmd.SetArgs([]string{"upload", "--file", tmpFile.Name(), "--folder-token", "fld_123", "--name", "report.txt"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive upload error: %v", err)
	}

	if !strings.Contains(buf.String(), "file_123\treport.txt\tfile\thttps://example.com/file") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveDownloadCommand(t *testing.T) {
	content := []byte("downloaded bytes")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/f1/download" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write(content)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	outDir := t.TempDir()
	outPath := filepath.Join(outDir, "download.txt")

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
	cmd.SetArgs([]string{"download", "--file-token", "f1", "--out", outPath})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("drive download error: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("unexpected file contents: %q", string(data))
	}
	if !strings.Contains(buf.String(), outPath) {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDriveExportCommand(t *testing.T) {
	cases := []struct {
		name    string
		useJSON bool
	}{
		{name: "text"},
		{name: "json", useJSON: true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			exported := []byte("exported bytes")
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
					if payload["token"] != "f1" {
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
					_ = json.NewEncoder(w).Encode(map[string]any{
						"code": 0,
						"msg":  "ok",
						"data": map[string]any{
							"result": map[string]any{
								"file_extension": "pdf",
								"type":           "docx",
								"file_name":      "Export.pdf",
								"file_token":     "file1",
								"file_size":      int64(len(exported)),
								"job_error_msg":  "success",
								"job_status":     0,
							},
						},
					})
				case r.Method == http.MethodGet && r.URL.Path == "/open-apis/drive/v1/export_tasks/file/file1/download":
					_, _ = w.Write(exported)
				default:
					t.Fatalf("unexpected path: %s %s", r.Method, r.URL.Path)
				}
			})
			httpClient, baseURL := testutil.NewTestClient(handler)
			legacyClient := &http.Client{Transport: testutil.HandlerRoundTripper{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				t.Fatalf("legacy client used for drive export")
			})}}

			outDir := t.TempDir()
			outPath := filepath.Join(outDir, "export.pdf")

			var buf bytes.Buffer
			state := &appState{
				Config: &config.Config{
					AppID:                      "app",
					AppSecret:                  "secret",
					BaseURL:                    baseURL,
					TenantAccessToken:          "token",
					TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
				},
				JSON:    tc.useJSON,
				Printer: output.Printer{Writer: &buf, JSON: tc.useJSON},
				Client:  &larkapi.Client{BaseURL: "http://legacy.test", HTTPClient: legacyClient},
			}
			sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
			if err != nil {
				t.Fatalf("sdk client error: %v", err)
			}
			state.SDK = sdkClient

			prevInterval := exportTaskPollInterval
			exportTaskPollInterval = 0
			defer func() {
				exportTaskPollInterval = prevInterval
			}()

			cmd := newDriveCmd(state)
			cmd.SetArgs([]string{"export", "--file-token", "f1", "--type", "docx", "--format", "pdf", "--out", outPath})
			if err := cmd.Execute(); err != nil {
				t.Fatalf("drive export error: %v", err)
			}

			data, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatalf("read exported file: %v", err)
			}
			if !bytes.Equal(data, exported) {
				t.Fatalf("unexpected export content: %q", string(data))
			}

			if tc.useJSON {
				var payload map[string]any
				if err := json.NewDecoder(bytes.NewReader(buf.Bytes())).Decode(&payload); err != nil {
					t.Fatalf("decode output: %v", err)
				}
				if payload["file_token"] != "f1" {
					t.Fatalf("unexpected file_token: %+v", payload["file_token"])
				}
				if payload["export_file_token"] != "file1" {
					t.Fatalf("unexpected export_file_token: %+v", payload["export_file_token"])
				}
			} else if !strings.Contains(buf.String(), outPath) {
				t.Fatalf("unexpected output: %q", buf.String())
			}
		})
	}
}
