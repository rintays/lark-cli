package main

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"
)

func newBaseCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "base",
		Short: "Manage Bitable bases",
	}
	cmd.AddCommand(newBaseTableCmd(state))
	cmd.AddCommand(newBaseFieldCmd(state))
	cmd.AddCommand(newBaseViewCmd(state))
	cmd.AddCommand(newBaseRecordCmd(state))
	return cmd
}

func newBaseTableCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "table",
		Short: "Manage Bitable tables",
	}
	cmd.AddCommand(newBaseTableListCmd(state))
	cmd.AddCommand(newBaseTableCreateCmd(state))
	cmd.AddCommand(newBaseTableDeleteCmd(state))
	return cmd
}

func newBaseFieldCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "field",
		Short: "Manage Bitable fields",
	}
	cmd.AddCommand(newBaseFieldListCmd(state))
	return cmd
}

func newBaseViewCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage Bitable views",
	}
	cmd.AddCommand(newBaseViewListCmd(state))
	return cmd
}

func newBaseRecordCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage Bitable records",
	}
	cmd.AddCommand(newBaseRecordCreateCmd(state))
	cmd.AddCommand(newBaseRecordSearchCmd(state))
	cmd.AddCommand(newBaseRecordInfoCmd(state))
	cmd.AddCommand(newBaseRecordUpdateCmd(state))
	cmd.AddCommand(newBaseRecordDeleteCmd(state))
	return cmd
}

func newBaseTableListCmd(state *appState) *cobra.Command {
	var appToken string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bitable tables",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			result, err := state.SDK.ListBaseTables(context.Background(), token, appToken)
			if err != nil {
				return err
			}
			tables := result.Items
			payload := map[string]any{"tables": tables}
			lines := make([]string, 0, len(tables))
			for _, table := range tables {
				lines = append(lines, fmt.Sprintf("%s\t%s", table.TableID, table.Name))
			}
			text := tableText([]string{"table_id", "name"}, lines, "no tables found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func newBaseFieldListCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bitable fields",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			result, err := state.SDK.ListBaseFields(context.Background(), token, appToken, tableID)
			if err != nil {
				return err
			}
			fields := result.Items
			payload := map[string]any{"fields": fields}
			lines := make([]string, 0, len(fields))
			for _, field := range fields {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%d", field.FieldID, field.FieldName, field.Type))
			}
			text := tableText([]string{"field_id", "name", "type"}, lines, "no fields found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	return cmd
}

func newBaseViewListCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bitable views",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			result, err := state.SDK.ListBaseViews(context.Background(), token, appToken, tableID)
			if err != nil {
				return err
			}
			views := result.Items
			payload := map[string]any{"views": views}
			lines := make([]string, 0, len(views))
			for _, view := range views {
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s", view.ViewID, view.Name, view.ViewType))
			}
			text := tableText([]string{"view_id", "name", "type"}, lines, "no views found")
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	return cmd
}

func newBaseRecordInfoCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string

	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show a Bitable record",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			record, err := state.SDK.GetBaseRecord(context.Background(), token, appToken, tableID, recordID)
			if err != nil {
				return err
			}
			payload := map[string]any{"record": record}
			text := tableTextRow(
				[]string{"record_id", "created_time", "last_modified_time"},
				[]string{record.RecordID, record.CreatedTime, record.LastModifiedTime},
			)
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("record-id")
	return cmd
}
