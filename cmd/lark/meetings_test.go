package main

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strconv"
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
	if err.Error() != "meeting-id is required" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestMeetingListCommand(t *testing.T) {
	t.Run("uses sdk client", func(t *testing.T) {
		startTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
		startUnix := strconv.FormatInt(startTime.Unix(), 10)
		endUnix := strconv.FormatInt(endTime.Unix(), 10)
		handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Path != "/open-apis/vc/v1/meeting_list" {
				t.Fatalf("unexpected path: %s", r.URL.Path)
			}
			if r.URL.Query().Get("start_time") != startUnix {
				t.Fatalf("unexpected start_time: %s", r.URL.Query().Get("start_time"))
			}
			if r.URL.Query().Get("end_time") != endUnix {
				t.Fatalf("unexpected end_time: %s", r.URL.Query().Get("end_time"))
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
		cmd.SetArgs([]string{"list", "--limit", "2", "--start", startTime.Format(time.RFC3339), "--end", endTime.Format(time.RFC3339)})
		if err := cmd.Execute(); err != nil {
			t.Fatalf("meetings list error: %v", err)
		}

		if !strings.Contains(buf.String(), "meet_1\tWeekly Sync\t2\t1700000000\t1700003600") {
			t.Fatalf("unexpected output: %q", buf.String())
		}
	})

	t.Run("requires sdk client", func(t *testing.T) {
		startTime := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
		endTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
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
		cmd.SetArgs([]string{"list", "--limit", "2", "--start", startTime.Format(time.RFC3339), "--end", endTime.Format(time.RFC3339)})
		err := cmd.Execute()
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "sdk client is required" {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestMeetingCreateCommand(t *testing.T) {
	endTime := time.Date(2024, time.January, 2, 0, 0, 0, 0, time.UTC)
	endUnix := strconv.FormatInt(endTime.Unix(), 10)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/vc/v1/reserves/apply" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload error: %v", err)
		}
		if payload["end_time"] != endUnix {
			t.Fatalf("unexpected end_time: %#v", payload["end_time"])
		}
		if payload["owner_id"] != "ou_owner" {
			t.Fatalf("unexpected owner_id: %#v", payload["owner_id"])
		}
		settings, ok := payload["meeting_settings"].(map[string]any)
		if !ok {
			t.Fatalf("expected meeting_settings map, got %#v", payload["meeting_settings"])
		}
		if settings["topic"] != "Weekly" {
			t.Fatalf("unexpected topic: %#v", settings["topic"])
		}
		if settings["auto_record"] != true {
			t.Fatalf("unexpected auto_record: %#v", settings["auto_record"])
		}
		if int(settings["meeting_initial_type"].(float64)) != 1 {
			t.Fatalf("unexpected meeting_initial_type: %#v", settings["meeting_initial_type"])
		}
		if settings["password"] != "1234" {
			t.Fatalf("unexpected password: %#v", settings["password"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"reserve": map[string]any{
					"id":         "res_1",
					"meeting_no": "123456789",
					"url":        "https://meetings.feishu.cn/s/demo",
					"end_time":   endUnix,
					"meeting_settings": map[string]any{
						"topic": "Weekly",
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

	cmd := newMeetingsCmd(state)
	cmd.SetArgs([]string{
		"create",
		"--end-time", endTime.Format(time.RFC3339),
		"--topic", "Weekly",
		"--auto-record",
		"--meeting-initial-type", "1",
		"--password", "1234",
		"--owner-id", "ou_owner",
		"--user-id-type", "open_id",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("meetings create error: %v", err)
	}

	if !strings.Contains(buf.String(), "res_1\t123456789\tWeekly\t"+endUnix+"\thttps://meetings.feishu.cn/s/demo") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMeetingUpdateCommand(t *testing.T) {
	endTime := time.Date(2024, time.January, 3, 0, 0, 0, 0, time.UTC)
	endUnix := strconv.FormatInt(endTime.Unix(), 10)
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/vc/v1/reserves/res_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload error: %v", err)
		}
		if payload["end_time"] != endUnix {
			t.Fatalf("unexpected end_time: %#v", payload["end_time"])
		}
		settings, ok := payload["meeting_settings"].(map[string]any)
		if !ok {
			t.Fatalf("expected meeting_settings map, got %#v", payload["meeting_settings"])
		}
		if settings["topic"] != "Updated" {
			t.Fatalf("unexpected topic: %#v", settings["topic"])
		}
		if settings["auto_record"] != true {
			t.Fatalf("unexpected auto_record: %#v", settings["auto_record"])
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"reserve": map[string]any{
					"id":         "res_1",
					"meeting_no": "123456789",
					"url":        "https://meetings.feishu.cn/s/demo",
					"end_time":   endUnix,
					"meeting_settings": map[string]any{
						"topic": "Updated",
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

	cmd := newMeetingsCmd(state)
	cmd.SetArgs([]string{
		"update",
		"res_1",
		"--end-time", endTime.Format(time.RFC3339),
		"--topic", "Updated",
		"--auto-record",
		"--user-id-type", "open_id",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("meetings update error: %v", err)
	}

	if !strings.Contains(buf.String(), "res_1\t123456789\tUpdated\t"+endUnix+"\thttps://meetings.feishu.cn/s/demo") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestMeetingDeleteCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			t.Fatalf("expected DELETE, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/vc/v1/reserves/res_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("unexpected authorization: %s", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
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
	cmd.SetArgs([]string{"delete", "res_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("meetings delete error: %v", err)
	}

	if !strings.Contains(buf.String(), "true\tres_1") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
