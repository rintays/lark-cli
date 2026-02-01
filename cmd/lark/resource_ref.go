package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var resourceURLPattern = regexp.MustCompile(`(?i)/(docx|docs|doc|sheet|sheets|file|drive|wiki|bitable|minutes)/([^/?#]+)`)

func parseResourceRef(input string) (string, string, error) {
	raw := strings.TrimSpace(input)
	if raw == "" {
		return "", "", errors.New("resource reference is required")
	}
	if !strings.Contains(raw, "://") {
		return raw, "", nil
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		return "", "", fmt.Errorf("invalid resource URL: %w", err)
	}
	if token, kind := tokenFromQuery(parsed); token != "" {
		return token, kind, nil
	}
	if match := resourceURLPattern.FindStringSubmatch(parsed.Path); len(match) == 3 {
		return strings.TrimSpace(match[2]), normalizeResourceKind(match[1]), nil
	}
	return "", "", fmt.Errorf("unsupported resource URL: %s", raw)
}

func tokenFromQuery(u *url.URL) (string, string) {
	if u == nil {
		return "", ""
	}
	query := u.Query()
	for _, key := range []string{
		"token",
		"file_token",
		"doc_token",
		"document_id",
		"docx",
		"sheet_token",
		"spreadsheet_token",
	} {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return value, normalizeResourceKind(key)
		}
	}
	return "", ""
}

func normalizeResourceKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "doc", "docs", "docx", "doc_token", "document_id":
		return "docx"
	case "sheet", "sheets", "sheet_token", "spreadsheet", "spreadsheet_token":
		return "sheet"
	case "drive":
		return "file"
	default:
		return strings.ToLower(strings.TrimSpace(raw))
	}
}
