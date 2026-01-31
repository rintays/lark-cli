package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func newSheetsCreateCmd(state *appState) *cobra.Command {
	var title string
	var folderID string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Sheets (spreadsheet) file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			spreadsheetToken, err := state.SDK.CreateSpreadsheet(context.Background(), token, title, folderID)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"spreadsheet_token": spreadsheetToken,
				"title":             title,
				"folder_id":         folderID,
			}
			text := tableTextRow([]string{"spreadsheet_token", "title"}, []string{spreadsheetToken, title})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "spreadsheet title")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
