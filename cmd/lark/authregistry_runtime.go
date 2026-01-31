package main

import (
	"fmt"
	"strings"

	"lark/internal/authregistry"
)

// userOAuthScopesForCommand returns the recommended user OAuth scopes for a
// command based on the command→service mapping and the service registry.
//
// ok reports whether the command matched a mapping and the mapped services
// require a user access token.
func userOAuthScopesForCommand(command string) (services []string, scopes []string, undeclared []string, ok bool, err error) {
	command = strings.TrimSpace(command)
	if command == "" {
		return nil, nil, nil, false, nil
	}

	services, matched := authregistry.ServicesForCommand(command)
	if !matched {
		return nil, nil, nil, false, nil
	}

	tokenTypes, err := authregistry.TokenTypesFromServices(services)
	if err != nil {
		return nil, nil, nil, false, err
	}

	requiresUser := false
	for _, tt := range tokenTypes {
		if tt == authregistry.TokenUser {
			requiresUser = true
			break
		}
	}
	if !requiresUser {
		return nil, nil, nil, false, nil
	}

	requiredUserScopes, undeclared, err := authregistry.RequiredUserScopesFromServicesReport(services)
	if err != nil {
		return nil, nil, nil, false, err
	}

	scopes = ensureOfflineAccess(requiredUserScopes)
	return services, scopes, undeclared, true, nil
}

// userOAuthReloginCommandForCommand returns a recommended re-login command for
// the given CLI command path.
//
// command should be a space-separated command path, excluding the root binary
// name (for example, "mail send").
func userOAuthReloginCommandForCommand(command string) (reloginCmd string, note string, ok bool, err error) {
	services, scopes, undeclared, ok, err := userOAuthScopesForCommand(command)
	if err != nil || !ok {
		return "", "", ok, err
	}

	scopeArg := strings.Join(scopes, " ")
	reloginCmd = fmt.Sprintf("lark auth user login --scopes %q --force-consent", scopeArg)

	command = strings.TrimSpace(command)
	note = fmt.Sprintf("required by command %q (services: %s)", command, strings.Join(services, ", "))
	if len(undeclared) > 0 {
		note = fmt.Sprintf("required by command %q (services: %s; missing scope declarations for: %s)", command, strings.Join(services, ", "), strings.Join(undeclared, ", "))
	}

	return reloginCmd, note, true, nil
}

func userOAuthReloginRecommendation(state *appState) (cmd string, note string) {
	if state == nil {
		return userOAuthReloginCommand, ""
	}
	reloginCmd, note, ok, err := userOAuthReloginCommandForCommand(state.Command)
	if err != nil || !ok || reloginCmd == "" {
		return userOAuthReloginCommand, ""
	}
	return reloginCmd, note
}
