package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"strconv"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/output"
	"lark/internal/testutil"
)

func TestWikiMemberRoleUpdateIntegration(t *testing.T) {
	testutil.RequireIntegration(t)

	spaceID := testutil.RequireEnv(t, "LARK_TEST_WIKI_SPACE_ID")
	email := testutil.RequireEnv(t, "LARK_TEST_USER_EMAIL")

	state := integrationTestState(t)
	if _, err := tokenFor(context.Background(), state, tokenTypesUser); err != nil {
		t.Skipf("user token missing/unavailable (run `lark auth user login` to cache a refresh token): %v", err)
	}

	current, ok, err := wikiMemberRole(t, spaceID, email)
	if err != nil {
		t.Fatalf("wiki member list failed: %v", err)
	}
	if ok && strings.EqualFold(current, "admin") {
		t.Skipf("precondition failed: %s is already admin in space %s; choose a member with role=member (or remove/re-add) to verify role upsert", email, spaceID)
	}

	out, err := executeLark(t, []string{
		"--json",
		"--token-type",
		"tenant",
		"wiki",
		"member",
		"add",
		"--space-id",
		spaceID,
		"--member-type",
		"email",
		"--member-id",
		email,
		"--role",
		"member",
	})
	if err != nil {
		role, ok, listErr := waitForWikiMemberRole(t, spaceID, email, "member")
		if listErr != nil {
			t.Fatalf("wiki member add (role=member) failed: %v; addOut=%q; listErr=%v", err, out, listErr)
		}
		if !ok || !strings.EqualFold(role, "member") {
			t.Fatalf("wiki member add (role=member) failed and member role is not member; err=%v; addOut=%q; role=%q", err, out, role)
		}
	}

	role, ok, err := waitForWikiMemberRole(t, spaceID, email, "member")
	if err != nil {
		t.Fatalf("wiki member list after role=member failed: %v", err)
	}
	if !ok {
		t.Fatalf("member %q not found after adding role=member", email)
	}
	if !strings.EqualFold(role, "member") {
		t.Fatalf("expected role=member before update, got %q", role)
	}

	out, err = executeLark(t, []string{
		"--json",
		"--token-type",
		"tenant",
		"wiki",
		"member",
		"add",
		"--space-id",
		spaceID,
		"--member-type",
		"email",
		"--member-id",
		email,
		"--role",
		"admin",
	})
	updatedRole, ok, listErr := waitForWikiMemberRole(t, spaceID, email, "admin")
	if listErr != nil {
		t.Fatalf("wiki member list after role=admin failed: %v (createErr=%v; addOut=%q)", listErr, err, out)
	}
	if !ok {
		t.Fatalf("member %q not found after adding role=admin (createErr=%v; addOut=%q)", email, err, out)
	}
	if !strings.EqualFold(updatedRole, "admin") {
		t.Fatalf("expected member role to update to admin, got %q (createErr=%v; addOut=%q)", updatedRole, err, out)
	}
}

func integrationTestState(t *testing.T) *appState {
	t.Helper()

	cfgPath, err := config.DefaultPath()
	if err != nil {
		t.Fatalf("default config path: %v", err)
	}
	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatalf("load config: %v", err)
	}

	state := &appState{ConfigPath: cfgPath, Config: cfg}
	if err := applyBaseURLOverrides(state, cfg); err != nil {
		t.Fatalf("apply base URL overrides: %v", err)
	}
	state.Printer = output.Printer{Writer: io.Discard}
	return state
}

func executeLark(t *testing.T, args []string) (string, error) {
	t.Helper()

	var buf bytes.Buffer
	cmd := newRootCmd()
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)
	cmd.SetArgs(args)
	return buf.String(), cmd.Execute()
}

type wikiMemberListPayload struct {
	Members []larksdk.WikiSpaceMember `json:"members"`
}

func wikiMembers(t *testing.T, spaceID string) ([]larksdk.WikiSpaceMember, string, error) {
	t.Helper()

	out, err := executeLark(t, []string{
		"--json",
		"--token-type",
		"tenant",
		"wiki",
		"member",
		"list",
		"--space-id",
		spaceID,
		"--limit",
		strconv.Itoa(1000),
	})
	if err != nil {
		return nil, out, err
	}

	var payload wikiMemberListPayload
	if err := json.Unmarshal([]byte(out), &payload); err != nil {
		return nil, out, err
	}
	return payload.Members, out, nil
}

func wikiMemberRole(t *testing.T, spaceID, email string) (role string, found bool, err error) {
	t.Helper()

	members, _, err := wikiMembers(t, spaceID)
	if err != nil {
		return "", false, err
	}
	for _, m := range members {
		if strings.EqualFold(m.MemberType, "email") && strings.EqualFold(m.MemberID, email) {
			return m.MemberRole, true, nil
		}
	}
	return "", false, nil
}

func waitForWikiMemberRole(t *testing.T, spaceID, email, want string) (role string, found bool, err error) {
	t.Helper()

	var lastRole string
	for i := 0; i < 5; i++ {
		role, found, err = wikiMemberRole(t, spaceID, email)
		if err != nil {
			return "", false, err
		}
		lastRole = role
		if found && strings.EqualFold(role, want) {
			return role, true, nil
		}
		time.Sleep(2 * time.Second)
	}
	return lastRole, found, nil
}
