package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

const integrationFixturePrefix = "clawdbot-it-"

type integrationFixtures struct {
	ConfigPath string

	// Convenience for SDK-backed integration tests.
	SDK   *larksdk.Client
	Token string

	// Common external fixtures.
	DriveFolderToken string
	ChatID           string
	UserEmail        string

	SpreadsheetToken string
	SheetTitle       string

	MailTo string
}

func (fx integrationFixtures) EnsureChatID(t *testing.T) string {
	t.Helper()
	if fx.ChatID == "" {
		t.Skip("missing LARK_TEST_CHAT_ID")
	}
	return fx.ChatID
}

func getIntegrationFixtures(t *testing.T) integrationFixtures {
	t.Helper()

	cfg := config.Default()
	cfg.AppID = testutil.RequireEnv(t, "LARK_APP_ID")
	cfg.AppSecret = testutil.RequireEnv(t, "LARK_APP_SECRET")
	if baseURL := os.Getenv("LARK_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	// Write an ephemeral config file so integration tests don't depend on any local config state.
	cfgDir := t.TempDir()
	cfgPath := filepath.Join(cfgDir, "config.json")
	f, err := os.Create(cfgPath)
	if err != nil {
		t.Fatalf("create config file: %v", err)
	}
	defer f.Close()
	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(cfg); err != nil {
		t.Fatalf("write config file: %v", err)
	}

	sdk, err := larksdk.New(cfg)
	if err != nil {
		t.Fatalf("sdk init: %v", err)
	}
	token, _, err := sdk.TenantAccessToken(t.Context())
	if err != nil {
		t.Fatalf("tenant token: %v", err)
	}

	spreadsheetToken := os.Getenv("LARK_TEST_SHEET_ID")
	sheetTitle := os.Getenv("LARK_TEST_SHEET_TITLE")
	if sheetTitle == "" {
		if r := os.Getenv("LARK_TEST_SHEET_RANGE"); r != "" {
			if before, _, ok := strings.Cut(r, "!"); ok {
				sheetTitle = before
			}
		}
	}

	return integrationFixtures{
		ConfigPath: cfgPath,
		SDK:        sdk,
		Token:      token,

		DriveFolderToken: os.Getenv("LARK_TEST_FOLDER_TOKEN"),
		ChatID:           os.Getenv("LARK_TEST_CHAT_ID"),
		UserEmail:        os.Getenv("LARK_TEST_USER_EMAIL"),

		SpreadsheetToken: spreadsheetToken,
		SheetTitle:       sheetTitle,

		MailTo: os.Getenv("LARK_TEST_MAIL_TO"),
	}
}
