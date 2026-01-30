package larkapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"

	"lark/internal/testutil"
)

func TestTenantAccessToken(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/auth/v3/tenant_access_token/internal/" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["app_id"] != "app" || payload["app_secret"] != "secret" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code":                0,
			"msg":                 "ok",
			"tenant_access_token": "token",
			"expire":              7200,
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, AppID: "app", AppSecret: "secret", HTTPClient: httpClient}
	gotToken, gotExpire, err := client.TenantAccessToken(context.Background())
	if err != nil {
		t.Fatalf("TenantAccessToken error: %v", err)
	}
	if gotToken != "token" {
		t.Fatalf("expected token, got %s", gotToken)
	}
	if gotExpire != 7200 {
		t.Fatalf("expected expire 7200, got %d", gotExpire)
	}
}

func TestWhoAmI(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/tenant/v2/tenant/query" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"tenant": map[string]string{
					"tenant_key": "tkey",
					"name":       "Tenant",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	info, err := client.WhoAmI(context.Background(), "token")
	if err != nil {
		t.Fatalf("WhoAmI error: %v", err)
	}
	if info.TenantKey != "tkey" || info.Name != "Tenant" {
		t.Fatalf("unexpected tenant info: %+v", info)
	}
}

func TestPrimaryCalendar(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/calendar/v4/calendars/primary" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"calendar": map[string]any{
					"calendar_id": "cal_1",
					"summary":     "Primary",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	cal, err := client.PrimaryCalendar(context.Background(), "token")
	if err != nil {
		t.Fatalf("PrimaryCalendar error: %v", err)
	}
	if cal.CalendarID != "cal_1" || cal.Summary != "Primary" {
		t.Fatalf("unexpected calendar: %+v", cal)
	}
}

func TestListCalendarEvents(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/calendar/v4/calendars/cal_1/events" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("start_time") != "1700000000" {
			t.Fatalf("unexpected start_time: %s", r.URL.Query().Get("start_time"))
		}
		if r.URL.Query().Get("end_time") != "1700003600" {
			t.Fatalf("unexpected end_time: %s", r.URL.Query().Get("end_time"))
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"event_id": "evt_1",
						"summary":  "Standup",
						"start_time": map[string]any{
							"timestamp": "1700000000",
						},
						"end_time": map[string]any{
							"timestamp": "1700003600",
						},
					},
				},
				"has_more": false,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.ListCalendarEvents(context.Background(), "token", ListCalendarEventsRequest{
		CalendarID: "cal_1",
		StartTime:  "1700000000",
		EndTime:    "1700003600",
		PageSize:   2,
	})
	if err != nil {
		t.Fatalf("ListCalendarEvents error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].EventID != "evt_1" {
		t.Fatalf("unexpected events: %+v", result.Items)
	}
}

func TestCreateCalendarEvent(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/calendar/v4/calendars/cal_1/events" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["summary"] != "Demo" {
			t.Fatalf("unexpected summary: %+v", payload["summary"])
		}
		start := payload["start_time"].(map[string]any)
		if start["timestamp"] != "1700000000" {
			t.Fatalf("unexpected start_time: %+v", payload["start_time"])
		}
		end := payload["end_time"].(map[string]any)
		if end["timestamp"] != "1700003600" {
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
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	event, err := client.CreateCalendarEvent(context.Background(), "token", CreateCalendarEventRequest{
		CalendarID: "cal_1",
		Summary:    "Demo",
		StartTime:  1700000000,
		EndTime:    1700003600,
	})
	if err != nil {
		t.Fatalf("CreateCalendarEvent error: %v", err)
	}
	if event.EventID != "evt_1" || event.Summary != "Demo" {
		t.Fatalf("unexpected event: %+v", event)
	}
}

