package larkapi

import (
	"context"
	"encoding/json"
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
