package main

import (
	"context"
	"errors"
	"strings"

	"github.com/spf13/cobra"
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
				return errors.New("table-id is required")
			}
			if recordID == "" {
				return errors.New("record-id is required")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			fieldsMap, err := parseBaseRecordFields(fieldsJSON, fields)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			record, err := state.SDK.UpdateBaseRecord(context.Background(), token, appToken, tableID, recordID, fieldsMap)
			if err != nil {
				return err
			}
			payload := map[string]any{"record": record}
			text := tableTextRow(
				[]string{"record_id", "created_time", "last_modified_time"},
				[]string{record.RecordID, record.CreatedTime.String(), record.LastModifiedTime.String()},
			)
			return state.Printer.Print(payload, text)
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
