package main

import (
	"errors"
	"fmt"
	"strings"
)

func preflightUserScopes(state *appState) error {
	if state == nil || state.Config == nil {
		return nil
	}
	account := resolveUserAccountName(state)
	acct, ok := loadUserAccount(state.Config, account)
	if !ok {
		return nil
	}
	granted := normalizeScopes(parseScopeList(acct.UserAccessTokenScope))
	if len(granted) == 0 {
		return nil
	}
	_, required, undeclared, ok, err := userOAuthScopesForCommand(state.Command)
	if err != nil || !ok || len(required) == 0 || len(undeclared) > 0 {
		return nil
	}
	missing := missingScopes(required, granted)
	if len(missing) == 0 {
		return nil
	}
	reloginCmd, note, _, _ := userOAuthReloginCommandForCommand(state.Command)
	if reloginCmd == "" {
		reloginCmd = userOAuthReloginCommand
	}
	message := fmt.Sprintf("user OAuth scopes missing for account %q: %s; run `%s`", account, strings.Join(missing, ", "), reloginCmd)
	if note != "" {
		message = fmt.Sprintf("%s; %s", message, note)
	}
	return errors.New(message)
}

func missingScopes(required []string, granted []string) []string {
	if len(required) == 0 {
		return nil
	}
	grantedSet := make(map[string]struct{}, len(granted))
	for _, scope := range granted {
		grantedSet[scope] = struct{}{}
	}
	missing := make([]string, 0, len(required))
	for _, scope := range required {
		if scopeSatisfied(scope, grantedSet) {
			continue
		}
		missing = append(missing, scope)
	}
	return missing
}

func scopeSatisfied(required string, granted map[string]struct{}) bool {
	if _, ok := granted[required]; ok {
		return true
	}
	if strings.HasSuffix(required, ":readonly") {
		base := strings.TrimSuffix(required, ":readonly")
		if _, ok := granted[base]; ok {
			return true
		}
		return false
	}
	readonly := required + ":readonly"
	if _, ok := granted[readonly]; ok {
		return true
	}
	return false
}
