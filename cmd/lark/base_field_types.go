package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

type baseFieldTypeInfo struct {
	ID           int
	Name         string
	PropertyHint string
}

func baseFieldTypeInfos() []baseFieldTypeInfo {
	return []baseFieldTypeInfo{
		{ID: 1, Name: "text", PropertyHint: "-"},
		{ID: 2, Name: "number", PropertyHint: "property.formatter"},
		{ID: 3, Name: "single_select", PropertyHint: "property.options[{name,color}]"},
		{ID: 4, Name: "multi_select", PropertyHint: "property.options[{name,color}]"},
		{ID: 5, Name: "date", PropertyHint: "property.date_formatter, property.auto_fill"},
		{ID: 7, Name: "checkbox", PropertyHint: "-"},
		{ID: 11, Name: "user", PropertyHint: "property.multiple"},
		{ID: 13, Name: "phone", PropertyHint: "-"},
		{ID: 15, Name: "url", PropertyHint: "-"},
		{ID: 17, Name: "attachment", PropertyHint: "-"},
		{ID: 18, Name: "single_link", PropertyHint: "property.table_id, property.multiple"},
		{ID: 20, Name: "formula", PropertyHint: "property.formula_expression; property.type{data_type,ui_type}"},
		{ID: 21, Name: "duplex_link", PropertyHint: "property.table_id, property.back_field_name, property.multiple"},
		{ID: 22, Name: "location", PropertyHint: "property.location{input_type}"},
		{ID: 23, Name: "group", PropertyHint: "-"},
		{ID: 1001, Name: "created_time", PropertyHint: "property.date_formatter, property.auto_fill"},
		{ID: 1002, Name: "last_modified_time", PropertyHint: "property.date_formatter"},
		{ID: 1003, Name: "created_by", PropertyHint: "-"},
		{ID: 1004, Name: "modified_by", PropertyHint: "-"},
		{ID: 1005, Name: "auto_number", PropertyHint: "property.auto_serial{type,options[]}"},
	}
}

func newBaseFieldTypesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "types",
		Short: "List Bitable field types and property hints",
		RunE: func(cmd *cobra.Command, args []string) error {
			items := baseFieldTypeInfos()
			lines := make([]string, 0, len(items))
			for _, item := range items {
				lines = append(lines, fmt.Sprintf("%d\t%s\t%s", item.ID, item.Name, item.PropertyHint))
			}
			payload := map[string]any{"types": items}
			text := tableText([]string{"type_id", "name", "property_hint"}, lines, "no field types found")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
