package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"lark/internal/output"
)

type authExplainPayload struct {
	Command                           string   `json:"command"`
	Services                          []string `json:"services"`
	TokenTypes                        []string `json:"token_types"`
	RequiresOffline                   bool     `json:"requires_offline"`
	RequiredUserScopes                []string `json:"required_user_scopes"`
	ServicesMissingRequiredUserScopes []string `json:"services_missing_required_user_scopes"`
	SuggestedUserLoginScopes          []string `json:"suggested_user_login_scopes"`
	SuggestedUserLoginCommand         string   `json:"suggested_user_login_command"`
}

func TestAuthExplainDriveSearchJSON(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{Printer: output.Printer{Writer: &buf, JSON: true}}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"explain", "drive", "search"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth explain error: %v", err)
	}

	var payload authExplainPayload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, buf.String())
	}

	if payload.Command != "drive search" {
		t.Fatalf("command: expected %q, got %q", "drive search", payload.Command)
	}
	if strings.Join(payload.Services, ",") != "drive" {
		t.Fatalf("services: expected [drive], got %v", payload.Services)
	}
	if strings.Join(payload.TokenTypes, ",") != "tenant,user" {
		t.Fatalf("token_types: expected [tenant user], got %v", payload.TokenTypes)
	}
	if !payload.RequiresOffline {
		t.Fatalf("requires_offline: expected true")
	}
	if strings.Join(payload.RequiredUserScopes, ",") != "drive:drive" {
		t.Fatalf("required_user_scopes: expected [drive:drive], got %v", payload.RequiredUserScopes)
	}
	if payload.SuggestedUserLoginCommand == "" {
		t.Fatalf("expected suggested_user_login_command")
	}
	if !strings.Contains(payload.SuggestedUserLoginCommand, "lark auth user login") {
		t.Fatalf("unexpected suggested_user_login_command: %q", payload.SuggestedUserLoginCommand)
	}
	if strings.Join(payload.SuggestedUserLoginScopes, " ") == "" {
		t.Fatalf("expected suggested_user_login_scopes")
	}
	if payload.SuggestedUserLoginScopes[0] != defaultUserOAuthScope {
		t.Fatalf("expected suggested scopes to include %q first, got %v", defaultUserOAuthScope, payload.SuggestedUserLoginScopes)
	}
}

func TestAuthExplainChatsListNoUserToken(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{Printer: output.Printer{Writer: &buf, JSON: true}}

	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"explain", "chats", "list"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth explain error: %v", err)
	}

	var payload authExplainPayload
	if err := json.Unmarshal(buf.Bytes(), &payload); err != nil {
		t.Fatalf("unmarshal output: %v\noutput=%s", err, buf.String())
	}

	if payload.Command != "chats list" {
		t.Fatalf("command: expected %q, got %q", "chats list", payload.Command)
	}
	if strings.Join(payload.Services, ",") != "im" {
		t.Fatalf("services: expected [im], got %v", payload.Services)
	}
	if strings.Join(payload.TokenTypes, ",") != "tenant" {
		t.Fatalf("token_types: expected [tenant], got %v", payload.TokenTypes)
	}
	if payload.SuggestedUserLoginCommand != "" {
		t.Fatalf("expected no suggested_user_login_command, got %q", payload.SuggestedUserLoginCommand)
	}
}

func TestAuthExplainUnknownCommandErrors(t *testing.T) {
	state := &appState{Printer: output.Printer{Writer: &bytes.Buffer{}}}
	cmd := newAuthCmd(state)
	cmd.SetArgs([]string{"explain", "nope"})
	if err := cmd.Execute(); err == nil {
		t.Fatalf("expected error")
	} else if !strings.Contains(err.Error(), "no auth registry mapping") {
		t.Fatalf("unexpected error: %v", err)
	}
}
