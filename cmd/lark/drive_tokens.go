package main

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"
)

const driveSearchUserTokenHint = "drive search requires a user access token; set LARK_USER_ACCESS_TOKEN or run `lark auth user login --scopes \"offline_access drive:drive:readonly\" --force-consent`"

func resolveDriveSearchToken(ctx context.Context, state *appState) (string, error) {
	if state == nil || state.Config == nil {
		return "", errors.New("config is required")
	}
	if token := strings.TrimSpace(os.Getenv("LARK_USER_ACCESS_TOKEN")); token != "" {
		return token, nil
	}
	account := resolveUserAccountName(state)
	stored, ok, err := loadUserToken(state, account)
	if err != nil {
		return "", err
	}
	if ok && cachedUserTokenValid(stored, time.Now()) {
		return stored.AccessToken, nil
	}
	acct, _ := loadUserAccount(state.Config, account)
	refreshToken := stored.RefreshToken
	if refreshToken == "" {
		refreshToken = acct.RefreshTokenValue()
	}
	if refreshToken == "" {
		return "", errors.New(driveSearchUserTokenHint)
	}
	token, err := ensureUserToken(ctx, state)
	if err != nil || token == "" {
		return "", err
	}
	return token, nil
}
