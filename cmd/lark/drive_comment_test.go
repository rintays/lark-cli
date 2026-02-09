package main

import (
	"strings"
	"testing"
)

func TestResolveDriveCommentFileRef(t *testing.T) {
	token, fileType, err := resolveDriveCommentFileRef("doccn123", "docx", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "doccn123" {
		t.Fatalf("unexpected token: %s", token)
	}
	if fileType != "docx" {
		t.Fatalf("unexpected file type: %s", fileType)
	}

	token, fileType, err = resolveDriveCommentFileRef("https://example.com/docx/doccnABC", "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token != "doccnABC" {
		t.Fatalf("unexpected token: %s", token)
	}
	if fileType != "docx" {
		t.Fatalf("unexpected inferred file type: %s", fileType)
	}

	_, _, err = resolveDriveCommentFileRef("https://example.com/docx/doccnABC", "unsupported", "")
	if err == nil || !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("expected unsupported type error, got: %v", err)
	}
}

func TestParseDriveCommentReplyContent(t *testing.T) {
	content, err := parseDriveCommentReplyContent("hello", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content == nil || len(content.Elements) != 1 {
		t.Fatalf("unexpected content: %#v", content)
	}
	if content.Elements[0] == nil || derefString(content.Elements[0].Type) != "text_run" {
		t.Fatalf("unexpected element: %#v", content.Elements[0])
	}
	if content.Elements[0].TextRun == nil || derefString(content.Elements[0].TextRun.Text) != "hello" {
		t.Fatalf("unexpected text_run: %#v", content.Elements[0].TextRun)
	}

	_, err = parseDriveCommentReplyContent("", "")
	if err == nil {
		t.Fatalf("expected error for empty content")
	}

	_, err = parseDriveCommentReplyContent("hello", "{}")
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("expected mutually exclusive error, got: %v", err)
	}

	_, err = parseDriveCommentReplyContent("", "not-json")
	if err == nil || !strings.Contains(err.Error(), "invalid") {
		t.Fatalf("expected invalid json error, got: %v", err)
	}

	_, err = parseDriveCommentReplyContent("", "{}")
	if err == nil || !strings.Contains(err.Error(), "at least one element") {
		t.Fatalf("expected elements validation error, got: %v", err)
	}

	content, err = parseDriveCommentReplyContent("", `{"elements":[{"type":"text_run","text_run":{"text":"ok"}}]}`)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if content == nil || len(content.Elements) != 1 {
		t.Fatalf("unexpected content: %#v", content)
	}
}

func TestBuildDriveFileCommentCreate(t *testing.T) {
	content, err := parseDriveCommentReplyContent("hello", "")
	if err != nil {
		t.Fatalf("content error: %v", err)
	}

	comment, err := buildDriveFileCommentCreate(nil, content, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment == nil || comment.IsWhole == nil || !*comment.IsWhole {
		t.Fatalf("expected whole comment, got: %#v", comment)
	}
	if comment.Quote != nil {
		t.Fatalf("expected no quote, got: %#v", comment.Quote)
	}
	if comment.ReplyList == nil || len(comment.ReplyList.Replies) != 1 {
		t.Fatalf("expected one reply, got: %#v", comment.ReplyList)
	}

	comment, err = buildDriveFileCommentCreate(nil, content, "{...}")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if comment.IsWhole == nil || *comment.IsWhole {
		t.Fatalf("expected partial comment, got: %#v", comment.IsWhole)
	}
	if comment.Quote == nil || *comment.Quote != "{...}" {
		t.Fatalf("expected quote, got: %#v", comment.Quote)
	}

	parent := "c123"
	reply, err := buildDriveFileCommentCreate(&parent, content, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if reply.CommentId == nil || *reply.CommentId != "c123" {
		t.Fatalf("expected comment_id set, got: %#v", reply.CommentId)
	}
	if reply.IsWhole != nil || reply.Quote != nil {
		t.Fatalf("expected reply to not set quote/is_whole, got: %#v", reply)
	}
}
