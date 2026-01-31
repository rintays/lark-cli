package main

import (
	"context"
	"errors"
	"strings"

	"lark/internal/larksdk"
)

func searchDriveFilesInFolder(ctx context.Context, state *appState, userToken, query string, fileTypes []string, folderID string, limit, pages int) ([]larksdk.DriveFile, error) {
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
	folderID = strings.TrimSpace(folderID)
	if folderID == "" {
		return nil, errors.New("folder id is required")
	}

	files := make([]larksdk.DriveFile, 0, limit)
	pageToken := ""
	remaining := limit
	pageCount := 0
	for {
		if pageCount >= pages {
			break
		}
		pageCount++

		pageSize := remaining
		if pageSize > maxDrivePageSize {
			pageSize = maxDrivePageSize
		}

		result, err := state.SDK.SearchDriveFilesWithUserToken(ctx, userToken, larksdk.SearchDriveFilesRequest{
			Query:       query,
			FileTypes:   fileTypes,
			FolderToken: folderID,
			PageSize:    pageSize,
			PageToken:   pageToken,
		})
		if err != nil {
			return nil, withUserScopeHintForCommand(state, err)
		}
		files = append(files, result.Files...)
		if len(files) >= limit || !result.HasMore {
			break
		}
		remaining = limit - len(files)
		pageToken = result.PageToken
		if pageToken == "" {
			break
		}
	}
	if len(files) > limit {
		files = files[:limit]
	}
	return files, nil
}
