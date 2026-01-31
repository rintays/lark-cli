package main

import (
	"fmt"
	"strings"

	"lark/internal/authregistry"

	"github.com/spf13/cobra"
)

func newAuthExplainCmd(state *appState) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "explain <command...>",
		Short: "Explain auth requirements for a command",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			command := strings.Join(args, " ")

			services, tokenTypes, requiresOffline, _, ok, err := authregistry.RequirementsForCommand(command)
			if err != nil {
				return err
			}
			if !ok {
				return fmt.Errorf("no auth registry mapping found for command %q", command)
			}

			requiredUserScopes, missingScopeDecls, err := authregistry.RequiredUserScopesFromServicesReport(services)
			if err != nil {
				return err
			}

			requiresUser := false
			for _, tt := range tokenTypes {
				if tt == authregistry.TokenUser {
					requiresUser = true
					break
				}
			}

			var suggestedScopes []string
			suggestedCmd := ""
			if requiresUser {
				suggestedScopes = requiredUserScopes
				if requiresOffline {
					suggestedScopes = ensureOfflineAccess(suggestedScopes)
				}
				scopeArg := strings.Join(suggestedScopes, " ")
				if scopeArg != "" {
					suggestedCmd = fmt.Sprintf("lark auth user login --scopes %q", scopeArg)
				}
			}

			payload := map[string]any{
				"command":                               command,
				"services":                              services,
				"token_types":                           tokenTypes,
				"requires_offline":                      requiresOffline,
				"required_user_scopes":                  requiredUserScopes,
				"services_missing_required_user_scopes": missingScopeDecls,
				"suggested_user_login_scopes":           suggestedScopes,
				"suggested_user_login_command":          suggestedCmd,
			}

			tt := make([]string, 0, len(tokenTypes))
			for _, t := range tokenTypes {
				tt = append(tt, string(t))
			}

			lines := []string{
				fmt.Sprintf("command: %s", command),
				fmt.Sprintf("services: %s", strings.Join(services, ", ")),
				fmt.Sprintf("token_types: %s", strings.Join(tt, ", ")),
				fmt.Sprintf("requires_offline: %t", requiresOffline),
			}

			if len(requiredUserScopes) == 0 {
				lines = append(lines, "required_user_scopes: (none)")
			} else {
				lines = append(lines, fmt.Sprintf("required_user_scopes: %s", strings.Join(requiredUserScopes, " ")))
			}

			if len(missingScopeDecls) == 0 {
				lines = append(lines, "services_missing_required_user_scopes: (none)")
			} else {
				lines = append(lines, fmt.Sprintf("services_missing_required_user_scopes: %s", strings.Join(missingScopeDecls, ", ")))
			}

			if suggestedCmd == "" {
				lines = append(lines, "suggested_user_login_command: (none)")
			} else {
				lines = append(lines, fmt.Sprintf("suggested_user_login_command: %s", suggestedCmd))
			}

			return state.Printer.Print(payload, strings.Join(lines, "\n"))
		},
	}
	return cmd
}
