package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type authUserStatusPayload struct {
	ConfigPath                      string   `json:"config_path"`
	UserAccessTokenPresent          bool     `json:"user_access_token_present"`
	RefreshTokenPresent             bool     `json:"refresh_token_present"`
	UserAccessTokenExpiresAt        int64    `json:"user_access_token_expires_at"`
	UserAccessTokenExpiresAtRFC3339 string   `json:"user_access_token_expires_at_rfc3339,omitempty"`
	UserAccessTokenScope            string   `json:"user_access_token_scope,omitempty"`
	RefreshTokenServices            []string `json:"refresh_token_services,omitempty"`
	RefreshTokenScopes              string   `json:"refresh_token_scopes,omitempty"`
	RefreshTokenCreatedAt           int64    `json:"refresh_token_created_at,omitempty"`
	RefreshTokenCreatedAtRFC3339    string   `json:"refresh_token_created_at_rfc3339,omitempty"`
	Remediation                     string   `json:"remediation,omitempty"`
}

func newAuthUserStatusCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show stored user OAuth credential status",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := strings.TrimSpace(state.Config.UserAccessTokenScope)
			refreshToken := state.Config.UserRefreshToken()
			payload := authUserStatusPayload{
				ConfigPath:               state.ConfigPath,
				UserAccessTokenPresent:   state.Config.UserAccessToken != "",
				RefreshTokenPresent:      refreshToken != "",
				UserAccessTokenExpiresAt: state.Config.UserAccessTokenExpiresAt,
				UserAccessTokenScope:     scope,
			}
			if payload.UserAccessTokenExpiresAt != 0 {
				payload.UserAccessTokenExpiresAtRFC3339 = time.Unix(payload.UserAccessTokenExpiresAt, 0).UTC().Format(time.RFC3339)
			}
			if state.Config.UserRefreshTokenPayload != nil {
				payload.RefreshTokenServices = state.Config.UserRefreshTokenPayload.Services
				payload.RefreshTokenScopes = strings.TrimSpace(state.Config.UserRefreshTokenPayload.Scopes)
				payload.RefreshTokenCreatedAt = state.Config.UserRefreshTokenPayload.CreatedAt
				if payload.RefreshTokenCreatedAt != 0 {
					payload.RefreshTokenCreatedAtRFC3339 = time.Unix(payload.RefreshTokenCreatedAt, 0).UTC().Format(time.RFC3339)
				}
			}
			if !payload.RefreshTokenPresent {
				payload.Remediation = userOAuthReloginCommand
			}

			text := fmt.Sprintf(
				"config_path: %s\nuser_access_token_present: %t\nrefresh_token_present: %t\nuser_access_token_expires_at: %d",
				payload.ConfigPath,
				payload.UserAccessTokenPresent,
				payload.RefreshTokenPresent,
				payload.UserAccessTokenExpiresAt,
			)
			if payload.UserAccessTokenExpiresAtRFC3339 != "" {
				text += fmt.Sprintf("\nuser_access_token_expires_at_rfc3339: %s", payload.UserAccessTokenExpiresAtRFC3339)
			}
			if payload.UserAccessTokenScope != "" {
				text += fmt.Sprintf("\nuser_access_token_scope: %s", payload.UserAccessTokenScope)
			}
			if payload.RefreshTokenCreatedAt != 0 {
				text += fmt.Sprintf("\nrefresh_token_created_at: %d", payload.RefreshTokenCreatedAt)
			}
			if payload.RefreshTokenCreatedAtRFC3339 != "" {
				text += fmt.Sprintf("\nrefresh_token_created_at_rfc3339: %s", payload.RefreshTokenCreatedAtRFC3339)
			}
			if payload.RefreshTokenScopes != "" {
				text += fmt.Sprintf("\nrefresh_token_scopes: %s", payload.RefreshTokenScopes)
			}
			if len(payload.RefreshTokenServices) > 0 {
				text += fmt.Sprintf("\nrefresh_token_services: %s", strings.Join(payload.RefreshTokenServices, " "))
			}
			if payload.Remediation != "" {
				text += fmt.Sprintf("\n\nMissing refresh_token. Re-run: `%s`", payload.Remediation)
			}
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
