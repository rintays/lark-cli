package main

import (
	"bytes"
	"encoding/json"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/output"
)

func TestAuthUserStatusJSONIncludesTokenPresenceAndExpiry(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		ConfigPath: "/tmp/config.json",
		Config:     &config.Config{},
		Printer:    output.Printer{Writer: &buf, JSON: true},
	}
	withUserAccount(state.Config, defaultUserAccountName, "access", "refresh", 1700000000, "offline_access contact:contact.base:readonly")

	cmd := newAuthUserCmd(state)
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth user status error: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v; output=%q", err, buf.String())
	}

	assertJSONBool(t, got, "user_access_token_present", true)
	assertJSONBool(t, got, "refresh_token_present", true)
	assertJSONInt64(t, got, "user_access_token_expires_at", 1700000000)
	assertJSONString(t, got, "user_access_token_scope", "offline_access contact:contact.base:readonly")

	wantRFC3339 := time.Unix(1700000000, 0).UTC().Format(time.RFC3339)
	assertJSONString(t, got, "user_access_token_expires_at_rfc3339", wantRFC3339)

	if _, ok := got["remediation"]; ok {
		t.Fatalf("expected remediation omitted when refresh_token present")
	}
}

func TestAuthUserStatusJSONIncludesRemediationWhenRefreshTokenMissing(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		ConfigPath: "/tmp/config.json",
		Config:     &config.Config{},
		Printer:    output.Printer{Writer: &buf, JSON: true},
	}
	withUserAccount(state.Config, defaultUserAccountName, "access", "", 0, "")

	cmd := newAuthUserCmd(state)
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth user status error: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v; output=%q", err, buf.String())
	}

	assertJSONBool(t, got, "user_access_token_present", true)
	assertJSONBool(t, got, "refresh_token_present", false)
	assertJSONInt64(t, got, "user_access_token_expires_at", 0)
	assertJSONString(t, got, "remediation", userOAuthReloginCommand)
}

func TestAuthUserStatusUsesRefreshTokenPayloadMetadata(t *testing.T) {
	var buf bytes.Buffer
	state := &appState{
		ConfigPath: "/tmp/config.json",
		Config:     &config.Config{},
		Printer:    output.Printer{Writer: &buf, JSON: true},
	}
	withUserAccount(state.Config, defaultUserAccountName, "access", "", 0, "")
	if acct, ok := state.Config.UserAccounts[defaultUserAccountName]; ok {
		acct.UserRefreshTokenPayload = &config.UserRefreshTokenPayload{
			RefreshToken: "payload-refresh",
			Services:     []string{"drive", "docs"},
			Scopes:       "offline_access drive:drive",
			CreatedAt:    1700000200,
		}
	}

	cmd := newAuthUserCmd(state)
	cmd.SetArgs([]string{"status"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("auth user status error: %v", err)
	}

	var got map[string]json.RawMessage
	if err := json.Unmarshal(buf.Bytes(), &got); err != nil {
		t.Fatalf("unmarshal output: %v; output=%q", err, buf.String())
	}

	assertJSONBool(t, got, "refresh_token_present", true)
	assertJSONStrings(t, got, "refresh_token_services", []string{"drive", "docs"})
	assertJSONString(t, got, "refresh_token_scopes", "offline_access drive:drive")
	assertJSONInt64(t, got, "refresh_token_created_at", 1700000200)

	wantRFC3339 := time.Unix(1700000200, 0).UTC().Format(time.RFC3339)
	assertJSONString(t, got, "refresh_token_created_at_rfc3339", wantRFC3339)

	if _, ok := got["remediation"]; ok {
		t.Fatalf("expected remediation omitted when refresh_token present")
	}
}

func assertJSONBool(t *testing.T, got map[string]json.RawMessage, key string, want bool) {
	t.Helper()
	raw, ok := got[key]
	if !ok {
		t.Fatalf("expected key %q", key)
	}
	var value bool
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if value != want {
		t.Fatalf("%s: expected %t, got %t", key, want, value)
	}
}

func assertJSONInt64(t *testing.T, got map[string]json.RawMessage, key string, want int64) {
	t.Helper()
	raw, ok := got[key]
	if !ok {
		t.Fatalf("expected key %q", key)
	}
	var value int64
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if value != want {
		t.Fatalf("%s: expected %d, got %d", key, want, value)
	}
}

func assertJSONString(t *testing.T, got map[string]json.RawMessage, key string, want string) {
	t.Helper()
	raw, ok := got[key]
	if !ok {
		t.Fatalf("expected key %q", key)
	}
	var value string
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if value != want {
		t.Fatalf("%s: expected %q, got %q", key, want, value)
	}
}

func assertJSONStrings(t *testing.T, got map[string]json.RawMessage, key string, want []string) {
	t.Helper()
	raw, ok := got[key]
	if !ok {
		t.Fatalf("expected key %q", key)
	}
	var value []string
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatalf("unmarshal %s: %v", key, err)
	}
	if len(value) != len(want) {
		t.Fatalf("%s: expected %v, got %v", key, want, value)
	}
	for i := range value {
		if value[i] != want[i] {
			t.Fatalf("%s: expected %v, got %v", key, want, value)
		}
	}
}
