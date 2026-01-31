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
