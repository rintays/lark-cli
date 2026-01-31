package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newBaseFieldCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldName string
	var fieldType int
	var propertyJSON string
	var descriptionJSON string

	cmd := &cobra.Command{
		Use:   "create <table-id> <name>",
		Short: "Create a Bitable field",
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
				if fieldName != "" && fieldName != args[1] {
					return errors.New("name provided twice")
				}
				if err := cmd.Flags().Set("name", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(tableID) == "" {
				return errors.New("table-id is required")
			}
			if strings.TrimSpace(fieldName) == "" {
				return errors.New("name is required")
			}
			if fieldType == 0 {
				return errors.New("type is required")
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
			field, err := state.SDK.CreateBaseField(context.Background(), token, appToken, tableID, fieldName, fieldType, property, description)
			if err != nil {
				return err
			}
			payload := map[string]any{"field": field}
			text := tableTextRow([]string{"field_id", "name", "type"}, []string{field.FieldID, field.FieldName, fmt.Sprintf("%d", field.Type)})
			return state.Printer.Print(payload, text)
		},
	}

	cmd.Flags().StringVar(&appToken, "app-token", "", "Bitable app token")
	cmd.Flags().StringVar(&tableID, "table-id", "", "Bitable table id (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldName, "name", "", "Field name (or provide as positional argument)")
	cmd.Flags().IntVar(&fieldType, "type", 0, "Field type (required)")
	cmd.Flags().StringVar(&propertyJSON, "property-json", "", "Field property JSON (object)")
	cmd.Flags().StringVar(&descriptionJSON, "description-json", "", "Field description JSON (object)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func parseOptionalJSONObject(flagName, raw string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil, nil
	}
	var payload map[string]any
	if err := json.Unmarshal([]byte(raw), &payload); err != nil {
		return nil, fmt.Errorf("%s must be a JSON object: %w", flagName, err)
	}
	if payload == nil {
		return nil, fmt.Errorf("%s must be a JSON object", flagName)
	}
	return payload, nil
}
