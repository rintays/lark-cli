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
			account := resolveUserAccountName(state)
			scopes, source, err := resolveUserOAuthScopes(state, userOAuthScopeOptions{})
			if err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"account":     account,
				"scopes":      scopes,
				"source":      source,
			}
			text := "scopes: " + strings.Join(scopes, " ")
			if source != "" {
				text += fmt.Sprintf(" (source: %s)", source)
			}
			if account != "" {
				text += fmt.Sprintf("\naccount: %s", account)
			}
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newAuthUserScopesSetCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "set <scopes...>",
		Short: "Set default user OAuth scopes",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if strings.TrimSpace(scopes) == "" {
					return errors.New("scopes is required")
				}
				return nil
			}
			if scopes != "" {
				return errors.New("scopes provided twice")
			}
			return cmd.Flags().Set("scopes", strings.Join(args, " "))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			account := resolveUserAccountName(state)
			list := normalizeScopes(parseScopeList(scopes))
			if len(list) == 0 {
				return errors.New("scopes must not be empty")
			}
			acct := ensureUserAccount(state.Config, account)
			acct.UserScopes = ensureOfflineAccess(list)
			saveUserAccount(state.Config, account, acct)
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"account":     account,
				"scopes":      acct.UserScopes,
			}
			return state.Printer.Print(payload, "saved user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated, or provide as positional arguments)")
	return cmd
}

func newAuthUserScopesAddCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "add <scopes...>",
		Short: "Add default user OAuth scopes",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if strings.TrimSpace(scopes) == "" {
					return errors.New("scopes is required")
				}
				return nil
			}
			if scopes != "" {
				return errors.New("scopes provided twice")
			}
			return cmd.Flags().Set("scopes", strings.Join(args, " "))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			account := resolveUserAccountName(state)
			added := normalizeScopes(parseScopeList(scopes))
			if len(added) == 0 {
				return errors.New("scopes must not be empty")
			}
			acct := ensureUserAccount(state.Config, account)
			current := normalizeScopes(acct.UserScopes)
			if len(current) == 0 {
				current = []string{defaultUserOAuthScope}
			}
			merged := append(current, added...)
			acct.UserScopes = ensureOfflineAccess(normalizeScopes(merged))
			saveUserAccount(state.Config, account, acct)
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"account":     account,
				"scopes":      acct.UserScopes,
			}
			return state.Printer.Print(payload, "added user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated, or provide as positional arguments)")
	return cmd
}

func newAuthUserScopesRemoveCmd(state *appState) *cobra.Command {
	var scopes string
	cmd := &cobra.Command{
		Use:   "remove <scopes...>",
		Short: "Remove default user OAuth scopes",
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) == 0 {
				if strings.TrimSpace(scopes) == "" {
					return errors.New("scopes is required")
				}
				return nil
			}
			if scopes != "" {
				return errors.New("scopes provided twice")
			}
			return cmd.Flags().Set("scopes", strings.Join(args, " "))
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			account := resolveUserAccountName(state)
			remove := normalizeScopes(parseScopeList(scopes))
			if len(remove) == 0 {
				return errors.New("scopes must not be empty")
			}
			acct := ensureUserAccount(state.Config, account)
			current := normalizeScopes(acct.UserScopes)
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
			acct.UserScopes = ensureOfflineAccess(normalizeScopes(remaining))
			saveUserAccount(state.Config, account, acct)
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"account":     account,
				"scopes":      acct.UserScopes,
			}
			return state.Printer.Print(payload, "removed user scopes")
		},
	}
	cmd.Flags().StringVar(&scopes, "scopes", "", "OAuth scopes (space/comma-separated, or provide as positional arguments)")
	return cmd
}

func newAuthUserServicesCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "List built-in OAuth service profiles",
		RunE: func(cmd *cobra.Command, args []string) error {
			services := authregistry.ListUserOAuthServices()
			serviceScopes := make(map[string]authregistry.ServiceScopeSet, len(services))
			requiredScopes := make(map[string][]string, len(services))
			for _, svc := range services {
				def := authregistry.Registry[svc]
				serviceScopes[svc] = def.UserScopes
				requiredScopes[svc] = def.RequiredUserScopes
			}
			payload := map[string]any{
				"services":                     services,
				"default_services":             authregistry.DefaultUserOAuthServices,
				"service_aliases":              authregistry.UserOAuthServiceAliases,
				"service_scopes":               serviceScopes,
				"service_required_user_scopes": requiredScopes,
				"drive_scope_values":           []string{"full", "readonly"},
			}
			lines := make([]string, 0, len(services))
			for _, svc := range services {
				set := serviceScopes[svc]
				full := strings.Join(set.Full, " ")
				ro := strings.Join(set.Readonly, " ")
				required := strings.Join(requiredScopes[svc], " ")
				lines = append(lines, fmt.Sprintf("%s\t%s\t%s\t%s", svc, full, ro, required))
			}
			text := tableText([]string{"service", "full", "readonly", "required"}, lines, "no services configured")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}
