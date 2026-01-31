package larksdk

import (
	"encoding/json"
	"testing"
)

func TestRevisionID_UnmarshalJSON_String(t *testing.T) {
	var r RevisionID
	if err := json.Unmarshal([]byte(`"abc"`), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(r) != "abc" {
		t.Fatalf("expected %q, got %q", "abc", r)
	}
}

func TestRevisionID_UnmarshalJSON_Number(t *testing.T) {
	var r RevisionID
	if err := json.Unmarshal([]byte(`123`), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(r) != "123" {
		t.Fatalf("expected %q, got %q", "123", r)
	}
}

func TestRevisionID_UnmarshalJSON_Null(t *testing.T) {
	var r RevisionID = "existing"
	if err := json.Unmarshal([]byte(`null`), &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if string(r) != "" {
		t.Fatalf("expected empty, got %q", r)
	}
}

func TestMailFolderType_UnmarshalJSON_String(t *testing.T) {
	var ft MailFolderType
	if err := json.Unmarshal([]byte(`"INBOX"`), &ft); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ft.String() != "INBOX" {
		t.Fatalf("expected %q, got %q", "INBOX", ft)
	}
}

func TestMailFolderType_UnmarshalJSON_Number(t *testing.T) {
	var ft MailFolderType
	if err := json.Unmarshal([]byte(`2`), &ft); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ft.String() != "2" {
		t.Fatalf("expected %q, got %q", "2", ft)
	}
}

func TestMailFolderType_UnmarshalJSON_Null(t *testing.T) {
	var ft MailFolderType = "existing"
	if err := json.Unmarshal([]byte(`null`), &ft); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if ft.String() != "" {
		t.Fatalf("expected empty, got %q", ft)
	}
}

func TestDocxDocument_Unmarshal_RevisionID_Number_Regression(t *testing.T) {
	payload := []byte(`{
		"code": 0,
		"msg": "ok",
		"data": {
			"document": {
				"document_id": "doccnxxxx",
				"title": "Weekly Update",
				"url": "https://example.invalid",
				"revision_id": 42
			}
		}
	}`)

	var resp createDocxDocumentResponse
	if err := json.Unmarshal(payload, &resp); err != nil {
		t.Fatalf("unmarshal response: %v", err)
	}
	if resp.Data == nil || resp.Data.Document == nil {
		t.Fatalf("expected document")
	}
	if got := string(resp.Data.Document.RevisionID); got != "42" {
		t.Fatalf("expected revision_id %q, got %q", "42", got)
	}
}
