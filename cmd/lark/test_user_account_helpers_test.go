package main

import "lark/internal/config"

func withUserAccount(cfg *config.Config, name, accessToken, refreshToken string, expiresAt int64, scope string) {
	if cfg.UserAccounts == nil {
		cfg.UserAccounts = map[string]*config.UserAccount{}
	}
	cfg.UserAccounts[name] = &config.UserAccount{
		UserAccessToken:          accessToken,
		RefreshToken:             refreshToken,
		UserAccessTokenExpiresAt: expiresAt,
		UserAccessTokenScope:     scope,
	}
}
