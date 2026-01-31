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

func TestDocsBlocksGetCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/docx/v1/documents/doc1/blocks/block1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"block": map[string]any{
					"block_id":   "block1",
					"block_type": 2,
					"text": map[string]any{
						"elements": []map[string]any{
							{"text_run": map[string]any{"content": "hello"}},
						},
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"blocks", "get", "doc1", "block1", "--revision-id", "0"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs blocks get error: %v", err)
	}

	if !strings.Contains(buf.String(), "block_id") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "block1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "hello") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsMarkdownConvertCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/docx/v1/documents/blocks/convert" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["content_type"] != "markdown" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if payload["content"] != "# Title" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"first_level_block_ids": []string{"tmp1"},
				"blocks": []map[string]any{
					{
						"block_id":   "tmp1",
						"block_type": 3,
						"heading1": map[string]any{
							"elements": []map[string]any{
								{"text_run": map[string]any{"content": "Title"}},
							},
						},
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

	cmd := newDocsCmd(state)
	cmd.SetArgs([]string{"markdown", "convert", "--content", "# Title"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs markdown convert error: %v", err)
	}

	if !strings.Contains(buf.String(), "first_level_blocks: 1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "total_blocks: 1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestDocsMarkdownOverwriteCommand(t *testing.T) {
	listCalls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/docx/v1/documents/blocks/convert":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"first_level_block_ids": []string{"tmp1"},
					"blocks": []map[string]any{
						{
							"block_id":   "tmp1",
							"block_type": 2,
							"text": map[string]any{
								"elements": []map[string]any{
									{"text_run": map[string]any{"content": "hello"}},
								},
							},
						},
					},
				},
			})
		case r.Method == http.MethodGet && r.URL.Path == "/open-apis/docx/v1/documents/doc1/blocks/doc1/children":
			listCalls++
			items := []map[string]any{}
			if listCalls == 1 {
				items = append(items, map[string]any{
					"block_id":   "old1",
					"block_type": 2,
					"text": map[string]any{
						"elements": []map[string]any{
							{"text_run": map[string]any{"content": "old"}},
						},
					},
				})
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items":      items,
					"has_more":   false,
					"page_token": "",
				},
			})
		case r.Method == http.MethodDelete && r.URL.Path == "/open-apis/docx/v1/documents/doc1/blocks/doc1/children/batch_delete":
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"document_revision_id": 2,
				},
			})
		case r.Method == http.MethodPost && r.URL.Path == "/open-apis/docx/v1/documents/doc1/blocks/doc1/descendant":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"document_revision_id": 3,
					"block_id_relations": []map[string]any{
						{"temporary_block_id": "tmp1", "block_id": "new1"},
					},
				},
			})
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
	cmd.SetArgs([]string{"markdown", "overwrite", "doc1", "--content", "hello"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs markdown overwrite error: %v", err)
	}

	if !strings.Contains(buf.String(), "doc1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "deleted_blocks") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
	if !strings.Contains(buf.String(), "inserted_blocks") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
