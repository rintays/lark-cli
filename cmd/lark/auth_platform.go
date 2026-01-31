package main

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newAuthPlatformCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "platform",
		Short: "Manage the default platform base URL",
	}
	cmd.AddCommand(newAuthPlatformSetCmd(state))
	cmd.AddCommand(newAuthPlatformInfoCmd(state))
	return cmd
}

func newAuthPlatformSetCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set [feishu|lark]",
		Short: "Persist platform base URL to config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			platform := strings.ToLower(strings.TrimSpace(args[0]))
			baseURL, err := platformBaseURL(platform)
			if err != nil {
				return err
			}
			normalized := normalizeBaseURL(baseURL)
			state.Config.BaseURL = normalized
			state.baseURLPersist = normalized
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"platform":    platform,
				"base_url":    normalized,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved platform %s to %s", platform, state.ConfigPath))
		},
	}
	return cmd
}

func newAuthPlatformInfoCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info",
		Short: "Show configured platform and base URL",
		RunE: func(cmd *cobra.Command, args []string) error {
			baseURL := state.baseURLPersist
			if baseURL == "" {
				baseURL = state.Config.BaseURL
			}
			normalized := normalizeBaseURL(baseURL)
			platform := platformFromBaseURL(normalized)
			payload := map[string]any{
				"platform": platform,
				"base_url": normalized,
			}
			return state.Printer.Print(payload, fmt.Sprintf("platform: %s\nbase_url: %s", platform, normalized))
		},
	}
	return cmd
}
