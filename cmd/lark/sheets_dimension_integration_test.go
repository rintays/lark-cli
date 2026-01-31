package main

import (
	"context"
	"testing"

	"lark/internal/testutil"
)

func TestIntegrationSheetsRowsInsertDelete(t *testing.T) {
	testutil.RequireIntegration(t)

	fx := getIntegrationFixtures(t)
	spreadsheetToken := fx.SpreadsheetToken
	sheetID := fx.SheetID

	ctx := context.Background()
	sdk := fx.SDK
	tenantToken := fx.Token

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
