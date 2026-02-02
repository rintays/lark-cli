package main

import (
	"errors"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

type authUserAccountStatus struct {
	Account              string `json:"account"`
	Default              bool   `json:"default"`
	AccessTokenPresent   bool   `json:"access_token_present"`
	RefreshTokenPresent  bool   `json:"refresh_token_present"`
	AccessTokenExpiresAt int64  `json:"access_token_expires_at"`
}

func newAuthUserAccountsCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage stored user OAuth accounts",
	}
	cmd.AddCommand(newAuthUserAccountsListCmd(state))
	cmd.AddCommand(newAuthUserAccountsSetCmd(state))
	cmd.AddCommand(newAuthUserAccountsRemoveCmd(state))
	return cmd
}

func newAuthUserAccountsListCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List stored user OAuth accounts",
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			defaultAccount := strings.TrimSpace(state.Config.DefaultUserAccount)
			if defaultAccount == "" {
				defaultAccount = defaultUserAccountName
			}
			names := listUserAccountNames(state.Config)
			statuses := make([]authUserAccountStatus, 0, len(names))
			lines := make([]string, 0, len(names))
			for _, name := range names {
				stored, ok, err := loadUserToken(state, name)
				if err != nil {
					return err
				}
				status := authUserAccountStatus{
					Account:             name,
					Default:             name == defaultAccount,
					AccessTokenPresent:  ok && stored.AccessToken != "",
					RefreshTokenPresent: ok && stored.RefreshToken != "",
				}
				if ok {
					status.AccessTokenExpiresAt = stored.ExpiresAt
				}
				statuses = append(statuses, status)
				lines = append(lines, fmt.Sprintf("%s\t%t\t%t\t%t\t%d", status.Account, status.Default, status.AccessTokenPresent, status.RefreshTokenPresent, status.AccessTokenExpiresAt))
			}
			payload := map[string]any{
				"config_path":     state.ConfigPath,
				"default_account": defaultAccount,
				"accounts":        statuses,
			}
			text := tableText([]string{"account", "default", "access_token", "refresh_token", "expires_at"}, lines, "no accounts found")
			return state.Printer.Print(payload, text)
		},
	}
	return cmd
}

func newAuthUserAccountsSetCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set <account>",
		Short: "Set the default user OAuth account",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if normalizeAccountName(args[0]) == "" {
				return argsUsageError(cmd, errors.New("account must not be empty"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			account := normalizeAccountName(args[0])
			state.Config.DefaultUserAccount = account
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path":     state.ConfigPath,
				"default_account": account,
			}
			return state.Printer.Print(payload, fmt.Sprintf("saved default user account to %s", state.ConfigPath))
		},
	}
	return cmd
}

func newAuthUserAccountsRemoveCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "remove <account>",
		Short: "Remove stored user OAuth account",
		Args: func(cmd *cobra.Command, args []string) error {
			if err := cobra.ExactArgs(1)(cmd, args); err != nil {
				return argsUsageError(cmd, err)
			}
			if normalizeAccountName(args[0]) == "" {
				return argsUsageError(cmd, errors.New("account must not be empty"))
			}
			return nil
		},
		RunE: func(cmd *cobra.Command, args []string) error {
			if state.Config == nil {
				return errors.New("config is required")
			}
			account := normalizeAccountName(args[0])
			if err := confirmDestructive(cmd, state, fmt.Sprintf("remove account %s", account)); err != nil {
				return err
			}
			if userTokenBackend(state.Config) == "keychain" {
				if err := deleteKeyringToken(state, account); err != nil {
					return err
				}
			}
			deleteUserAccount(state.Config, account)
			if state.Config.DefaultUserAccount == account {
				state.Config.DefaultUserAccount = defaultUserAccountName
			}
			if err := state.saveConfig(); err != nil {
				return err
			}
			payload := map[string]any{
				"config_path": state.ConfigPath,
				"account":     account,
				"removed":     true,
			}
			return state.Printer.Print(payload, fmt.Sprintf("removed user account %s", account))
		},
	}
	return cmd
}
