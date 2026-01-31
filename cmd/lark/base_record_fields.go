package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

func parseBaseRecordFields(raw string, entries []string) (map[string]any, error) {
	raw = strings.TrimSpace(raw)
	if raw != "" && len(entries) > 0 {
		return nil, errors.New("fields-json and field are mutually exclusive")
	}
	if raw != "" {
		var fields map[string]any
		if err := json.Unmarshal([]byte(raw), &fields); err != nil {
			return nil, fmt.Errorf("fields-json must be a JSON object: %w", err)
		}
		if fields == nil {
			return nil, errors.New("fields-json must be a JSON object")
		}
		return fields, nil
	}
	if len(entries) == 0 {
		return nil, errors.New("fields-json or field is required")
	}
	return parseBaseRecordFieldAssignments(entries)
}

func parseBaseRecordFieldAssignments(entries []string) (map[string]any, error) {
	fields := map[string]any{}
	for _, entry := range entries {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		key, value, err := parseBaseRecordFieldAssignment(entry)
		if err != nil {
			return nil, err
		}
		if _, exists := fields[key]; exists {
			return nil, fmt.Errorf("field %q provided twice", key)
		}
		fields[key] = value
	}
	if len(fields) == 0 {
		return nil, errors.New("fields-json or field is required")
	}
	return fields, nil
}

func parseBaseRecordFieldAssignment(raw string) (string, any, error) {
	if strings.Contains(raw, ",") && (strings.Contains(raw, "name=") || strings.Contains(raw, "value=") || strings.Contains(raw, "value:=")) {
		return parseBaseRecordNameValueAssignment(raw)
	}
	return parseBaseRecordKeyValueAssignment(raw)
}

func parseBaseRecordNameValueAssignment(raw string) (string, any, error) {
	var name string
	var valueRaw string
	var valueJSON bool
	valueSet := false

	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, value, isJSON, err := parseAssignmentToken(part)
		if err != nil {
			return "", nil, err
		}
		switch key {
		case "name":
			name = value
		case "value":
			valueRaw = value
			valueJSON = isJSON
			valueSet = true
		default:
			return "", nil, fmt.Errorf("field assignment %q must use name= and value=", raw)
		}
	}
	if name == "" {
		return "", nil, fmt.Errorf("field assignment %q is missing name=", raw)
	}
	if !valueSet {
		return "", nil, fmt.Errorf("field assignment %q is missing value=", raw)
	}
	value, err := parseAssignmentValue(valueRaw, valueJSON)
	if err != nil {
		return "", nil, fmt.Errorf("field %q value: %w", name, err)
	}
	return name, value, nil
}

func parseBaseRecordKeyValueAssignment(raw string) (string, any, error) {
	key, valueRaw, isJSON, err := parseAssignmentToken(raw)
	if err != nil {
		return "", nil, err
	}
	if key == "" {
		return "", nil, fmt.Errorf("field assignment %q is missing field name", raw)
	}
	value, err := parseAssignmentValue(valueRaw, isJSON)
	if err != nil {
		return "", nil, fmt.Errorf("field %q value: %w", key, err)
	}
	return key, value, nil
}

func parseAssignmentToken(raw string) (string, string, bool, error) {
	if strings.Contains(raw, ":=") {
		parts := strings.SplitN(raw, ":=", 2)
		if len(parts) != 2 {
			return "", "", false, fmt.Errorf("invalid field assignment %q", raw)
		}
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), true, nil
	}
	if strings.Contains(raw, "=") {
		parts := strings.SplitN(raw, "=", 2)
		if len(parts) != 2 {
			return "", "", false, fmt.Errorf("invalid field assignment %q", raw)
		}
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), false, nil
	}
	return "", "", false, fmt.Errorf("invalid field assignment %q", raw)
}

func parseAssignmentValue(raw string, isJSON bool) (any, error) {
	if !isJSON {
		return raw, nil
	}
	if strings.TrimSpace(raw) == "" {
		return nil, errors.New("JSON value is required")
	}
	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		return nil, fmt.Errorf("invalid JSON: %w", err)
	}
	return value, nil
}
