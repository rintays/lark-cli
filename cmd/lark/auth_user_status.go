package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

type authUserStatusPayload struct {
	ConfigPath                      string `json:"config_path"`
	UserAccessTokenPresent          bool   `json:"user_access_token_present"`
	RefreshTokenPresent             bool   `json:"refresh_token_present"`
	UserAccessTokenExpiresAt        int64  `json:"user_access_token_expires_at"`
	UserAccessTokenExpiresAtRFC3339 string `json:"user_access_token_expires_at_rfc3339,omitempty"`
	UserAccessTokenScope            string `json:"user_access_token_scope,omitempty"`
	Remediation                     string `json:"remediation,omitempty"`
}

func newAuthUserStatusCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show stored user OAuth credential status",
		RunE: func(cmd *cobra.Command, args []string) error {
			scope := strings.TrimSpace(state.Config.UserAccessTokenScope)
			payload := authUserStatusPayload{
				ConfigPath:               state.ConfigPath,
				UserAccessTokenPresent:   state.Config.UserAccessToken != "",
				RefreshTokenPresent:      state.Config.RefreshToken != "",
				UserAccessTokenExpiresAt: state.Config.UserAccessTokenExpiresAt,
				UserAccessTokenScope:     scope,
			}
			if payload.UserAccessTokenExpiresAt != 0 {
				payload.UserAccessTokenExpiresAtRFC3339 = time.Unix(payload.UserAccessTokenExpiresAt, 0).UTC().Format(time.RFC3339)
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
			if payload.Remediation != "" {
				text += fmt.Sprintf("\n\nMissing refresh_token. Re-run: `%s`", payload.Remediation)
			}
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
