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
	// Wiki member v2 endpoints support tenant tokens.
	if _, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser); err != nil {
		t.Skipf("tenant token missing/unavailable (run `lark auth tenant` to cache a tenant token): %v", err)
	}

	initialRole, initialFound, err := wikiMemberRole(t, spaceID, email)
	if err != nil {
		t.Fatalf("wiki member list failed: %v", err)
	}
	if initialFound && strings.EqualFold(initialRole, "admin") {
		t.Skipf("precondition failed: %s is already admin in space %s; choose a non-admin member to avoid downgrading privileges during this test", email, spaceID)
	}

	// Best-effort cleanup: avoid leaving the test member as admin.
	t.Cleanup(func() {
		if !initialFound {
			_, _ = executeLark(t, []string{"--json", "--force", "--token-type", "tenant", "wiki", "member", "delete", "email", email, "--space-id", spaceID})
			return
		}

		// Some environments require delete+add to change roles.
		_, _ = executeLark(t, []string{"--json", "--force", "--token-type", "tenant", "wiki", "member", "delete", "email", email, "--space-id", spaceID})
		_, _ = executeLark(t, []string{"--json", "--token-type", "tenant", "wiki", "member", "add", "email", email, "--space-id", spaceID, "--role", initialRole})
	})

	// (1) Ensure the member starts at role=member.
	_, _ = executeLark(t, []string{"--json", "--force", "--token-type", "tenant", "wiki", "member", "delete", "email", email, "--space-id", spaceID})

	out, err := executeLark(t, []string{
		"--json",
		"--token-type",
		"tenant",
		"wiki",
		"member",
		"add",
		"email",
		email,
		"--space-id",
		spaceID,
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
		t.Fatalf("expected role=member before update attempt, got %q", role)
	}

	// (2) Re-add the same member with a different role.
	out, createErr := executeLark(t, []string{
		"--json",
		"--token-type",
		"tenant",
		"wiki",
		"member",
		"add",
		"email",
		email,
		"--space-id",
		spaceID,
		"--role",
		"admin",
	})

	// (3) Assert via list that the role changed (upsert) OR document that it does not.
	updatedRole, ok, listErr := waitForWikiMemberRole(t, spaceID, email, "admin")
	if listErr != nil {
		t.Fatalf("wiki member list after role=admin attempt failed: %v (addErr=%v; addOut=%q)", listErr, createErr, out)
	}
	if !ok {
		t.Fatalf("member %q not found after role=admin attempt (addErr=%v; addOut=%q)", email, createErr, out)
	}

	if strings.EqualFold(updatedRole, "admin") {
		t.Logf("SpaceMember.Create appears to be an upsert in this environment (member→admin)")
		return
	}

	if createErr == nil {
		t.Fatalf("expected wiki member add(role=admin) to either update role to admin or return an error; got role=%q with no error (out=%q)", updatedRole, out)
	}
	if !strings.EqualFold(updatedRole, "member") {
		t.Fatalf("expected role to remain member when SpaceMember.Create is not an upsert, got %q (addErr=%v; out=%q)", updatedRole, createErr, out)
	}
	if !strings.Contains(strings.ToLower(createErr.Error()+" "+out), "exist") {
		t.Logf("note: create returned unexpected error for already-exists case: %v; out=%q", createErr, out)
	}
	// Official docs mention error 400131008 (already exist) for repeated add operations.
	t.Logf("SpaceMember.Create is not an upsert in this environment: re-adding an existing member did not change role")
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
	for i := 0; i < 10; i++ {
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
