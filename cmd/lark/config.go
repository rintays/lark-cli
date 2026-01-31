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

	cmd := &cobra.Command{
		Use:   "set",
		Short: "Persist configuration values",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			normalized := normalizeBaseURL(baseURL)
			state.Config.BaseURL = normalized
			state.baseURLPersist = normalized
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"base_url":    normalized,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved base_url to %s", state.ConfigPath))
		},
	}
	cmd.Flags().StringVar(&baseURL, "base-url", "", "base URL to persist")
	_ = cmd.MarkFlagRequired("base-url")
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
