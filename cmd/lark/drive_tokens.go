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
	if cachedUserTokenValid(state.Config, time.Now()) {
		return state.Config.UserAccessToken, nil
	}
	if state.Config.UserRefreshToken() == "" {
		return "", errors.New(driveSearchUserTokenHint)
	}
	token, err := ensureUserToken(ctx, state)
	if err != nil || token == "" {
		return "", err
	}
	return token, nil
}
