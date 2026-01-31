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

func TestWikiMemberListCommandUsesV2EndpointAndOutputsJSON(t *testing.T) {
	calls := 0
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Method != http.MethodGet {
			t.Fatalf("unexpected method: %s", r.Method)
		}
		if r.URL.Path != "/open-apis/wiki/v2/spaces/spc1/members" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		if r.URL.Query().Get("page_size") != "1" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}

		w.Header().Set("Content-Type", "application/json")
		switch calls {
		case 1:
			if r.URL.Query().Get("page_token") != "" {
				t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"members": []map[string]any{
						{"member_type": "userid", "member_id": "u1", "member_role": "admin", "type": "user"},
					},
					"has_more":   true,
					"page_token": "t2",
				},
			})
		case 2:
			if r.URL.Query().Get("page_token") != "t2" {
				t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"members": []map[string]any{
						{"member_type": "email", "member_id": "a@example.com", "member_role": "member", "type": "user"},
					},
					"has_more":   false,
					"page_token": "",
				},
			})
		default:
			t.Fatalf("unexpected call count: %d", calls)
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
		Printer: output.Printer{Writer: &buf, JSON: true},
	}
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "list", "--space-id", "spc1", "--limit", "2", "--page-size", "1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("wiki member list error: %v", err)
	}
	if calls != 2 {
		t.Fatalf("expected 2 calls, got %d", calls)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%q", err, buf.String())
	}
	members, ok := payload["members"].([]any)
	if !ok {
		t.Fatalf("expected members array, got: %#v", payload["members"])
	}
	if len(members) != 2 {
		t.Fatalf("expected 2 members, got %d", len(members))
	}
	first, _ := members[0].(map[string]any)
	if first["member_type"] != "userid" || first["member_id"] != "u1" {
		t.Fatalf("unexpected first member: %#v", first)
	}
	second, _ := members[1].(map[string]any)
	if second["member_type"] != "email" || second["member_id"] != "a@example.com" {
		t.Fatalf("unexpected second member: %#v", second)
	}

	// Ensure we didn't accidentally print text output.
	if strings.Contains(buf.String(), "\t") {
		t.Fatalf("expected JSON-only output, got: %q", buf.String())
	}
}
