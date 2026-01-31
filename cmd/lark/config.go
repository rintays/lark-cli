package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"lark/internal/config"
)

func newConfigCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "Manage CLI configuration",
	}
	cmd.AddCommand(newConfigInfoCmd(state))
	cmd.AddCommand(newConfigSetCmd(state))
	cmd.AddCommand(newConfigUnsetCmd(state))
	cmd.AddCommand(newConfigListKeysCmd(state))
	return cmd
}

func newConfigInfoCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show the loaded configuration",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			payload := state.Config
			text := formatConfigHuman(state.Config)
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newConfigSetCmd(state *appState) *cobra.Command {
	var baseURL string
	var platform string
	var defaultMailboxID string
	var defaultTokenType string
	var appID string
	var appSecret string

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Persist configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}

			useBaseURL := cmd.Flags().Changed("base-url")
			usePlatform := cmd.Flags().Changed("platform")
			useDefaultMailboxID := cmd.Flags().Changed("default-mailbox-id")
			useDefaultTokenType := cmd.Flags().Changed("default-token-type")
			useAppID := cmd.Flags().Changed("app-id")
			useAppSecret := cmd.Flags().Changed("app-secret")

			usedBaseURLGroup := useBaseURL || usePlatform
			usedMailboxGroup := useDefaultMailboxID
			usedTokenTypeGroup := useDefaultTokenType
			usedAppCredsGroup := useAppID || useAppSecret

			groupsUsed := 0
			if usedBaseURLGroup {
				groupsUsed++
			}
			if usedMailboxGroup {
				groupsUsed++
			}
			if usedTokenTypeGroup {
				groupsUsed++
			}
			if usedAppCredsGroup {
				groupsUsed++
			}
			if groupsUsed == 0 {
				return errors.New("one of --base-url, --platform, --default-mailbox-id, --default-token-type, --app-id, or --app-secret is required")
			}
			if groupsUsed > 1 {
				return errors.New("flags are mutually exclusive; choose one of: (--base-url|--platform), --default-mailbox-id, --default-token-type, or (--app-id/--app-secret)")
			}

			if usedAppCredsGroup {
				payload := map[string]any{
					"config_path": state.ConfigPath,
				}

				if useAppID {
					appID = strings.TrimSpace(appID)
					if appID == "" {
						return errors.New("app-id must not be empty")
					}
					state.Config.AppID = appID
					payload["app_id"] = appID
				}
				if useAppSecret {
					appSecret = strings.TrimSpace(appSecret)
					if appSecret == "" {
						return errors.New("app-secret must not be empty")
					}
					state.Config.AppSecret = appSecret
					payload["app_secret_set"] = true
				}

				if err := state.saveConfig(); err != nil {
					return err
				}
				return state.Printer.Print(payload, fmt.Sprintf("saved app credentials to %s", state.ConfigPath))
			}

			if useDefaultMailboxID {
				defaultMailboxID = strings.TrimSpace(defaultMailboxID)
				if defaultMailboxID == "" {
					return errors.New("default-mailbox-id must not be empty")
				}
				state.Config.DefaultMailboxID = defaultMailboxID
				if err := state.saveConfig(); err != nil {
					return err
				}
				payload := map[string]any{
					"config_path":        state.ConfigPath,
					"default_mailbox_id": defaultMailboxID,
				}
				return state.Printer.Print(payload, fmt.Sprintf("saved default_mailbox_id to %s", state.ConfigPath))
			}
			if useDefaultTokenType {
				defaultTokenType = strings.ToLower(strings.TrimSpace(defaultTokenType))
				if defaultTokenType == "" {
					return errors.New("default-token-type must not be empty")
				}
				if defaultTokenType != "tenant" && defaultTokenType != "user" {
					return errors.New("default-token-type must be tenant or user")
				}
				state.Config.DefaultTokenType = defaultTokenType
				if err := state.saveConfig(); err != nil {
					return err
				}
				payload := map[string]any{
					"config_path":        state.ConfigPath,
					"default_token_type": defaultTokenType,
				}
				return state.Printer.Print(payload, fmt.Sprintf("saved default_token_type to %s", state.ConfigPath))
			}

			var normalized string
			if usePlatform {
				platform = strings.ToLower(strings.TrimSpace(platform))
				if platform == "" {
					return errors.New("platform must not be empty")
				}
				mapped, err := platformBaseURL(platform)
				if err != nil {
					return err
				}
				normalized = normalizeBaseURL(mapped)
			} else {
				baseURL = strings.TrimSpace(baseURL)
				if baseURL == "" {
					return errors.New("base-url must not be empty")
				}
				normalized = normalizeBaseURL(baseURL)
			}

			state.Config.BaseURL = normalized
			state.baseURLPersist = normalized
			if err := state.saveConfig(); err != nil {
				return err
			}

			payload := map[string]any{
				"config_path": state.ConfigPath,
				"base_url":    normalized,
			}
			if usePlatform {
				payload["platform"] = platform
				return state.Printer.Print(payload, fmt.Sprintf("saved platform %s to %s", platform, state.ConfigPath))
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved base_url to %s", state.ConfigPath))
		},
	}

	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL to persist")
	cmd.Flags().StringVar(&platform, "platform", "", "platform to persist (feishu or lark)")
	cmd.Flags().StringVar(&defaultMailboxID, "default-mailbox-id", "", "default mailbox id to persist (or 'me')")
	cmd.Flags().StringVar(&defaultTokenType, "default-token-type", "", "default token type to persist (tenant or user)")
	cmd.Flags().StringVar(&appID, "app-id", "", "app ID to persist")
	cmd.Flags().StringVar(&appSecret, "app-secret", "", "app secret to persist (stored in plain text)")
	cmd.MarkFlagsMutuallyExclusive("base-url", "platform", "default-mailbox-id", "default-token-type")
	cmd.MarkFlagsOneRequired("base-url", "platform", "default-mailbox-id", "default-token-type", "app-id", "app-secret")

	return cmd
}

