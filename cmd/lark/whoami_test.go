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

func TestWhoamiUserToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/authen/v1/user_info" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer user-token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"name":             "Ada",
				"en_name":          "Ada",
				"user_id":          "u1",
				"open_id":          "o1",
				"union_id":         "un1",
				"tenant_key":       "t1",
				"email":            "ada@example.com",
				"enterprise_email": "ada@corp.example",
				"mobile":           "+1",
				"employee_no":      "E123",
				"avatar_url":       "https://example.com/ada.png",
				"avatar_thumb":     "https://example.com/ada-thumb.png",
				"avatar_middle":    "https://example.com/ada-middle.png",
				"avatar_big":       "https://example.com/ada-big.png",
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	var buf bytes.Buffer
	state := &appState{
		TokenType: "user",
		Config: &config.Config{
			AppID:     "app",
			AppSecret: "secret",
			BaseURL:   baseURL,
		},
		Printer: output.Printer{Writer: &buf, JSON: true},
	}
	withUserAccount(state.Config, defaultUserAccountName, "user-token", "", time.Now().Add(2*time.Hour).Unix(), "")
	sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
	if err != nil {
		t.Fatalf("sdk client error: %v", err)
	}
	state.SDK = sdkClient

	cmd := newWhoamiCmd(state)
	if err := cmd.Execute(); err != nil {
		t.Fatalf("whoami error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}
	if payload["name"] != "Ada" {
		t.Fatalf("expected name, got %v", payload["name"])
	}
	if payload["tenant_key"] != "t1" {
		t.Fatalf("expected tenant_key, got %v", payload["tenant_key"])
	}
	if payload["user_id"] != "u1" {
		t.Fatalf("expected user_id, got %v", payload["user_id"])
	}
	if payload["open_id"] != "o1" {
		t.Fatalf("expected open_id, got %v", payload["open_id"])
	}
	if payload["union_id"] != "un1" {
		t.Fatalf("expected union_id, got %v", payload["union_id"])
	}
	if payload["email"] != "ada@example.com" {
		t.Fatalf("expected email, got %v", payload["email"])
	}
}
