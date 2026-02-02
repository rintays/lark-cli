package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/larksdk"
)

func newBaseFieldCreateCmd(state *appState) *cobra.Command {
	var appToken string
	var tableID string
	var fieldName string
	var fieldType int
	var fieldTypeName string
	var propertyJSON string
	var descriptionJSON string

	cmd := &cobra.Command{
		Use:   "create <table-id> <name>",
		Short: "Create a Bitable field",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.RangeArgs(1, 2)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			tableID = strings.TrimSpace(args[0])
			if tableID == "" {
				return errors.New("table-id is required")
			}
			if len(args) > 1 {
				if fieldName != "" && fieldName != args[1] {
					return errors.New("name provided twice")
				}
				if err := cmd.Flags().Set("name", args[1]); err != nil {
					return err
				}
			}
			if strings.TrimSpace(fieldName) == "" {
				return errors.New("name is required")
			}
			fieldTypeNameSet := cmd.Flags().Changed("field-type")
			fieldTypeIDSet := cmd.Flags().Changed("type")
			if fieldTypeNameSet && fieldTypeIDSet {
				return errors.New("--field-type and --type are mutually exclusive")
			}
			if fieldTypeNameSet {
				v, err := parseBaseFieldType(fieldTypeName)
				if err != nil {
					return err
				}
				fieldType = v
			}
			if !fieldTypeNameSet && !fieldTypeIDSet {
				fieldType = 1 // default to text
			}
			if fieldType <= 0 {
				return fmt.Errorf("invalid field type: %d", fieldType)
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
				field, err := sdk.CreateBaseField(ctx, token, appToken, tableID, fieldName, fieldType, property, description)
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
	cmd.Flags().StringVar(&fieldName, "name", "", "Field name (or provide as positional argument)")
	cmd.Flags().StringVar(&fieldTypeName, "field-type", "", "Field type name or id (e.g. text, number, 1). Prefer this over --type")
	cmd.Flags().IntVar(&fieldType, "type", 0, "Field type id (deprecated; use --field-type)")
	cmd.Flags().StringVar(&propertyJSON, "property-json", "", "Field property JSON (object; see `bases field types` for hints)")
	cmd.Flags().StringVar(&descriptionJSON, "description-json", "", "Field description JSON (object)")
	_ = cmd.MarkFlagRequired("app-token")
	return cmd
}

func parseBaseFieldType(raw string) (int, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0, errors.New("field type is required")
	}
	// numeric form
	if n, err := parseInt(raw); err == nil {
		if n <= 0 {
			return 0, fmt.Errorf("invalid field type: %s", raw)
		}
		return n, nil
	}
	// name form
	normalized := strings.ToLower(raw)
	for _, info := range baseFieldTypeInfos() {
		if info.Name == normalized {
			return info.ID, nil
		}
	}
	return 0, fmt.Errorf("unknown field type: %s (see `bases field types`)", raw)
}

func parseInt(raw string) (int, error) {
	var n int
	_, err := fmt.Sscanf(raw, "%d", &n)
	if err != nil {
		return 0, err
	}
	return n, nil
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
