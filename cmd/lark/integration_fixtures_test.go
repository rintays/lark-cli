package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

const integrationFixturePrefix = "lark-cli-it-"

const (
	integrationBaseAppName     = "lark-cli-it-base"
	integrationBaseTablePrefix = "lark-cli-it-"
)

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
	SheetID          string
	SheetTitle       string

	// Base/Bitable fixtures.
	BaseAppToken string

	MailTo string
}

func (fx integrationFixtures) EnsureChatID(t *testing.T) string {
	t.Helper()
	if fx.ChatID != "" {
		return fx.ChatID
	}

	ctx := t.Context()
	userIDs := []string{}
	if fx.UserEmail != "" {
		users, err := fx.SDK.BatchGetUserIDs(ctx, fx.Token, larksdk.BatchGetUserIDRequest{Emails: []string{fx.UserEmail}})
		if err != nil {
			t.Fatalf("batch get user id for %s: %v", fx.UserEmail, err)
		}
		if len(users) > 0 && users[0].UserID != "" {
			userIDs = []string{users[0].UserID}
		}
	}

	chatName := integrationFixturePrefix + "chat-" + time.Now().Format("20060102-150405")
	chatID, err := fx.SDK.CreateChat(ctx, fx.Token, chatName, userIDs)
	if err != nil {
		t.Fatalf("create chat: %v", err)
	}
	// Best-effort cleanup.
	t.Cleanup(func() {
		if err := fx.SDK.DeleteChat(context.Background(), fx.Token, chatID); err != nil {
			t.Logf("cleanup: delete chat %s: %v", chatID, err)
		}
	})
	return chatID
}

func (fx *integrationFixtures) EnsureBaseAppToken(t *testing.T) string {
	t.Helper()
	if fx.BaseAppToken != "" {
		return fx.BaseAppToken
	}

	ctx := t.Context()

	res, err := fx.SDK.SearchDriveFiles(ctx, fx.Token, larksdk.SearchDriveFilesRequest{
		Query:     integrationBaseAppName,
		FileTypes: []string{"bitable"},
		PageSize:  50,
	})
	if err != nil {
		// Some tenants may not allow drive search with tenant token.
		// Fall back to creating the dedicated app.
		app, err2 := fx.SDK.CreateBitableApp(ctx, fx.Token, integrationBaseAppName, fx.DriveFolderToken)
		if err2 != nil {
			t.Fatalf("search drive for base app: %v; create bitable app failed: %v", err, err2)
		}
		fx.BaseAppToken = app.AppToken
		return fx.BaseAppToken
	}
	files := res.Files
	for res.HasMore {
		res, err = fx.SDK.SearchDriveFiles(ctx, fx.Token, larksdk.SearchDriveFilesRequest{
			Query:     integrationBaseAppName,
			FileTypes: []string{"bitable"},
			PageSize:  50,
			PageToken: res.PageToken,
		})
		if err != nil {
			t.Fatalf("search drive for base app (page): %v", err)
		}
		files = append(files, res.Files...)
		if !res.HasMore {
			break
		}
	}

	for _, f := range files {
		if f.Name != integrationBaseAppName {
			continue
		}
		if token, ok := parseBitableAppTokenFromURL(f.URL); ok {
			fx.BaseAppToken = token
			return fx.BaseAppToken
		}
	}

	app, err := fx.SDK.CreateBitableApp(ctx, fx.Token, integrationBaseAppName, fx.DriveFolderToken)
	if err != nil {
		t.Fatalf("create bitable app: %v", err)
	}
	fx.BaseAppToken = app.AppToken
	return fx.BaseAppToken
}

func (fx integrationFixtures) SweepBaseTables(t *testing.T, appToken string) {
	t.Helper()
	if appToken == "" {
		return
	}
	ctx := t.Context()
	tables, err := fx.SDK.ListBaseTablesAll(ctx, fx.Token, appToken)
	if err != nil {
		t.Fatalf("list base tables: %v", err)
	}
	for _, tbl := range tables {
		if tbl.Name == "" || !strings.HasPrefix(tbl.Name, integrationBaseTablePrefix) {
			continue
		}
		_, err := fx.SDK.DeleteBaseTable(ctx, fx.Token, appToken, tbl.TableID)
		if err != nil {
			// Best-effort: avoid failing the whole suite on cleanup.
			t.Logf("sweep: delete table %s (%s): %v", tbl.Name, tbl.TableID, err)
		}
	}
}

