package main

import (
	"fmt"
	"strings"
)

func formatInfoTable(rows [][]string, emptyText string) string {
	return tableTextFromRows([]string{"name", "value"}, rows, emptyText)
}

func infoValue(value string) string {
	if strings.TrimSpace(value) == "" {
		return "-"
	}
	return value
}

func infoValueBoolPtr(value *bool) string {
	if value == nil {
		return "-"
	}
	if *value {
		return "true"
	}
	return "false"
}

func infoValueFloatPtr(value *float64) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%g", *value)
}

func infoValueIntPtr(value *int) string {
	if value == nil {
		return "-"
	}
	return fmt.Sprintf("%d", *value)
}

func infoValueIntZeroDash(value int) string {
	if value == 0 {
		return "-"
	}
	return fmt.Sprintf("%d", value)
}
