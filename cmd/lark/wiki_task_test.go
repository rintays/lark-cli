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

func TestWikiTaskCommandsRegistered(t *testing.T) {
	cmd := newRootCmd()
	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"wiki", "task", "info", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("help failed: %v", err)
	}
}

func TestWikiTaskInfoCommandUsesV2EndpointAndOutputsJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/wiki/v2/tasks/t1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("task_type") != "move" {
			t.Fatalf("unexpected task_type: %s", r.URL.Query().Get("task_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"task": map[string]any{
					"task_id": "t1",
					"move_result": []map[string]any{
						{
							"status":     0,
							"status_msg": "success",
							"node": map[string]any{
								"space_id":          "spc1",
								"node_token":        "n1",
								"obj_token":         "d1",
								"obj_type":          "docx",
								"title":             "Doc",
								"has_child":         false,
								"node_type":         "origin",
								"parent_node_token": "",
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
		Printer: output.Printer{Writer: &buf, JSON: true},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"task", "info", "--task-id", "t1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("wiki task info error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%q", err, buf.String())
	}
	task, ok := payload["task"].(map[string]any)
	if !ok {
		t.Fatalf("expected task object, got: %#v", payload["task"])
	}
	if task["task_id"] != "t1" {
		t.Fatalf("unexpected task_id: %#v", task["task_id"])
	}
	moveResult, ok := task["move_result"].([]any)
	if !ok || len(moveResult) != 1 {
		t.Fatalf("unexpected move_result: %#v", task["move_result"])
	}

	// Ensure we didn't accidentally print text output.
	if strings.Contains(buf.String(), "\t") {
		t.Fatalf("expected JSON-only output, got: %q", buf.String())
	}
}
