package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestWikiMemberAdd(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		Config: &config.Config{
			AppID:                      "app",
			AppSecret:                  "secret",
			TenantAccessToken:          "tenant-token",
			TenantAccessTokenExpiresAt: time.Now().Add(2 * time.Hour).Unix(),
		},
		Printer: output.Printer{Writer: &buf, JSON: true},
	}

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("method=%s", r.Method)
		}
		if r.URL.Path != "/open-apis/wiki/v2/spaces/s1/members" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer tenant-token" {
			t.Fatalf("authorization=%q", got)
		}
		if got := r.URL.Query().Get("need_notification"); got != "true" {
			t.Fatalf("need_notification=%q", got)
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["member_type"] != "userid" {
			t.Fatalf("member_type=%v", body["member_type"])
		}
		if body["member_id"] != "u1" {
			t.Fatalf("member_id=%v", body["member_id"])
		}
		if body["member_role"] != "member" {
			t.Fatalf("member_role=%v", body["member_role"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"member": map[string]any{
					"member_type": "userid",
					"member_id":   "u1",
					"member_role": "member",
					"type":        "member",
				},
			},
		})
	})

	httpClient, baseURL := testutil.NewTestClient(handler)
	state.Config.BaseURL = baseURL
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "add", "--space-id", "s1", "userid", "u1", "--role", "member", "--need-notification"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("wiki member add error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v", err)
	}
	if payload["member"] == nil {
		t.Fatalf("expected member in output, got %v", payload)
	}
}
