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

func TestMeetingInfoCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/vc/v1/meetings/meet_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("with_participants") != "true" {
			t.Fatalf("unexpected with_participants: %s", r.URL.Query().Get("with_participants"))
		}
		if r.URL.Query().Get("with_meeting_ability") != "true" {
			t.Fatalf("unexpected with_meeting_ability: %s", r.URL.Query().Get("with_meeting_ability"))
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		if r.URL.Query().Get("query_mode") != "1" {
			t.Fatalf("unexpected query_mode: %s", r.URL.Query().Get("query_mode"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"meeting": map[string]any{
					"id":         "meet_1",
					"topic":      "Demo",
					"start_time": "1700000000",
					"end_time":   "1700003600",
					"status":     2,
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

	cmd := newMeetingsCmd(state)
	if cmd.Use != "meetings" {
		t.Fatalf("expected command name meetings, got %s", cmd.Use)
	}
	cmd.SetArgs([]string{
		"info",
		"meet_1",
		"--with-participants",
		"--with-ability",
		"--user-id-type", "open_id",
		"--query-mode", "1",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("meetings info error: %v", err)
	}

	if !strings.Contains(buf.String(), "meet_1\tDemo\t2\t1700000000\t1700003600") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMeetingInfoRequiresMeetingID(t *testing.T) {
	cmd := newMeetingsCmd(&appState{})
	cmd.SetArgs([]string{"info"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "required flag(s) \"meeting-id\" not set" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMeetingListCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/open-apis/vc/v1/meeting_list" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			if r.Header.Get("Authorization") != "Bearer token" {
				t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"meeting_list": []map[string]any{
						{
							"meeting_id":         "meet_1",
							"meeting_topic":      "Weekly Sync",
							"meeting_status":     2,
							"meeting_start_time": "1700000000",
							"meeting_end_time":   "1700003600",
						},
					},
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
		}
		sdkClient, err := larksdk.New(state.Config, larksdk.WithHTTPClient(httpClient))
		if err != nil {
			t.Fatalf("sdk client error: %v", err)
		}
		state.SDK = sdkClient

		cmd := newMeetingsCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2"})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("meetings list error: %v", err)
		}

		if !strings.Contains(buf.String(), "meet_1\tWeekly Sync\t2\t1700000000\t1700003600") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires sdk client", func(t *testing.T) {
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

		cmd := newMeetingsCmd(state)
		cmd.SetArgs([]string{"list", "--limit", "2"})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "sdk client is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}
