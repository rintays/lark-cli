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

func TestWikiMemberDeleteCommandRequiresMemberID(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call")
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

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "delete", "--space-id", "spc1", "userid"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"member-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWikiMemberDeleteCommandRequiresMemberType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("unexpected HTTP call")
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

	cmd := newWikiCmd(state)
	cmd.SetArgs([]string{"member", "delete", "--space-id", "spc1", "--member-id", "u1"})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"member-type\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestWikiMemberDeleteCommandUsesV2EndpointAndOutputsJSON(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/wiki/v2/spaces/spc1/members/u1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode body: %v", err)
		}
		if body["member_type"] != "userid" {
			t.Fatalf("unexpected member_type: %#v", body["member_type"])
		}
		if body["member_id"] != "u1" {
			t.Fatalf("unexpected member_id: %#v", body["member_id"])
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
					"type":        "user",
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
	cmd.SetArgs([]string{"member", "delete", "--space-id", "spc1", "userid", "u1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("wiki member delete error: %v", err)
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%q", err, buf.String())
	}
	if payload["deleted"] != true {
		t.Fatalf("expected deleted=true, got %#v", payload["deleted"])
	}
	member, ok := payload["member"].(map[string]any)
	if !ok {
		t.Fatalf("expected member object, got %#v", payload["member"])
	}
	if member["member_type"] != "userid" || member["member_id"] != "u1" {
		t.Fatalf("unexpected member: %#v", member)
	}
}
