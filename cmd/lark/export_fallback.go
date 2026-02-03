package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	"lark/internal/larksdk"
)

const exportTaskUserReloginHint = "lark auth user login --scopes \"offline_access drive:export:readonly\" --force-consent"

func shouldRetryExportWithUserToken(err error) bool {
	if err == nil {
		return false
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "code=99992402") || strings.Contains(msg, "99992402") {
		return true
	}
	if strings.Contains(msg, "field validation failed") {
		return true
	}
	return false
}

func ensureUserTokenForExport(ctx context.Context, state *appState) (string, error) {
	token, err := ensureUserToken(ctx, state)
	if err != nil {
		return "", err
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("user access token missing; run `" + exportTaskUserReloginHint + "`")
	}
	return token, nil
}

func createExportTaskWithFallback(ctx context.Context, state *appState, token string, tokenType tokenType, req larksdk.CreateExportTaskRequest) (ticket string, outToken string, outType tokenType, err error) {
	if state == nil || state.SDK == nil {
		return "", token, tokenType, errors.New("sdk unavailable")
	}

	ticket, err = state.SDK.CreateExportTask(ctx, token, larksdk.AccessTokenType(tokenType), req)
	if err == nil {
		return ticket, token, tokenType, nil
	}

	if tokenType != tokenTypeTenant {
		return "", token, tokenType, err
	}
	if !shouldRetryExportWithUserToken(err) {
		return "", token, tokenType, err
	}

	userToken, userErr := ensureUserTokenForExport(ctx, state)
	if userErr != nil {
		return "", token, tokenType, fmt.Errorf("%w; export requires a user token in this environment; run `%s`", err, exportTaskUserReloginHint)
	}

	ticket, err = state.SDK.CreateExportTask(ctx, userToken, larksdk.AccessTokenType(tokenTypeUser), req)
	if err != nil {
		return "", userToken, tokenTypeUser, err
	}
	return ticket, userToken, tokenTypeUser, nil
}

func pollExportTaskWithFallback(ctx context.Context, state *appState, token string, tokenType tokenType, ticket string, exportToken string) (result larksdk.ExportTaskResult, outToken string, outType tokenType, err error) {
	if state == nil || state.SDK == nil {
		return larksdk.ExportTaskResult{}, token, tokenType, errors.New("sdk unavailable")
	}

	result, err = pollExportTask(ctx, state.SDK, token, larksdk.AccessTokenType(tokenType), ticket, exportToken)
	if err == nil {
		return result, token, tokenType, nil
	}

	if tokenType != tokenTypeTenant {
		return larksdk.ExportTaskResult{}, token, tokenType, err
	}
	if !shouldRetryExportWithUserToken(err) {
		return larksdk.ExportTaskResult{}, token, tokenType, err
	}

	userToken, userErr := ensureUserTokenForExport(ctx, state)
	if userErr != nil {
		return larksdk.ExportTaskResult{}, token, tokenType, fmt.Errorf("%w; export requires a user token in this environment; run `%s`", err, exportTaskUserReloginHint)
	}

	result, err = pollExportTask(ctx, state.SDK, userToken, larksdk.AccessTokenType(tokenTypeUser), ticket, exportToken)
	if err != nil {
		return larksdk.ExportTaskResult{}, userToken, tokenTypeUser, err
	}
	return result, userToken, tokenTypeUser, nil
}

func downloadExportedFileWithFallback(ctx context.Context, state *appState, token string, tokenType tokenType, fileToken string) (rc io.ReadCloser, outToken string, outType tokenType, err error) {
	if state == nil || state.SDK == nil {
		return nil, token, tokenType, errors.New("sdk unavailable")
	}

	reader, err := state.SDK.DownloadExportedFile(ctx, token, larksdk.AccessTokenType(tokenType), fileToken)
	if err == nil {
		return reader, token, tokenType, nil
	}

	if tokenType != tokenTypeTenant {
		return nil, token, tokenType, err
	}
	if !shouldRetryExportWithUserToken(err) {
		return nil, token, tokenType, err
	}

	userToken, userErr := ensureUserTokenForExport(ctx, state)
	if userErr != nil {
		return nil, token, tokenType, fmt.Errorf("%w; export requires a user token in this environment; run `%s`", err, exportTaskUserReloginHint)
	}

	reader, err = state.SDK.DownloadExportedFile(ctx, userToken, larksdk.AccessTokenType(tokenTypeUser), fileToken)
	if err != nil {
		return nil, userToken, tokenTypeUser, err
	}
	return reader, userToken, tokenTypeUser, nil
}
