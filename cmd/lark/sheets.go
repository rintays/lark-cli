package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larkapi"
)

func newSheetsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sheets",
		Short: "Read Sheets data",
	}
	cmd.AddCommand(newSheetsReadCmd(state))
	cmd.AddCommand(newSheetsMetadataCmd(state))
	return cmd
}

func newSheetsReadCmd(state *appState) *cobra.Command {
	var spreadsheetID string
	var sheetRange string

	cmd := &cobra.Command{
		Use:   "read",
		Short: "Read a range from Sheets",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			if sheetRange == "" {
				return errors.New("range is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			valueRange, err := state.Client.ReadSheetRange(context.Background(), token, spreadsheetID, sheetRange)
			if err != nil {
				return err
			}
			payload := map[string]any{"valueRange": valueRange}
			text := formatSheetValues(valueRange.Values)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	cmd.Flags().StringVar(&sheetRange, "range", "", "A1 range, e.g. Sheet1!A1:B2")
	return cmd
}

func newSheetsMetadataCmd(state *appState) *cobra.Command {
	var spreadsheetID string

	cmd := &cobra.Command{
		Use:   "metadata",
		Short: "Get spreadsheet metadata",
		RunE: func(cmd *cobra.Command, args []string) error {
			if spreadsheetID == "" {
				return errors.New("spreadsheet-id is required")
			}
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			metadata, err := state.Client.GetSpreadsheetMetadata(context.Background(), token, spreadsheetID)
			if err != nil {
				return err
			}
			payload := map[string]any{"metadata": metadata}
			text := formatSpreadsheetMetadata(metadata)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&spreadsheetID, "spreadsheet-id", "", "spreadsheet token")
	return cmd
}

func formatSheetValues(values [][]any) string {
	if len(values) == 0 {
		return "no values found"
	}
	lines := make([]string, 0, len(values))
	for _, row := range values {
		cells := make([]string, 0, len(row))
		for _, cell := range row {
			cells = append(cells, fmt.Sprint(cell))
		}
		lines = append(lines, strings.Join(cells, "\t"))
	}
	return strings.Join(lines, "\n")
}

func formatSpreadsheetMetadata(metadata larkapi.SpreadsheetMetadata) string {
	lines := make([]string, 0, len(metadata.Sheets)+1)
	if title := strings.TrimSpace(metadata.Properties.Title); title != "" {
		lines = append(lines, title)
	}
	for _, sheet := range metadata.Sheets {
		if name := strings.TrimSpace(sheet.Title); name != "" {
			lines = append(lines, name)
		}
	}
	if len(lines) == 0 {
		return "no metadata found"
	}
	return strings.Join(lines, "\n")
}
