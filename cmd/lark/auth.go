package main

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Fetch and cache a tenant access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := ensureTenantToken(context.Background(), state)
			if err != nil {
				return err
			}
			payload := map[string]any{
				"tenant_access_token": token,
				"expires_at":          state.Config.TenantAccessTokenExpiresAt,
			}
			return state.Printer.Print(
				payload,
				fmt.Sprintf("tenant_access_token: %s\nexpires_at: %d", token, state.Config.TenantAccessTokenExpiresAt),
			)
		},
	}
	cmd.AddCommand(newAuthLoginCmd(state))
	cmd.AddCommand(newAuthPlatformCmd(state))
	cmd.AddCommand(newAuthUserCmd(state))
	cmd.AddCommand(newAuthExplainCmd(state))
	return cmd
}

func newAuthLoginCmd(state *appState) *cobra.Command {
	var appID string
	var appSecret string
	var baseURL string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Save app credentials to config",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Allow env/config fallback so `lark auth login` works with
			// LARK_APP_ID/LARK_APP_SECRET without forcing flags.
			if appID == "" {
				appID = state.Config.AppID
			}
			if appSecret == "" {
				appSecret = state.Config.AppSecret
			}
			if appID == "" || appSecret == "" {
				return fmt.Errorf("missing credentials: provide --app-id/--app-secret or set LARK_APP_ID/LARK_APP_SECRET")
			}

			state.Config.AppID = appID
			state.Config.AppSecret = appSecret
			if baseURL != "" {
				normalized := normalizeBaseURL(baseURL)
				state.Config.BaseURL = normalized
				state.baseURLPersist = normalized
			}
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"app_id":      state.Config.AppID,
				"base_url":    state.Config.BaseURL,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved config to %s", state.ConfigPath))
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "app ID (fallback: LARK_APP_ID)")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "app secret (fallback: LARK_APP_SECRET)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL (default: https://open.feishu.cn)")

	return cmd
}
