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
	cmd.AddCommand(newConfigGetCmd(state))
	cmd.AddCommand(newConfigSetCmd(state))
	return cmd
}

func newConfigGetCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
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

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Persist configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}

			useBaseURL := cmd.Flags().Changed("base-url")
			usePlatform := cmd.Flags().Changed("platform")
			if !useBaseURL && !usePlatform {
				return errors.New("either --base-url or --platform is required")
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
	cmd.MarkFlagsMutuallyExclusive("base-url", "platform")
	cmd.MarkFlagsOneRequired("base-url", "platform")

	return cmd
}

func formatConfigHuman(cfg *config.Config) string {
	lines := []string{
		fmt.Sprintf("app_id: %s", cfg.AppID),
		fmt.Sprintf("base_url: %s", cfg.BaseURL),
		fmt.Sprintf("default_mailbox_id: %s", cfg.DefaultMailboxID),
		fmt.Sprintf("tenant_access_token_expires_at: %d", cfg.TenantAccessTokenExpiresAt),
		fmt.Sprintf("user_access_token_expires_at: %d", cfg.UserAccessTokenExpiresAt),
	}
	return strings.Join(lines, "\n")
}
