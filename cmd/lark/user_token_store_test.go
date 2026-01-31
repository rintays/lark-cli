package main

import (
	"testing"

	"lark/internal/config"
)

func TestUserTokenBucketID_StableAndDependsOnAppBaseURLAndProfile(t *testing.T) {
	baseState := func() *appState {
		return &appState{
			ConfigPath: "/tmp/lark/profiles/p1/config.json",
			Config: &config.Config{
				AppID:   "cli-app-1",
				BaseURL: "https://open.feishu.cn/open-apis/",
			},
		}
	}

	a := baseState()
	b := baseState()
	if userTokenBucketID(a) == "" {
		t.Fatalf("expected non-empty bucket id")
	}
	if userTokenBucketID(a) != userTokenBucketID(b) {
		t.Fatalf("expected stable bucket id")
	}

	// Base URL normalization should not affect the bucket.
	c := baseState()
	c.Config.BaseURL = "https://open.feishu.cn"
	if userTokenBucketID(a) != userTokenBucketID(c) {
		t.Fatalf("expected bucket id to treat normalized base_url as equivalent")
	}

	// app_id differences must isolate tokens.
	d := baseState()
	d.Config.AppID = "cli-app-2"
	if userTokenBucketID(a) == userTokenBucketID(d) {
		t.Fatalf("expected different bucket id for different app_id")
	}

	// base_url differences must isolate tokens.
	e := baseState()
	e.Config.BaseURL = "https://open.larksuite.com"
	if userTokenBucketID(a) == userTokenBucketID(e) {
		t.Fatalf("expected different bucket id for different base_url")
	}

	// profile (config path) differences must isolate tokens.
	f := baseState()
	f.ConfigPath = "/tmp/lark/profiles/p2/config.json"
	if userTokenBucketID(a) == userTokenBucketID(f) {
		t.Fatalf("expected different bucket id for different profile/config path")
	}
}

func TestKeyringUsername_IncludesBucketAndAccount(t *testing.T) {
	state := &appState{
		ConfigPath: "/tmp/lark/profiles/p1/config.json",
		Config: &config.Config{
			AppID:   "cli-app-1",
			BaseURL: "https://open.feishu.cn",
		},
	}
	bucket := userTokenBucketID(state)
	got := keyringUsername(state, "acct")
	want := bucket + ":acct"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestLegacyKeyringUsername_DoesNotIsolateAcrossApps(t *testing.T) {
	path := "/tmp/lark/profiles/p1/config.json"

	a := &appState{ConfigPath: path, Config: &config.Config{AppID: "app-a", BaseURL: "https://open.feishu.cn"}}
	b := &appState{ConfigPath: path, Config: &config.Config{AppID: "app-b", BaseURL: "https://open.feishu.cn"}}

	legacy := legacyKeyringUsername(path, "acct")
	if legacy == "" {
		t.Fatalf("expected non-empty legacy username")
	}
	// Legacy scheme depends only on profile/config path, so it does not change as
	// app_id changes.
	if legacy != legacyKeyringUsername(path, "acct") {
		t.Fatalf("expected legacy username to be stable")
	}

	if userTokenBucketID(a) == userTokenBucketID(b) {
		t.Fatalf("expected different bucket id for different app_id")
	}
	if keyringUsername(a, "acct") == keyringUsername(b, "acct") {
		t.Fatalf("expected different keyring username for different app_id")
	}
}
