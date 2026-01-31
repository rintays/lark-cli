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

func TestCalendarListCommand(t *testing.T) {
	start := "2026-01-02T03:04:05Z"
	end := "2026-01-02T04:04:05Z"
	startUnix := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix()
	endUnix := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC).Unix()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events":
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			if r.URL.Query().Get("start_time") != strconv.FormatInt(startUnix, 10) {
				t.Fatalf("unexpected start_time: %s", r.URL.Query().Get("start_time"))
			}
			if r.URL.Query().Get("end_time") != strconv.FormatInt(endUnix, 10) {
				t.Fatalf("unexpected end_time: %s", r.URL.Query().Get("end_time"))
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"event_id": "evt_1",
							"summary":  "Standup",
							"start_time": map[string]any{
								"timestamp": strconv.FormatInt(startUnix, 10),
							},
							"end_time": map[string]any{
								"timestamp": strconv.FormatInt(endUnix, 10),
							},
						},
					},
					"has_more": false,
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"list", "--start", start, "--end", end, "--limit", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar list error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\t"+start+"\t"+end+"\tStandup") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCalendarCreateCommand(t *testing.T) {
	start := "2026-01-02T03:04:05Z"
	end := "2026-01-02T04:04:05Z"
	startUnix := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix()
	endUnix := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC).Unix()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events":
			w.Header().Set("Content-Type", "application/json")
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["summary"] != "Demo" {
				t.Fatalf("unexpected summary: %+v", payload["summary"])
			}
			startPayload := payload["start_time"].(map[string]any)
			if startPayload["timestamp"] != strconv.FormatInt(startUnix, 10) {
				t.Fatalf("unexpected start_time: %+v", payload["start_time"])
			}
			endPayload := payload["end_time"].(map[string]any)
			if endPayload["timestamp"] != strconv.FormatInt(endUnix, 10) {
				t.Fatalf("unexpected end_time: %+v", payload["end_time"])
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"event": map[string]any{
						"event_id": "evt_1",
						"summary":  "Demo",
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events/evt_1/attendees":
			w.Header().Set("Content-Type", "application/json")
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			attendees := payload["attendees"].([]any)
			if len(attendees) != 2 {
				t.Fatalf("unexpected attendees: %+v", attendees)
			}
			first := attendees[0].(map[string]any)
			if first["third_party_email"] != "dev@example.com" {
				t.Fatalf("unexpected attendee: %+v", first)
			}
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{
		"create",
		"--start", start,
		"--end", end,
		"--summary", "Demo",
		"--description", "Notes",
		"--attendee", "dev@example.com",
		"--attendee", "ops@example.com",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar create error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\t"+start+"\t"+end+"\tDemo") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCalendarSearchCommand(t *testing.T) {
	start := "2026-01-02T03:04:05Z"
	end := "2026-01-02T04:04:05Z"
	startUnix := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix()
	endUnix := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC).Unix()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events/search":
			if r.Method != http.MethodPost {
				t.Fatalf("expected POST, got %s", r.Method)
			}
			if r.URL.Query().Get("page_size") != "2" {
				t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["query"] != "Demo" {
				t.Fatalf("unexpected query: %+v", payload["query"])
			}
			filter := payload["filter"].(map[string]any)
			startPayload := filter["start_time"].(map[string]any)
			if startPayload["timestamp"] != strconv.FormatInt(startUnix, 10) {
				t.Fatalf("unexpected start_time: %+v", filter["start_time"])
			}
			endPayload := filter["end_time"].(map[string]any)
			if endPayload["timestamp"] != strconv.FormatInt(endUnix, 10) {
				t.Fatalf("unexpected end_time: %+v", filter["end_time"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"items": []map[string]any{
						{
							"event_id": "evt_1",
							"summary":  "Standup",
							"start_time": map[string]any{
								"timestamp": strconv.FormatInt(startUnix, 10),
							},
							"end_time": map[string]any{
								"timestamp": strconv.FormatInt(endUnix, 10),
							},
						},
					},
					"page_token": "next",
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"search", "--query", "Demo", "--start", start, "--end", end, "--limit", "2"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar search error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\t"+start+"\t"+end+"\tStandup") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCalendarGetCommand(t *testing.T) {
	start := "2026-01-02T03:04:05Z"
	end := "2026-01-02T04:04:05Z"
	startUnix := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix()
	endUnix := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC).Unix()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events/evt_1":
			if r.Method != http.MethodGet {
				t.Fatalf("expected GET, got %s", r.Method)
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"event": map[string]any{
						"event_id": "evt_1",
						"summary":  "Review",
						"start_time": map[string]any{
							"timestamp": strconv.FormatInt(startUnix, 10),
						},
						"end_time": map[string]any{
							"timestamp": strconv.FormatInt(endUnix, 10),
						},
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"get", "--event-id", "evt_1"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar get error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\t"+start+"\t"+end+"\tReview") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCalendarUpdateCommand(t *testing.T) {
	start := "2026-01-02T03:04:05Z"
	end := "2026-01-02T04:04:05Z"
	startUnix := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC).Unix()
	endUnix := time.Date(2026, 1, 2, 4, 4, 5, 0, time.UTC).Unix()

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events/evt_1":
			if r.Method != http.MethodPatch {
				t.Fatalf("expected PATCH, got %s", r.Method)
			}
			var payload map[string]any
			if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
				t.Fatalf("decode payload: %v", err)
			}
			if payload["summary"] != "Updated" {
				t.Fatalf("unexpected summary: %+v", payload["summary"])
			}
			startPayload := payload["start_time"].(map[string]any)
			if startPayload["timestamp"] != strconv.FormatInt(startUnix, 10) {
				t.Fatalf("unexpected start_time: %+v", payload["start_time"])
			}
			endPayload := payload["end_time"].(map[string]any)
			if endPayload["timestamp"] != strconv.FormatInt(endUnix, 10) {
				t.Fatalf("unexpected end_time: %+v", payload["end_time"])
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"event": map[string]any{
						"event_id": "evt_1",
						"summary":  "Updated",
						"start_time": map[string]any{
							"timestamp": strconv.FormatInt(startUnix, 10),
						},
						"end_time": map[string]any{
							"timestamp": strconv.FormatInt(endUnix, 10),
						},
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"update", "--event-id", "evt_1", "--summary", "Updated", "--start", start, "--end", end})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar update error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\t"+start+"\t"+end+"\tUpdated") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}

func TestCalendarDeleteCommand(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/open-apis/calendar/v4/calendars/primary":
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"calendars": []map[string]any{
						{
							"calendar": map[string]any{
								"calendar_id": "cal_1",
							},
						},
					},
				},
			})
		case "/open-apis/calendar/v4/calendars/cal_1/events/evt_1":
			if r.Method != http.MethodDelete {
				t.Fatalf("expected DELETE, got %s", r.Method)
			}
			if r.URL.Query().Get("need_notification") != "false" {
				t.Fatalf("unexpected need_notification: %s", r.URL.Query().Get("need_notification"))
			}
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(map[string]any{
				"code": 0,
				"msg":  "ok",
				"data": map[string]any{
					"event_id": "evt_1",
					"deleted":  true,
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

	cmd := newCalendarCmd(state)
	cmd.SetArgs([]string{"delete", "--event-id", "evt_1", "--notify=false"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("calendar delete error: %v", err)
	}

	if !strings.Contains(buf.String(), "evt_1\ttrue") {
		t.Fatalf("unexpected output: %q", buf.String())
	}
}
