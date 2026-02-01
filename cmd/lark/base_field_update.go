package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseFieldUpdateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldID string
	var fieldName string
	var propertyJSON string
	var descriptionJSON string

	cmd := &cobra.Command{
		Use:   "update <table-id> <field-id>",
		Short: "Update a Bitable field",
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
			if strings.TrimSpace(tableID) == "" {
				return errors.New("table-id is required")
			}
			if strings.TrimSpace(fieldID) == "" {
				return errors.New("field-id is required")
			}
			name := strings.TrimSpace(fieldName)
			if name == "" && strings.TrimSpace(propertyJSON) == "" && strings.TrimSpace(descriptionJSON) == "" {
				return errors.New("at least one of --name, --property-json, or --description-json is required")
			}
			if cmd.Flags().Changed("name") && name == "" {
				return errors.New("name must not be empty")
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.SDK == nil {
				return errors.New("sdk client is required")
			}
			property, err := parseOptionalJSONObject("property-json", propertyJSON)
			if err != nil {
				return err
			}
			description, err := parseOptionalJSONObject("description-json", descriptionJSON)
			if err != nil {
				return err
			}
			token, err := tokenFor(context.Background(), state, tokenTypesTenantOrUser)
			if err != nil {
				return err
			}
			field, err := state.SDK.UpdateBaseField(context.Background(), token, appToken, tableID, fieldID, strings.TrimSpace(fieldName), property, description)
			if err != nil {
				return err
			}
			payload := map[string]any{"field": field}
			text := tableTextRow([]string{"field_id", "field_name", "type"}, []string{field.FieldID, field.FieldName, fmt.Sprintf("%d", field.Type)})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldID, "field-id", "", "Bitable field id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldName, "name", "", "New field name")
	cmd.Flags().StringVar(&propertyJSON, "property-json", "", "Field property JSON (object; see `bases field types` for hints)")
	cmd.Flags().StringVar(&descriptionJSON, "description-json", "", "Field description JSON (object)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