func TestCreateCalendarEventAttendees(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/calendar/v4/calendars/cal_1/events/evt_1/attendees" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		attendees := payload["attendees"].([]any)
		first := attendees[0].(map[string]any)
		if first["type"] != "third_party" || first["third_party_email"] != "dev@example.com" {
			t.Fatalf("unexpected attendee: %+v", first)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	err := client.CreateCalendarEventAttendees(context.Background(), "token", CreateCalendarEventAttendeesRequest{
		CalendarID: "cal_1",
		EventID:    "evt_1",
		Attendees: []CalendarEventAttendee{
			{Type: "third_party", ThirdPartyEmail: "dev@example.com"},
		},
	})
	if err != nil {
		t.Fatalf("CreateCalendarEventAttendees error: %v", err)
	}
}

func TestGetMeeting(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
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
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"meeting": map[string]any{
					"id":           "meet_1",
					"topic":        "Demo",
					"meeting_no":   "123456789",
					"start_time":   "1700000000",
					"end_time":     "1700003600",
					"status":       2,
					"host_user":    map[string]any{"id": "ou_1", "user_type": 1},
					"participants": []map[string]any{{"id": "ou_2", "user_type": 1, "status": 2}},
					"ability":      map[string]any{"use_video": true},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	meeting, err := client.GetMeeting(context.Background(), "token", GetMeetingRequest{
		MeetingID:          "meet_1",
		WithParticipants:   true,
		WithMeetingAbility: true,
		UserIDType:         "open_id",
		QueryMode:          1,
	})
	if err != nil {
		t.Fatalf("GetMeeting error: %v", err)
	}
	if meeting.ID != "meet_1" || meeting.Topic != "Demo" || meeting.Status != 2 {
		t.Fatalf("unexpected meeting: %+v", meeting)
	}
	if len(meeting.Participants) != 1 || meeting.Participants[0].ID != "ou_2" {
		t.Fatalf("unexpected participants: %+v", meeting.Participants)
	}
	if !meeting.Ability.UseVideo {
		t.Fatalf("unexpected meeting ability: %+v", meeting.Ability)
	}
}

func TestSendMessage(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/im/v1/messages" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("receive_id_type") != "chat_id" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["receive_id"] != "chat" || payload["msg_type"] != "text" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if !strings.Contains(payload["content"], "hello") {
			t.Fatalf("unexpected content: %s", payload["content"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]string{"message_id": "mid"},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	messageID, err := client.SendMessage(context.Background(), "token", MessageRequest{ReceiveID: "chat", Text: "hello"})
	if err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
	if messageID != "mid" {
		t.Fatalf("expected message_id, got %s", messageID)
	}
}

func TestSendMessageWithReceiveIDType(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("receive_id_type") != "email" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["receive_id"] != "dev@example.com" {
			t.Fatalf("unexpected receive_id: %s", payload["receive_id"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]string{"message_id": "mid"},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	_, err := client.SendMessage(context.Background(), "token", MessageRequest{
		ReceiveID:     "dev@example.com",
		ReceiveIDType: "email",
		Text:          "hello",
	})
	if err != nil {
		t.Fatalf("SendMessage error: %v", err)
	}
}

func TestListChats(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/im/v1/chats" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		if r.URL.Query().Get("page_token") != "next" {
			t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{"chat_id": "c1", "name": "Chat One"},
					{"chat_id": "c2", "name": "Chat Two"},
				},
				"page_token": "token",
				"has_more":   true,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.ListChats(context.Background(), "token", ListChatsRequest{
		PageSize:  2,
		PageToken: "next",
	})
	if err != nil {
		t.Fatalf("ListChats error: %v", err)
	}
	if len(result.Items) != 2 {
		t.Fatalf("expected 2 chats, got %d", len(result.Items))
	}
	if result.Items[0].ChatID != "c1" || result.Items[0].Name != "Chat One" {
		t.Fatalf("unexpected chat: %+v", result.Items[0])
	}
	if !result.HasMore || result.PageToken != "token" {
		t.Fatalf("unexpected pagination: %+v", result)
	}
}

func TestBatchGetUserIDs(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/contact/v3/users/batch_get_id" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string][]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["emails"][0] != "dev@example.com" {
			t.Fatalf("unexpected emails: %+v", payload["emails"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"user_list": []map[string]any{
					{"user_id": "u1", "email": "dev@example.com"},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	users, err := client.BatchGetUserIDs(context.Background(), "token", BatchGetUserIDRequest{
		Emails: []string{"dev@example.com"},
	})
	if err != nil {
		t.Fatalf("BatchGetUserIDs error: %v", err)
	}
	if len(users) != 1 || users[0].UserID != "u1" {
		t.Fatalf("unexpected users: %+v", users)
	}
}

func TestGetContactUser(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/contact/v3/users/ou_1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("user_id_type") != "open_id" {
			t.Fatalf("unexpected user_id_type: %s", r.URL.Query().Get("user_id_type"))
		}
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

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	user, err := client.GetContactUser(context.Background(), "token", GetContactUserRequest{
		UserID:     "ou_1",
		UserIDType: "open_id",
	})
	if err != nil {
		t.Fatalf("GetContactUser error: %v", err)
	}
	if user.OpenID != "ou_1" || user.Email != "ada@example.com" {
		t.Fatalf("unexpected user: %+v", user)
	}
}

func TestListUsersByDepartment(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.URL.Path != "/open-apis/contact/v3/users/find_by_department" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("department_id") != "0" {
			t.Fatalf("unexpected department_id: %s", r.URL.Query().Get("department_id"))
		}
		if r.URL.Query().Get("page_size") != "1" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items":    []map[string]any{{"user_id": "u1", "name": "Ada Lovelace"}},
				"has_more": false,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.ListUsersByDepartment(context.Background(), "token", ListUsersByDepartmentRequest{
		DepartmentID: "0",
		PageSize:     1,
	})
	if err != nil {
		t.Fatalf("ListUsersByDepartment error: %v", err)
	}
	if len(result.Items) != 1 || result.Items[0].Name != "Ada Lovelace" {
		t.Fatalf("unexpected users: %+v", result.Items)
	}
}

func TestListDriveFiles(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("folder_token") != "fld" {
			t.Fatalf("unexpected folder_token: %s", r.URL.Query().Get("folder_token"))
		}
		if r.URL.Query().Get("page_size") != "2" {
			t.Fatalf("unexpected page_size: %s", r.URL.Query().Get("page_size"))
		}
		if r.URL.Query().Get("page_token") != "next" {
			t.Fatalf("unexpected page_token: %s", r.URL.Query().Get("page_token"))
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files": []map[string]any{
					{"token": "f1", "name": "Doc", "type": "docx", "url": "https://example.com/doc"},
				},
				"page_token": "token",
				"has_more":   true,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.ListDriveFiles(context.Background(), "token", ListDriveFilesRequest{
		FolderToken: "fld",
		PageSize:    2,
		PageToken:   "next",
	})
	if err != nil {
		t.Fatalf("ListDriveFiles error: %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Token != "f1" {
		t.Fatalf("unexpected files: %+v", result.Files)
	}
	if !result.HasMore || result.PageToken != "token" {
		t.Fatalf("unexpected pagination: %+v", result)
	}
}

func TestSearchDriveFiles(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/search" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["query"] != "budget" {
			t.Fatalf("unexpected query: %+v", payload)
		}
		if payload["page_size"].(float64) != 2 {
			t.Fatalf("unexpected page_size: %+v", payload)
		}
		if payload["page_token"] != "next" {
			t.Fatalf("unexpected page_token: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"files": []map[string]any{
					{"token": "f2", "name": "Budget", "type": "sheet", "url": "https://example.com/sheet"},
				},
				"page_token": "token",
				"has_more":   false,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.SearchDriveFiles(context.Background(), "token", SearchDriveFilesRequest{
		Query:     "budget",
		PageSize:  2,
		PageToken: "next",
	})
	if err != nil {
		t.Fatalf("SearchDriveFiles error: %v", err)
	}
	if len(result.Files) != 1 || result.Files[0].Token != "f2" {
		t.Fatalf("unexpected files: %+v", result.Files)
	}
	if result.HasMore || result.PageToken != "token" {
		t.Fatalf("unexpected pagination: %+v", result)
	}
}

func TestGetDriveFileMetadata(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/fmeta" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"file": map[string]any{
					"token": "fmeta",
					"name":  "Metadata Doc",
					"type":  "docx",
					"url":   "https://example.com/meta",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	file, err := client.GetDriveFileMetadata(context.Background(), "token", "fmeta")
	if err != nil {
		t.Fatalf("GetDriveFileMetadata error: %v", err)
	}
	if file.Token != "fmeta" || file.URL != "https://example.com/meta" {
		t.Fatalf("unexpected file: %+v", file)
	}
}

func TestGetDriveFile(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/f1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"file": map[string]any{
					"token": "f1",
					"name":  "Doc",
					"type":  "docx",
					"url":   "https://example.com/doc",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	file, err := client.GetDriveFile(context.Background(), "token", "f1")
	if err != nil {
		t.Fatalf("GetDriveFile error: %v", err)
	}
	if file.Token != "f1" || file.URL != "https://example.com/doc" {
		t.Fatalf("unexpected file: %+v", file)
	}
}

func TestDownloadDriveFile(t *testing.T) {
	content := []byte("file bytes")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/files/f1/download" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write(content)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	reader, err := client.DownloadDriveFile(context.Background(), "token", "f1")
	if err != nil {
		t.Fatalf("DownloadDriveFile error: %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("unexpected download: %q", string(data))
	}
}

func TestUpdateDrivePermissionPublic(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPatch {
			t.Fatalf("expected PATCH, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/permissions/f1/public" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("type") != "docx" {
			t.Fatalf("unexpected type: %s", r.URL.Query().Get("type"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["link_share_entity"] != "tenant_readable" {
			t.Fatalf("unexpected link_share_entity: %+v", payload)
		}
		if payload["external_access"] != true {
			t.Fatalf("unexpected external_access: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"permission_public": map[string]any{
					"link_share_entity": "tenant_readable",
					"external_access":   true,
					"invite_external":   false,
					"share_entity":      "tenant_editable",
					"security_entity":   "tenant_editable",
					"comment_entity":    "tenant_editable",
					"lock_switch":       false,
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	allow := true
	permission, err := client.UpdateDrivePermissionPublic(context.Background(), "token", "f1", "docx", UpdateDrivePermissionPublicRequest{
		LinkShareEntity: "tenant_readable",
		ExternalAccess:  &allow,
	})
	if err != nil {
		t.Fatalf("UpdateDrivePermissionPublic error: %v", err)
	}
	if !permission.ExternalAccess || permission.LinkShareEntity != "tenant_readable" {
		t.Fatalf("unexpected permission: %+v", permission)
	}
}

func TestCreateDocxDocument(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/docx/v1/documents" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["title"] != "Specs" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		if payload["folder_token"] != "fld" {
			t.Fatalf("unexpected folder_token: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"document": map[string]any{
					"document_id": "doc1",
					"title":       "Specs",
					"url":         "https://example.com/doc",
					"revision_id": 123,
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	doc, err := client.CreateDocxDocument(context.Background(), "token", CreateDocxDocumentRequest{
		Title:       "Specs",
		FolderToken: "fld",
	})
	if err != nil {
		t.Fatalf("CreateDocxDocument error: %v", err)
	}
	if doc.DocumentID != "doc1" || doc.Title != "Specs" || string(doc.RevisionID) != "123" {
		t.Fatalf("unexpected doc: %+v", doc)
	}
}

func TestGetDocxDocument(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/docx/v1/documents/doc1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"document": map[string]any{
					"document_id": "doc1",
					"title":       "Specs",
					"url":         "https://example.com/doc",
					"revision_id": "rev1",
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	doc, err := client.GetDocxDocument(context.Background(), "token", "doc1")
	if err != nil {
		t.Fatalf("GetDocxDocument error: %v", err)
	}
	if doc.DocumentID != "doc1" || string(doc.RevisionID) != "rev1" {
		t.Fatalf("unexpected doc: %+v", doc)
	}
}

func TestCreateExportTask(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/export_tasks" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["token"] != "doc1" || payload["type"] != "docx" || payload["file_extension"] != "pdf" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"ticket": "ticket1",
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	ticket, err := client.CreateExportTask(context.Background(), "token", CreateExportTaskRequest{
		Token:         "doc1",
		Type:          "docx",
		FileExtension: "pdf",
	})
	if err != nil {
		t.Fatalf("CreateExportTask error: %v", err)
	}
	if ticket != "ticket1" {
		t.Fatalf("unexpected ticket: %s", ticket)
	}
}

func TestGetExportTask(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/export_tasks/ticket1" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"result": map[string]any{
					"file_extension": "pdf",
					"type":           "docx",
					"file_name":      "Doc.pdf",
					"file_token":     "file1",
					"file_size":      10,
					"job_error_msg":  "success",
					"job_status":     0,
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	result, err := client.GetExportTask(context.Background(), "token", "ticket1")
	if err != nil {
		t.Fatalf("GetExportTask error: %v", err)
	}
	if result.FileToken != "file1" || result.JobStatus != 0 {
		t.Fatalf("unexpected result: %+v", result)
	}
}

func TestDownloadExportedFile(t *testing.T) {
	content := []byte("export bytes")
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/drive/v1/export_tasks/file/file1/download" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_, _ = w.Write(content)
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	reader, err := client.DownloadExportedFile(context.Background(), "token", "file1")
	if err != nil {
		t.Fatalf("DownloadExportedFile error: %v", err)
	}
	defer reader.Close()
	data, err := io.ReadAll(reader)
	if err != nil {
		t.Fatalf("read download: %v", err)
	}
	if !bytes.Equal(data, content) {
		t.Fatalf("unexpected download: %q", string(data))
	}
}

func TestReadSheetRange(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values/Sheet1%21A1:B2" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"valueRange": map[string]any{
					"range":           "Sheet1!A1:B2",
					"major_dimension": "ROWS",
					"values": [][]any{
						{"Name", "Amount"},
						{"Ada", 42},
					},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	valueRange, err := client.ReadSheetRange(context.Background(), "token", "spreadsheet", "Sheet1!A1:B2")
	if err != nil {
		t.Fatalf("ReadSheetRange error: %v", err)
	}
	if valueRange.Range != "Sheet1!A1:B2" || len(valueRange.Values) != 2 {
		t.Fatalf("unexpected value range: %+v", valueRange)
	}
}

func TestUpdateSheetRange(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPut {
			t.Fatalf("expected PUT, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("valueInputOption") != "RAW" {
			t.Fatalf("unexpected valueInputOption: %s", r.URL.Query().Get("valueInputOption"))
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		valueRange, ok := payload["valueRange"].(map[string]any)
		if !ok {
			t.Fatalf("missing valueRange")
		}
		if valueRange["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range: %v", valueRange["range"])
		}
		if values, ok := valueRange["values"].([]any); !ok || len(values) != 2 {
			t.Fatalf("unexpected values: %#v", valueRange["values"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"revision":         12,
				"spreadsheetToken": "spreadsheet",
				"updatedRange":     "Sheet1!A1:B2",
				"updatedRows":      2,
				"updatedColumns":   2,
				"updatedCells":     4,
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	update, err := client.UpdateSheetRange(context.Background(), "token", "spreadsheet", "Sheet1!A1:B2", [][]any{
		{"Name", "Amount"},
		{"Ada", 42},
	})
	if err != nil {
		t.Fatalf("UpdateSheetRange error: %v", err)
	}
	if update.UpdatedRange != "Sheet1!A1:B2" || update.UpdatedCells != 4 {
		t.Fatalf("unexpected update: %+v", update)
	}
}

func TestAppendSheetRange(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values_append" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.URL.Query().Get("insertDataOption") != "INSERT_ROWS" {
			t.Fatalf("unexpected query: %s", r.URL.RawQuery)
		}
		var payload map[string]any
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		valueRange, ok := payload["valueRange"].(map[string]any)
		if !ok {
			t.Fatalf("missing valueRange")
		}
		if valueRange["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected range: %v", valueRange["range"])
		}
		if values, ok := valueRange["values"].([]any); !ok || len(values) != 2 {
			t.Fatalf("unexpected values: %#v", valueRange["values"])
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"revision":         12,
				"spreadsheetToken": "spreadsheet",
				"tableRange":       "Sheet1!A1:B2",
				"updates": map[string]any{
					"updatedRange":   "Sheet1!A1:B2",
					"updatedRows":    2,
					"updatedColumns": 2,
					"updatedCells":   4,
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	appendResult, err := client.AppendSheetRange(context.Background(), "token", "spreadsheet", "Sheet1!A1:B2", [][]any{
		{"Name", "Amount"},
		{"Ada", 42},
	}, "INSERT_ROWS")
	if err != nil {
		t.Fatalf("AppendSheetRange error: %v", err)
	}
	if appendResult.TableRange != "Sheet1!A1:B2" || appendResult.Updates.UpdatedCells != 4 {
		t.Fatalf("unexpected append result: %+v", appendResult)
	}
}

func TestClearSheetRange(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("expected POST, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/values_clear" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		var payload map[string]string
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			t.Fatalf("decode payload: %v", err)
		}
		if payload["range"] != "Sheet1!A1:B2" {
			t.Fatalf("unexpected payload: %+v", payload)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"clearedRange": "Sheet1!A1:B2",
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	clearedRange, err := client.ClearSheetRange(context.Background(), "token", "spreadsheet", "Sheet1!A1:B2")
	if err != nil {
		t.Fatalf("ClearSheetRange error: %v", err)
	}
	if clearedRange != "Sheet1!A1:B2" {
		t.Fatalf("unexpected cleared range: %s", clearedRange)
	}
}

func TestGetSpreadsheetMetadata(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/sheets/v2/spreadsheets/spreadsheet/metainfo" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"properties": map[string]any{
					"title": "Budget Q1",
				},
				"sheets": []map[string]any{
					{"sheetId": "s1", "title": "Summary", "index": 0},
					{"sheetId": "s2", "title": "Details", "index": 1},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	metadata, err := client.GetSpreadsheetMetadata(context.Background(), "token", "spreadsheet")
	if err != nil {
		t.Fatalf("GetSpreadsheetMetadata error: %v", err)
	}
	if metadata.Properties.Title != "Budget Q1" || len(metadata.Sheets) != 2 {
		t.Fatalf("unexpected metadata: %+v", metadata)
	}
}
