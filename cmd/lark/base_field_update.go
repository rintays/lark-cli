package main

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
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
			if err := cobra.ExactArgs(2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			fieldID = strings.TrimSpace(args[1])
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
			return runWithToken(cmd, state, nil, nil, func(ctx context.Context, sdk *larksdk.Client, token string, tokenType tokenType) (any, string, error) {
				property, err := parseOptionalJSONObject("property-json", propertyJSON)
				if err != nil {
					return nil, "", err
				}
				description, err := parseOptionalJSONObject("description-json", descriptionJSON)
				if err != nil {
					return nil, "", err
				}
				field, err := sdk.UpdateBaseField(ctx, token, appToken, tableID, fieldID, strings.TrimSpace(fieldName), property, description)
				if err != nil {
					return nil, "", err
				}
				payload := map[string]any{"field": field}
				text := tableTextRow([]string{"field_id", "field_name", "type"}, []string{field.FieldID, field.FieldName, fmt.Sprintf("%d", field.Type)})
				return payload, text, nil
			})
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&fieldName, "name", "", "New field name")
	cmd.Flags().StringVar(&propertyJSON, "property-json", "", "Field property JSON (object; see `bases field types` for hints)")
	cmd.Flags().StringVar(&descriptionJSON, "description-json", "", "Field description JSON (object)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}