func newConfigUnsetCmd(state *appState) *cobra.Command {
	var unsetBaseURL bool
	var unsetDefaultMailboxID bool
	var unsetDefaultTokenType bool
	var unsetUserTokens bool

	cmd := &cobra.Command{
		Use:   "unset",
		Short: "Clear persisted configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}

			useBaseURL := cmd.Flags().Changed("base-url")
			useDefaultMailboxID := cmd.Flags().Changed("default-mailbox-id")
			useDefaultTokenType := cmd.Flags().Changed("default-token-type")
			useUserTokens := cmd.Flags().Changed("user-tokens")
			if !useBaseURL && !useDefaultMailboxID && !useDefaultTokenType && !useUserTokens {
				return errors.New("one of --base-url, --default-mailbox-id, --default-token-type, or --user-tokens is required")
			}

			if useBaseURL {
				if !unsetBaseURL {
					return errors.New("--base-url must be true")
				}
				state.Config.BaseURL = ""
				state.baseURLPersist = ""
				if err := state.saveConfig(); err != nil {
					return err
				}
				payload := map[string]any{
					"config_path": state.ConfigPath,
					"base_url":    "",
				}
				return state.Printer.Print(payload, fmt.Sprintf("cleared base_url in %s", state.ConfigPath))
			}

			if useDefaultMailboxID {
				if !unsetDefaultMailboxID {
					return errors.New("--default-mailbox-id must be true")
				}
				state.Config.DefaultMailboxID = ""
				if err := state.saveConfig(); err != nil {
					return err
				}
				payload := map[string]any{
					"config_path":        state.ConfigPath,
					"default_mailbox_id": "",
				}
				return state.Printer.Print(payload, fmt.Sprintf("cleared default_mailbox_id in %s", state.ConfigPath))
			}

			if useDefaultTokenType {
				if !unsetDefaultTokenType {
					return errors.New("--default-token-type must be true")
				}
				state.Config.DefaultTokenType = ""
				if err := state.saveConfig(); err != nil {
					return err
				}
				payload := map[string]any{
					"config_path":        state.ConfigPath,
					"default_token_type": "",
				}
				return state.Printer.Print(payload, fmt.Sprintf("cleared default_token_type in %s", state.ConfigPath))
			}

			if !unsetUserTokens {
				return errors.New("--user-tokens must be true")
			}
			state.Config.UserAccessToken = ""
			state.Config.RefreshToken = ""
			state.Config.UserAccessTokenExpiresAt = 0
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path":         state.ConfigPath,
				"user_tokens_cleared": true,
			}
			return state.Printer.Print(payload, fmt.Sprintf("cleared user tokens in %s", state.ConfigPath))
		},
	}
	cmd.Flags().BoolVar(&unsetBaseURL, "base-url", false, "clear the persisted base URL")
	cmd.Flags().BoolVar(&unsetDefaultMailboxID, "default-mailbox-id", false, "clear the persisted default mailbox id")
	cmd.Flags().BoolVar(&unsetDefaultTokenType, "default-token-type", false, "clear the persisted default token type")
	cmd.Flags().BoolVar(&unsetUserTokens, "user-tokens", false, "clear persisted user access tokens")
	cmd.MarkFlagsMutuallyExclusive("base-url", "default-mailbox-id", "default-token-type", "user-tokens")
	cmd.MarkFlagsOneRequired("base-url", "default-mailbox-id", "default-token-type", "user-tokens")

	return cmd
}

