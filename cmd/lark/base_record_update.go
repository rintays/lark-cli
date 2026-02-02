package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseRecordUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string
	var fieldsJSON string
	var fields []string

	cmd := &cobra.Command{
		Use:   "update <table-id> <record-id>",
		Short: "Update a Bitable record",
		Long: `Update a Bitable record.

Provide fields via --fields-json (JSON object) or repeatable --field. Use := to pass JSON-typed values; = always sends a string.

Value formats (write):
- text/email/barcode: "text"
- number/progress/currency/rating: 12.3
- single select: "Option"
- multi select: ["A","B"]
- date: 1674206443000 (ms)
- checkbox: true|false
- user: [{"id":"ou_x"}]
- group: [{"id":"oc_x"}]
- phone: "1302616xxxx"
- url: {"text":"Feishu","link":"https://..."}
- attachment: [{"file_token":"xxx"}]
- link/duplex link: ["rec_x","rec_y"]
- location: "116.397755,39.903179"`,
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			recordID = strings.TrimSpace(args[1])
			if tableID == "" {
				return argsUsageError(cmd, errors.New("table-id is required"))
			}
			if recordID == "" {
				return argsUsageError(cmd, errors.New("record-id is required"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				fieldsMap, err := parseBaseRecordFields(fieldsJSON, fields)
				if err != nil {
					return nil, "", err
				}
				record, err := sdk.UpdateBaseRecord(ctx, token, appToken, tableID, recordID, fieldsMap)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"record": record}
				text := tableTextRow(
					[]string{"record_id", "created_time", "last_modified_time"},
					[]string{record.RecordID, record.CreatedTime.String(), record.LastModifiedTime.String()},
				)
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&fieldsJSON, "fields-json", "", `Record fields JSON object (full typed payload; quote for shell).
Example: {"Title":"Task","Done":true,"Score":3.5,"Tags":["A","B"],"Owner":[{"id":"ou_x"}]}`)
	cmd.Flags().StringArrayVar(&fields, "field", nil, `Record field assignment (repeatable; use := for JSON-typed values).
Formats: <name>=<string>, <name>:=<json>, or name=<field>,value=<value> (value:=<json>).
Examples: --field Title=Task --field Done:=true --field Score:=3.5 --field Tags:='["A","B"]' --field Owner:='[{"id":"ou_x"}]'`)
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
