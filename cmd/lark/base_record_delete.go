package main

import (
	"context"
	"errors"

	"github.com/spf13/cobra"
)

func newBaseRecordDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var recordID string

	cmd := &cobra.Command{
		Use:   "delete <table-id> <record-id>",
		Short: "Delete a Bitable record",
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
			result, err := state.SDK.DeleteBaseRecord(context.Background(), token, appToken, tableID, recordID)
			if err != nil {
				return err
			}
			payload := map[string]any{"record_id": result.RecordID, "deleted": result.Deleted}
			text := "deleted"
			if result.RecordID != "" {
				text = result.RecordID
			}
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&recordID, "record-id", "", "Bitable record id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("record-id")
	return cmd
}
