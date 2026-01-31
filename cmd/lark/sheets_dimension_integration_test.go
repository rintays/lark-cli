package main

import (
	"context"
	"os"
	"testing"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

func TestIntegrationSheetsRowsInsertDelete(t *testing.T) {
	testutil.RequireIntegration(t)

	appID := os.Getenv("LARK_APP_ID")
	appSecret := os.Getenv("LARK_APP_SECRET")
	if appID == "" || appSecret == "" {
		t.Skip("missing LARK_APP_ID/LARK_APP_SECRET")
	}
	spreadsheetToken := os.Getenv("LARK_TEST_SHEET_ID")
	sheetID := os.Getenv("LARK_TEST_SHEET_SHEET_ID")
	if spreadsheetToken == "" || sheetID == "" {
		t.Skip("missing LARK_TEST_SHEET_ID/LARK_TEST_SHEET_SHEET_ID")
	}

	cfg := config.Default()
	// pull app_id/app_secret from env
	if cfg.AppID == "" {
		cfg.AppID = appID
	}
	if cfg.AppSecret == "" {
		cfg.AppSecret = appSecret
	}
	if baseURL := os.Getenv("LARK_BASE_URL"); baseURL != "" {
		cfg.BaseURL = baseURL
	}

	sdk, err := larksdk.New(cfg)
	if err != nil {
		t.Fatalf("sdk init: %v", err)
	}

	ctx := context.Background()
	tenantToken, _, err := sdk.TenantAccessToken(ctx)
	if err != nil {
		t.Fatalf("tenant token: %v", err)
	}
	if tenantToken == "" {
		t.Fatal("empty tenant token")
	}

	startIndex := 0
	count := 1
	endIndex := startIndex + count

	insertRes, err := sdk.InsertSheetRows(ctx, tenantToken, spreadsheetToken, sheetID, startIndex, count)
	if err != nil {
		t.Fatalf("insert sheet rows: %v", err)
	}
	if insertRes.StartIndex != startIndex || insertRes.Count != count || insertRes.EndIndex != endIndex {
		t.Fatalf("unexpected insert result: got start=%d count=%d end=%d", insertRes.StartIndex, insertRes.Count, insertRes.EndIndex)
	}

	deleteRes, err := sdk.DeleteSheetRows(ctx, tenantToken, spreadsheetToken, sheetID, startIndex, count)
	if err != nil {
		t.Fatalf("delete sheet rows: %v", err)
	}
	if deleteRes.StartIndex != startIndex || deleteRes.Count != count || deleteRes.EndIndex != endIndex {
		t.Fatalf("unexpected delete result: got start=%d count=%d end=%d", deleteRes.StartIndex, deleteRes.Count, deleteRes.EndIndex)
	}
}
