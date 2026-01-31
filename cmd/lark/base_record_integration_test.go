package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"lark/internal/testutil"
)

func TestIntegrationBaseRecordCreateGetDelete(t *testing.T) {
	testutil.RequireIntegration(t)

	_ = testutil.RequireEnv(t, "LARK_APP_ID")
	_ = testutil.RequireEnv(t, "LARK_APP_SECRET")
	appToken := testutil.RequireEnv(t, "LARK_TEST_APP_TOKEN")
	tableID := testutil.RequireEnv(t, "LARK_TEST_TABLE_ID")

	fx := getIntegrationFixtures(t)
	sdk := fx.SDK
	ctx := context.Background()
	tenantToken := fx.Token

	fieldName := os.Getenv("LARK_TEST_FIELD_NAME")
	if fieldName == "" {
		fields, err := sdk.ListBaseFields(ctx, tenantToken, appToken, tableID)
		if err != nil {
			t.Fatalf("list base fields: %v", err)
		}
		for _, f := range fields.Items {
			if f.Type == 1 && f.FieldName != "" {
				fieldName = f.FieldName
				break
			}
		}
		if fieldName == "" {
			t.Skip("could not auto-select a text field; set LARK_TEST_FIELD_NAME")
		}
	}

	uniqueValue := fmt.Sprintf("%sbase-record-%d", integrationFixturePrefix, time.Now().UnixNano())
	record, err := sdk.CreateBaseRecord(ctx, tenantToken, appToken, tableID, map[string]any{fieldName: uniqueValue})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if record.RecordID == "" {
		t.Fatal("create record returned empty record_id")
	}
	recordID := record.RecordID

	defer func() {
		if recordID == "" {
			return
		}
		if err := sdk.DeleteBaseRecord(ctx, tenantToken, appToken, tableID, recordID); err != nil {
			t.Fatalf("delete record: %v", err)
		}
	}()

	got, err := sdk.GetBaseRecord(ctx, tenantToken, appToken, tableID, recordID)
	if err != nil {
		t.Fatalf("get record: %v", err)
	}
	if got.RecordID != recordID {
		t.Fatalf("expected record_id %s, got %s", recordID, got.RecordID)
	}
}
