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
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestContactsUserGetCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/open-apis/contact/v3/users/ou_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"user": map[string]any{
					"user_id": "u_1",
					"open_id": "ou_1",
					"name":    "Ada",
					"email":   "ada@example.com",
					"mobile":  "+1-555-0100",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)
	legacyClient := &http.Client{Transport: testutil.HandlerRoundTripper{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatalf("legacy client used for contacts user get")
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

	cmd := newContactsCmd(state)
	cmd.SetArgs([]string{"user", "get", "--open-id", "ou_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("contacts user get error: %v", err)
	}

	if !strings.Contains(buf.String(), "u_1\tAda\tada@example.com\t+1-555-0100") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
