package main

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newAuthCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Authentication and tokens",
		RunE: func(cmd *cobra.Command, args []string) error {
			return cmd.Help()
		},
	}
	cmd.AddCommand(newAuthTenantCmd(state))
	cmd.AddCommand(newAuthLoginCmd(state))
	cmd.AddCommand(newAuthPlatformCmd(state))
	cmd.AddCommand(newAuthUserCmd(state))
	cmd.AddCommand(newAuthExplainCmd(state))
	return cmd
}

func newAuthTenantCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tenant",
		Short: "Fetch and cache a tenant access token",
		RunE: func(cmd *cobra.Command, args []string) error {
			token, err := ensureTenantToken(cmd.Context(), state)
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
	return cmd
}

func newAuthLoginCmd(state *appState) *cobra.Command {
	var appID string
	var appSecret string
	var baseURL string
	var storeSecretInKeyring bool
	var storeSecretInConfig bool

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

			if baseURL != "" {
				normalized := normalizeBaseURL(baseURL)
				state.Config.BaseURL = normalized
				state.baseURLPersist = normalized
			}
			state.Config.AppID = appID
			storeInKeyring, err := resolveAppSecretStorage(state, storeSecretInKeyring, storeSecretInConfig)
			if err != nil {
				return err
			}
			if err := persistAppSecret(state, appSecret, storeInKeyring); err != nil {
				return err
			}
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path":           state.ConfigPath,
				"app_id":                state.Config.AppID,
				"base_url":              state.Config.BaseURL,
				"app_secret_in_keyring": storeInKeyring,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved config to %s", state.ConfigPath))
		},
	}

	cmd.Flags().StringVar(&appID, "app-id", "", "app ID (fallback: LARK_APP_ID)")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "app secret (fallback: LARK_APP_SECRET)")
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL (default: https://open.feishu.cn)")
	cmd.Flags().BoolVar(&storeSecretInKeyring, "store-secret-in-keyring", false, "store app secret in keychain instead of config")
	cmd.Flags().BoolVar(&storeSecretInConfig, "store-secret-in-config", false, "store app secret in config (disables keychain storage)")

	return cmd
}
