package main

import (
	"context"
	"errors"
	"strconv"
	"strings"

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
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return err
			}
			tableID = strings.TrimSpace(args[0])
			fieldID = strings.TrimSpace(args[1])
			if tableID == "" {
				return errors.New("table-id is required")
			}
			if fieldID == "" {
				return errors.New("field-id is required")
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
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
