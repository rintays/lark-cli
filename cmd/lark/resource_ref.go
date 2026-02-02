package main

import (
	"errors"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

var resourceURLPattern = regexp.MustCompile(`(?i)/(docx|docs|doc|sheet|sheets|file|drive|wiki|bitable|base|minutes|slides|mindnote)/([^/?#]+)`)

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
	if token, kind := tokenFromPath(parsed.Path); token != "" {
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
		"app_token",
		"base_token",
		"wiki_token",
		"node_token",
	} {
		if value := strings.TrimSpace(query.Get(key)); value != "" {
			return value, normalizeQueryKind(key)
		}
	}
	return "", ""
}

func tokenFromPath(path string) (string, string) {
	trimmed := strings.Trim(strings.TrimSpace(path), "/")
	if trimmed == "" {
		return "", ""
	}
	parts := strings.Split(trimmed, "/")
	for i := 0; i < len(parts)-1; i++ {
		kind := normalizeResourceKind(parts[i])
		if kind == "" {
			continue
		}
		token := strings.TrimSpace(parts[i+1])
		if token == "" {
			continue
		}
		return token, kind
	}
	return "", ""
}

func normalizeQueryKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "doc_token", "document_id", "docx":
		return "docx"
	case "sheet_token", "spreadsheet_token":
		return "sheet"
	case "file_token":
		return "file"
	case "app_token", "base_token":
		return "bitable"
	case "wiki_token", "node_token":
		return "wiki"
	default:
		return ""
	}
}

func normalizeResourceKind(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case "doc", "docs", "docx":
		return "docx"
	case "sheet", "sheets", "spreadsheet":
		return "sheet"
	case "drive", "file":
		return "file"
	case "bitable", "base":
		return "bitable"
	case "wiki":
		return "wiki"
	case "minutes":
		return "minutes"
	case "slides", "slide":
		return "slides"
	case "mindnote":
		return "mindnote"
	default:
		return ""
	}
}
