package main

import (
	"context"
	"errors"
	"strings"

	"lark/internal/larksdk"
)

const maxDocsSearchCount = 50
const maxDocsSearchWindow = 199

func docsSearchDriveFiles(ctx context.Context, state *appState, token, label, query string, docTypes []string, limit, pages int) ([]larksdk.DriveFile, error) {
	if state == nil {
		return nil, errors.New("state is required")
	}
	if state.SDK == nil {
		return nil, errors.New("sdk client is required")
	}
	if limit <= 0 {
		return nil, errors.New("limit must be greater than 0")
	}
	if pages <= 0 {
		return nil, errors.New("pages must be greater than 0")
	}
	normalizedTypes := normalizeDocsSearchTypes(docTypes)

	debugf(state, "%s search: query=%q types=%v limit=%d pages=%d\n", label, query, normalizedTypes, limit, pages)

	files := make([]larksdk.DriveFile, 0, limit)
	offset := 0
	remaining := limit
	pageCount := 0
	for {
		pageCount++
		count := remaining
		if count > maxDocsSearchCount {
			count = maxDocsSearchCount
		}
		maxAllowed := maxDocsSearchWindow - offset
		if maxAllowed <= 0 {
			break
		}
		if count > maxAllowed {
			count = maxAllowed
		}
		if count <= 0 {
			break
		}
		debugf(state, "%s search request: page=%d/%d count=%d offset=%d\n", label, pageCount, pages, count, offset)

		result, err := state.SDK.SearchDocsObjectsWithUserToken(ctx, token, larksdk.DocsSearchRequest{
			Query:    query,
			DocTypes: normalizedTypes,
			Count:    count,
			Offset:   offset,
		})
		if err != nil {
			return nil, withUserScopeHintForCommand(state, err)
		}
		debugf(state, "%s search response: entities=%d has_more=%t total=%d\n", label, len(result.Entities), result.HasMore, result.Total)
		files = append(files, mapDocsSearchEntities(result.Entities)...)
		if len(files) >= limit || !result.HasMore || pageCount >= pages {
			break
		}
		remaining = limit - len(files)
		offset += count
		if offset <= 0 {
			break
		}
	}
	if len(files) > limit {
		files = files[:limit]
	}
	return files, nil
}

func normalizeDocsSearchTypes(types []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(types))
	for _, t := range types {
		clean := normalizeDocsSearchType(strings.TrimSpace(strings.ToLower(t)))
		if clean == "" {
			continue
		}
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}

func normalizeDocsSearchType(t string) string {
	switch t {
	case "docx", "doc":
		return "doc"
	default:
		return t
	}
}

func mapDocsSearchEntities(entities []larksdk.DocsSearchEntity) []larksdk.DriveFile {
	files := make([]larksdk.DriveFile, 0, len(entities))
	for _, entity := range entities {
		fileType := normalizeDocsSearchResultType(entity.DocsType)
		url := entity.OpenURL
		if url == "" {
			url = entity.URL
		}
		files = append(files, larksdk.DriveFile{
			Token:    entity.DocsToken,
			Name:     entity.Title,
			FileType: fileType,
			OwnerID:  entity.OwnerID,
			URL:      url,
		})
	}
	return files
}

func normalizeDocsSearchResultType(t string) string {
	clean := strings.TrimSpace(strings.ToLower(t))
	if clean == "doc" {
		return "docx"
	}
	return clean
}
