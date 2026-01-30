package larkapi

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"lark/internal/testutil"
)

func TestListMailFolders(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			t.Fatalf("expected GET, got %s", r.Method)
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Fatalf("missing auth header")
		}
		if r.URL.Path != "/open-apis/mail/v1/user_mailboxes/mbx_1/folders" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"msg":  "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{
						"folder_id":   "fld_1",
						"name":        "Inbox",
						"folder_type": "INBOX",
					},
				},
			},
		})
	})
	httpClient, baseURL := testutil.NewTestClient(handler)

	client := &Client{BaseURL: baseURL, HTTPClient: httpClient}
	folders, err := client.ListMailFolders(context.Background(), "token", "mbx_1")
	if err != nil {
		t.Fatalf("ListMailFolders error: %v", err)
	}
	if len(folders) != 1 {
		t.Fatalf("unexpected folders: %+v", folders)
	}
	if folders[0].FolderID != "fld_1" || folders[0].Name != "Inbox" || folders[0].FolderType != "INBOX" {
		t.Fatalf("unexpected folder: %+v", folders[0])
	}
}