type configKeyInfo struct {
	Key         string `json:"key"`
	Description string `json:"description"`
}

func configKeys() []configKeyInfo {
	return []configKeyInfo{
		{
			Key:         "base-url",
			Description: "Base URL to persist (config set) or clear (config unset)",
		},
		{
			Key:         "platform",
			Description: "Platform shortcut to set the base URL (feishu or lark)",
		},
		{
			Key:         "default-mailbox-id",
			Description: "Default mailbox ID to persist (config set) or clear (config unset)",
		},
		{
			Key:         "user-tokens",
			Description: "Clear persisted user access/refresh tokens (config unset)",
		},
		{
			Key:         "keyring-backend",
			Description: "OAuth token storage backend (file|keychain|auto) (currently only file is implemented)",
		},
		{
			Key:         "app-id",
			Description: "App ID to persist (config set)",
		},
		{
			Key:         "app-secret",
			Description: "App secret to persist (config set)",
		},
	}
}

func newConfigListKeysCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list-keys",
		Short: "List supported config keys",
		RunE: func(cmd *cobra.Command, args []string) error {
			keys := configKeys()
			text := formatConfigKeysHuman(keys)
			return state.Printer.Print(keys, text)
		},
	}
	return cmd
}

func formatConfigKeysHuman(keys []configKeyInfo) string {
	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		if key.Description == "" {
			lines = append(lines, key.Key)
			continue
		}
		lines = append(lines, fmt.Sprintf("%s - %s", key.Key, key.Description))
	}
	return strings.Join(lines, "\n")
}

func formatConfigHuman(cfg *config.Config) string {
	lines := []string{
		fmt.Sprintf("app_id: %s", cfg.AppID),
		fmt.Sprintf("base_url: %s", cfg.BaseURL),
		fmt.Sprintf("default_mailbox_id: %s", cfg.DefaultMailboxID),
		fmt.Sprintf("default_token_type: %s", cfg.DefaultTokenType),
		fmt.Sprintf("keyring_backend: %s", cfg.KeyringBackend),
		fmt.Sprintf("user_scopes: %s", strings.Join(cfg.UserScopes, " ")),
		fmt.Sprintf("tenant_access_token_expires_at: %d", cfg.TenantAccessTokenExpiresAt),
		fmt.Sprintf("user_access_token_expires_at: %d", cfg.UserAccessTokenExpiresAt),
	}
	return strings.Join(lines, "\n")
}