func (fx integrationFixtures) CreateTempBaseTable(t *testing.T, appToken string) (string, func()) {
	t.Helper()
	if tableID := os.Getenv("LARK_TEST_TABLE_ID"); tableID != "" {
		return tableID, func() {}
	}
	fx.SweepBaseTables(t, appToken)

	ctx := t.Context()
	tableName := fmt.Sprintf("%stable-%s-%d", integrationBaseTablePrefix, sanitizeForFixtureName(t.Name()), time.Now().UnixNano())
	table, err := fx.SDK.CreateBaseTable(ctx, fx.Token, appToken, tableName, "")
	if err != nil {
		t.Fatalf("create base table: %v", err)
	}
	if table.TableID == "" {
		t.Fatalf("create base table returned empty table_id")
	}
	cleanup := func() {
		_, err := fx.SDK.DeleteBaseTable(context.Background(), fx.Token, appToken, table.TableID)
		if err != nil {
			t.Fatalf("delete base table: %v", err)
		}
	}
	return table.TableID, cleanup
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

	fx := integrationFixtures{
		ConfigPath: cfgPath,
		SDK:        sdk,
		Token:      token,

		DriveFolderToken: os.Getenv("LARK_TEST_FOLDER_TOKEN"),
		ChatID:           os.Getenv("LARK_TEST_CHAT_ID"),
		UserEmail:        os.Getenv("LARK_TEST_USER_EMAIL"),

		SpreadsheetToken: os.Getenv("LARK_TEST_SHEET_ID"),
		SheetID:          os.Getenv("LARK_TEST_SHEET_SHEET_ID"),
		SheetTitle:       os.Getenv("LARK_TEST_SHEET_TITLE"),

		BaseAppToken: os.Getenv("LARK_TEST_APP_TOKEN"),
		MailTo:       os.Getenv("LARK_TEST_MAIL_TO"),
	}

	// If spreadsheet token is not provided, create one (best-effort) for tests.
	if fx.SpreadsheetToken == "" {
		title := integrationFixturePrefix + "sheet-" + time.Now().Format("20060102-150405")
		ssToken, err := sdk.CreateSpreadsheet(t.Context(), token, title, fx.DriveFolderToken)
		if err != nil {
			t.Fatalf("create spreadsheet: %v", err)
		}
		fx.SpreadsheetToken = ssToken
		t.Cleanup(func() {
			if err := sdk.DeleteDriveFile(context.Background(), token, ssToken, "sheet"); err != nil {
				t.Logf("cleanup: delete spreadsheet %s: %v", ssToken, err)
			}
		})
	}

	// Derive sheet id/title from API.
	if fx.SheetID == "" || fx.SheetTitle == "" {
		sheets, err := sdk.ListSpreadsheetSheets(t.Context(), token, fx.SpreadsheetToken)
		if err != nil {
			t.Fatalf("list spreadsheet sheets: %v", err)
		}
		if len(sheets) == 0 {
			t.Fatalf("no sheets found in spreadsheet %s", fx.SpreadsheetToken)
		}
		if fx.SheetID == "" {
			fx.SheetID = sheets[0].SheetID
		}
		if fx.SheetTitle == "" {
			fx.SheetTitle = sheets[0].Title
		}
	}
	if fx.SheetTitle == "" {
		fx.SheetTitle = "Sheet1"
	}

	return fx
}

func sanitizeForFixtureName(name string) string {
	name = strings.ToLower(name)
	b := make([]rune, 0, len(name))
	for _, r := range name {
		switch {
		case r >= 'a' && r <= 'z':
			b = append(b, r)
		case r >= '0' && r <= '9':
			b = append(b, r)
		default:
			b = append(b, '-')
		}
	}
	return strings.Trim(collapseDashes(string(b)), "-")
}

func collapseDashes(s string) string {
	for strings.Contains(s, "--") {
		s = strings.ReplaceAll(s, "--", "-")
	}
	return s
}

func parseBitableAppTokenFromURL(raw string) (string, bool) {
	if raw == "" {
		return "", false
	}
	if u, err := url.Parse(raw); err == nil {
		parts := strings.Split(strings.Trim(u.Path, "/"), "/")
		for i := 0; i < len(parts)-1; i++ {
			if parts[i] == "base" || parts[i] == "bitable" {
				tok := parts[i+1]
				if tok != "" {
					return tok, true
				}
			}
		}
		if tok := u.Query().Get("app_token"); tok != "" {
			return tok, true
		}
	}

	re := regexp.MustCompile(`\b(bas[a-zA-Z0-9]{6,})\b`)
	if m := re.FindStringSubmatch(raw); len(m) == 2 {
		return m[1], true
	}
	return "", false
}
