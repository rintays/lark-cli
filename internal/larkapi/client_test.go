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
	if doc.DocumentID != "doc1" || doc.Title != "Specs" {
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
	if doc.DocumentID != "doc1" || doc.RevisionID != "rev1" {
		t.Fatalf("unexpected doc: %+v", doc)
	}
}
