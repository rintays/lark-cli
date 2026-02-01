package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newSheetsCreateCmd(state *appState) *cobra.Command {
	var title string
	var folderID string
	var sheetTitle string

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
			normalizedFolderID := strings.TrimSpace(folderID)
			if strings.EqualFold(normalizedFolderID, "root") {
				normalizedFolderID = ""
			}
			spreadsheetToken, err := state.SDK.CreateSpreadsheet(context.Background(), token, title, normalizedFolderID)
			if err != nil {
				return err
			}
			var defaultSheetID string
			var defaultSheetTitle string
			if state.SDK != nil {
				metadata, err := state.SDK.GetSpreadsheetMetadata(context.Background(), token, spreadsheetToken)
				if err != nil {
					if strings.TrimSpace(sheetTitle) != "" {
						return fmt.Errorf("resolve sheet id for --sheet-title: %w", err)
					}
				} else if len(metadata.Sheets) > 0 {
					defaultSheetID = strings.TrimSpace(metadata.Sheets[0].SheetID)
					defaultSheetTitle = strings.TrimSpace(metadata.Sheets[0].Title)
				}
			}
			if strings.TrimSpace(sheetTitle) != "" {
				if defaultSheetID == "" {
					return errors.New("sheet id is required to set --sheet-title")
				}
				if err := state.SDK.UpdateSpreadsheetSheetTitle(context.Background(), token, spreadsheetToken, defaultSheetID, sheetTitle); err != nil {
					return err
				}
				defaultSheetTitle = sheetTitle
			}
			payload := map[string]any{
				"spreadsheet_token": spreadsheetToken,
				"title":             title,
				"folder_id":         folderID,
			}
			headers := []string{"spreadsheet_token", "title"}
			values := []string{spreadsheetToken, title}
			if defaultSheetID != "" {
				defaultRange := fmt.Sprintf("%s!A1", defaultSheetID)
				payload["sheet_id"] = defaultSheetID
				payload["sheet_title"] = defaultSheetTitle
				payload["default_range"] = defaultRange
				headers = append(headers, "sheet_id", "sheet_title", "default_range")
				values = append(values, defaultSheetID, defaultSheetTitle, defaultRange)
			}
			text := tableTextRow(headers, values)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&title, "title", "", "spreadsheet title")
	cmd.Flags().StringVar(&sheetTitle, "sheet-title", "", "default sheet (tab) title")
	cmd.Flags().StringVar(&folderID, "folder-id", "", "Drive folder token (default: root; pass root or omit)")
	_ = cmd.MarkFlagRequired("title")
	return cmd
}
