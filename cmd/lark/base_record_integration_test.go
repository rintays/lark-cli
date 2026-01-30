package main

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"

	"lark/internal/config"
	"lark/internal/larksdk"
	"lark/internal/testutil"
)

func TestIntegrationBaseRecordCreateGetDelete(t *testing.T) {
	testutil.RequireIntegration(t)

	if os.Getenv("LARK_APP_ID") == "" || os.Getenv("LARK_APP_SECRET") == "" {
		t.Skip("missing LARK_APP_ID/LARK_APP_SECRET")
	}
	appToken := os.Getenv("LARK_TEST_APP_TOKEN")
	tableID := os.Getenv("LARK_TEST_TABLE_ID")
	if appToken == "" || tableID == "" {
		t.Skip("missing LARK_TEST_APP_TOKEN/LARK_TEST_TABLE_ID")
	}

	cfg := config.Default()
	// pull app_id/app_secret from env
	if cfg.AppID == "" {
		cfg.AppID = os.Getenv("LARK_APP_ID")
	}
	if cfg.AppSecret == "" {
		cfg.AppSecret = os.Getenv("LARK_APP_SECRET")
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

	uniqueValue := fmt.Sprintf("clawdbot-it-%d", time.Now().UnixNano())
	record, err := sdk.CreateBaseRecord(ctx, tenantToken, appToken, tableID, map[string]any{fieldName: uniqueValue})
	if err != nil {
		t.Fatalf("create record: %v", err)
	}
	if record.RecordID == "" {
		t.Fatal("create record returned empty record_id")
	}
	recordID := record.RecordID

	defer func() {
		res, derr := sdk.DeleteBaseRecord(ctx, tenantToken, appToken, tableID, recordID)
		if derr != nil {
			t.Fatalf("delete record: %v", derr)
		}
		if !res.Deleted {
			t.Fatalf("expected deleted=true, got false (record_id=%s)", res.RecordID)
		}
	}()

	got, err := sdk.GetBaseRecord(ctx, tenantToken, appToken, tableID, recordID)
	if err != nil {
		t.Fatalf("get record: %v", err)
	}
	if got.RecordID != recordID {
		t.Fatalf("expected record_id %q, got %q", recordID, got.RecordID)
	}
	if got.Fields != nil {
		if v, ok := got.Fields[fieldName]; ok {
			if fmt.Sprint(v) != uniqueValue {
				t.Fatalf("expected field %q=%q, got %#v", fieldName, uniqueValue, v)
			}
		}
	}
}
