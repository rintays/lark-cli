package main

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"
	"time"

	"lark/internal/larksdk"
	"lark/internal/testutil"
)

func TestDocsCatMarkdownIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	ctx := t.Context()

	documentID := os.Getenv("LARK_TEST_DOC_ID")
	createdTemp := false

	title := integrationFixturePrefix + "docs-cat-md-" + time.Now().Format("20060102-150405")
	needle := integrationFixturePrefix + "hello-md-" + time.Now().Format("150405.000")

	if documentID == "" {
		doc, err := fx.SDK.CreateDocxDocument(ctx, fx.Token, larksdk.CreateDocxDocumentRequest{
			Title:       title,
			FolderToken: fx.DriveFolderToken,
		})
		if err != nil {
			t.Fatalf("create docx document: %v", err)
		}
		if doc.DocumentID == "" {
			t.Fatalf("create docx document returned empty document_id: %#v", doc)
		}
		documentID = doc.DocumentID
		createdTemp = true
		t.Logf("created temp doc for docs cat test: document_id=%s title=%q", documentID, title)
	} else {
		t.Logf("using existing doc from LARK_TEST_DOC_ID=%s", documentID)
	}

	if createdTemp {
		t.Cleanup(func() {
			_, err := fx.SDK.DeleteDriveFile(context.Background(), fx.Token, documentID, "docx")
			if err != nil {
				// Best-effort cleanup. Some tenants may not allow delete.
				t.Logf("cleanup: delete docx %s failed (best-effort): %v", documentID, err)
			}
		})
	}

	if _, err := fx.SDK.AppendDocxTextBlock(ctx, fx.Token, documentID, needle); err != nil {
		t.Fatalf("append docx text block: %v", err)
	}

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs([]string{"--config", fx.ConfigPath, "--json", "docs", "cat", documentID, "--format", "md"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("docs cat failed: %v; out=%q", err, buf.String())
	}

	var payload map[string]any
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("invalid json output: %v; out=%q", err, buf.String())
	}

	content, _ := payload["content"].(string)
	if strings.TrimSpace(content) == "" {
		t.Fatalf("expected non-empty content; payload=%v", payload)
	}
	if !strings.Contains(content, needle) {
		t.Fatalf("expected markdown to contain inserted text %q; got=%q", needle, content)
	}
}
