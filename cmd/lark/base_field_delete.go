package main

import (
	"context"
	"errors"
	"strconv"

	"github.com/spf13/cobra"
)

func newBaseFieldDeleteCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldID string

	cmd := &cobra.Command{
		Use:   "delete <table-id> <field-id>",
		Short: "Delete a Bitable field",
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
				if fieldID != "" && fieldID != args[1] {
					return errors.New("field-id provided twice")
				}
				if err := cmd.Flags().Set("field-id", args[1]); err != nil {
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
			res, err := state.SDK.DeleteBaseField(context.Background(), token, appToken, tableID, fieldID)
			if err != nil {
				return err
			}
			payload := map[string]any{"result": res}
			text := tableTextRow([]string{"deleted", "field_id"}, []string{strconv.FormatBool(res.Deleted), res.FieldID})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldID, "field-id", "", "Bitable field id (or provide as positional argument)")
	_ = cmd.MarkFlagRequired("app-token")
	_ = cmd.MarkFlagRequired("table-id")
	_ = cmd.MarkFlagRequired("field-id")
	return cmd
}
