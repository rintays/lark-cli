package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "bases",
		Aliases: []string{"base"},
		Short:   "Manage Bitable bases",
		Long: `Bitable (Base) is Lark/Feishu's database product.

A base (also called an app) is the top-level container. Each base contains one or more tables.
- Table: a grid that defines fields (columns) and stores records (rows).
- Field: a column definition (type + name) used by every record in the table.
- Record: a single row of data for the table's fields.
- View: a saved presentation of a table (filters/sorts/grouping/hidden columns); it doesn't change the underlying records.

Relationships: app -> tables -> fields + records; views belong to a table.
Most subcommands require the Bitable app token to identify the base.`,
	}
	cmd.AddCommand(newBaseListCmd(state))
	cmd.AddCommand(newBaseAppCmd(state))
	cmd.AddCommand(newBaseTableCmd(state))
	cmd.AddCommand(newBaseFieldCmd(state))
	cmd.AddCommand(newBaseViewCmd(state))
	cmd.AddCommand(newBaseRecordCmd(state))
	return cmd
}

func newBaseAppCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "app",
		Short: "Manage Bitable apps",
	}
	cmd.AddCommand(newBaseAppListCmd(state))
	cmd.AddCommand(newBaseAppCreateCmd(state))
	cmd.AddCommand(newBaseAppCopyCmd(state))
	cmd.AddCommand(newBaseAppGetCmd(state))
	cmd.AddCommand(newBaseAppUpdateCmd(state))
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
	cmd.AddCommand(newBaseFieldCreateCmd(state))
	cmd.AddCommand(newBaseFieldUpdateCmd(state))
	cmd.AddCommand(newBaseFieldDeleteCmd(state))
	cmd.AddCommand(newBaseFieldListCmd(state))
	cmd.AddCommand(newBaseFieldTypesCmd(state))
	return cmd
}

func newBaseViewCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view",
		Short: "Manage Bitable views",
	}
	cmd.AddCommand(newBaseViewCreateCmd(state))
	cmd.AddCommand(newBaseViewDeleteCmd(state))
	cmd.AddCommand(newBaseViewListCmd(state))
	return cmd
}

func newBaseRecordCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record",
		Short: "Manage Bitable records",
	}
	cmd.AddCommand(newBaseRecordCreateCmd(state))
	cmd.AddCommand(newBaseRecordBatchCreateCmd(state))
	cmd.AddCommand(newBaseRecordBatchUpdateCmd(state))
	cmd.AddCommand(newBaseRecordBatchDeleteCmd(state))
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
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(tableID) == "" {
					return errors.New("table-id is required")
				}
				return nil
			}
			if tableID != "" && tableID != args[0] {
				return errors.New("table-id provided twice")
			}
			return cmd.Flags().Set("table-id", args[0])
		},
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func newBaseViewListCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bitable views",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(1)(cmd, args); err != nil {
				return err
			}
			if len(args) == 0 {
				if strings.TrimSpace(tableID) == "" {
					return errors.New("table-id is required")
				}
				return nil
			}
			if tableID != "" && tableID != args[0] {
				return errors.New("table-id provided twice")
			}
			return cmd.Flags().Set("table-id", args[0])
		},
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func newBaseRecordInfoCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string

	cmd := &cobra.Command{
		Use:   "info <table-id> <record-id>",
		Short: "Show a Bitable record",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.MaximumNArgs(2)(cmd, args); err != nil {
				return err
			}
			if len(args) > 0 {
				if tableID != "" && tableID != args[0] {
					return errors.New("table-id provided twice")
				}
				if err := cmd.Flags().Set("table-id", args[0]); err != nil {
					return err
				}
			}
			if len(args) > 1 {
				if recordID != "" && recordID != args[1] {
					return errors.New("record-id provided twice")
				}
				if err := cmd.Flags().Set("record-id", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(tableID) == "" {
				return errors.New("table-id is required")
			}
			if strings.TrimSpace(recordID) == "" {
				return errors.New("record-id is required")
			}
			return nil
		},
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
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
