package main

import (
	"errors"
	"fmt"
	"strings"

	"lark/internal/authregistry"

	"github.com/spf13/cobra"
)

func newAuthUserScopesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scopes",
		Short: "Manage default user OAuth scopes",
	}
	cmd.AddCommand(newAuthUserScopesListCmd(state))
	cmd.AddCommand(newAuthUserScopesSetCmd(state))
	cmd.AddCommand(newAuthUserScopesAddCmd(state))
	cmd.AddCommand(newAuthUserScopesRemoveCmd(state))
	return cmd
}

func newAuthUserScopesListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List default user OAuth scopes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			scopes, source, err := resolveUserOAuthScopes(state, userOAuthScopeOptions{})
			if err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"scopes":      scopes,
				"source":      source,
			}
			text := "scopes: " + strings.Join(scopes, " ")
			if source != "" {
				text += fmt.Sprintf(" (source: %s)", source)
			}
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newAuthUserScopesSetCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "set",
		Short: "Set default user OAuth scopes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			list := normalizeScopes(parseScopeList(scopes))
			if len(list) == 0 {
				return errors.New("scopes must not be empty")
			}
			state.Config.UserScopes = ensureOfflineAccess(list)
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"scopes":      state.Config.UserScopes,
			}
			return state.Printer.Print(payload, "saved user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated)")
	_ = cmd.MarkFlagRequired("scopes")
	return cmd
}

func newAuthUserScopesAddCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "add",
		Short: "Add default user OAuth scopes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			added := normalizeScopes(parseScopeList(scopes))
			if len(added) == 0 {
				return errors.New("scopes must not be empty")
			}
			current := normalizeScopes(state.Config.UserScopes)
			if len(current) == 0 {
				current = []string{defaultUserOAuthScope}
			}
			merged := append(current, added...)
			state.Config.UserScopes = ensureOfflineAccess(normalizeScopes(merged))
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"scopes":      state.Config.UserScopes,
			}
			return state.Printer.Print(payload, "added user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated)")
	_ = cmd.MarkFlagRequired("scopes")
	return cmd
}

func newAuthUserScopesRemoveCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "remove",
		Short: "Remove default user OAuth scopes",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			remove := normalizeScopes(parseScopeList(scopes))
			if len(remove) == 0 {
				return errors.New("scopes must not be empty")
			}
			current := normalizeScopes(state.Config.UserScopes)
			if len(current) == 0 {
				current = []string{defaultUserOAuthScope}
			}
			remaining := make([]string, 0, len(current))
			removeSet := make(map[string]struct{}, len(remove))
			for _, scope := range remove {
				removeSet[scope] = struct{}{}
			}
			for _, scope := range current {
				if _, ok := removeSet[scope]; ok {
					continue
				}
				remaining = append(remaining, scope)
			}
			state.Config.UserScopes = ensureOfflineAccess(normalizeScopes(remaining))
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"scopes":      state.Config.UserScopes,
			}
			return state.Printer.Print(payload, "removed user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated)")
	_ = cmd.MarkFlagRequired("scopes")
	return cmd
}

func newAuthUserServicesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "List built-in OAuth service profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			services := authregistry.ListUserOAuthServices()
			serviceScopes := make(map[string]authregistry.ServiceScopeSet, len(services))
			for _, svc := range services {
				serviceScopes[svc] = authregistry.Registry[svc].UserScopes
			}
			payload := map[string]any{
				"services":           services,
				"default_services":   authregistry.DefaultUserOAuthServices,
				"service_aliases":    authregistry.UserOAuthServiceAliases,
				"service_scopes":     serviceScopes,
				"drive_scope_values": []string{"full", "readonly"},
			}
			lines := make([]string, 0, len(services))
			for _, svc := range services {
				set := serviceScopes[svc]
				full := strings.Join(set.Full, " ")
				ro := strings.Join(set.Readonly, " ")
				lines = append(lines, fmt.Sprintf("%s\tfull=%s\treadonly=%s", svc, full, ro))
			}
			text := tableText([]string{"service", "full", "readonly"}, lines, "no services configured")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
