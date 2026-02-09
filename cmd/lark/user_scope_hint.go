package main

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

var scopeBracketPattern = regexp.MustCompile(`\[(.*?)\]`)
var errorCodePattern = regexp.MustCompile(`code=(\d+)`)

func withUserScopeHint(err error) error {
	return withUserScopeHintForCommand(nil, err)
}

func withUserScopeHintForCommand(state *appState, err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "Re-authorize with:") {
		return err
	}
	if !shouldSuggestScopes(msg) {
		return err
	}

	scopes := extractScopesFromErrorMessage(msg)
	fromError := len(scopes) > 0
	var services []string
	var undeclared []string
	if !fromError && state != nil {
		inferredServices, inferred, missingDecls, ok, inferErr := userOAuthScopesForCommand(state.Command)
		if inferErr == nil && ok {
			scopes = inferred
			services = inferredServices
			undeclared = missingDecls
		}
	}
	if len(scopes) == 0 {
		return err
	}

	scopes = ensureOfflineAccess(selectPreferredScopes(scopes))
	scopeArg := strings.Join(scopes, " ")
	reloginCmd := ""
	if !fromError && len(services) > 0 && len(undeclared) == 0 {
		reloginCmd = fmt.Sprintf("lark auth user login --services %q --force-consent", strings.Join(services, ","))
	}
	if reloginCmd == "" {
		reloginCmd = fmt.Sprintf("lark auth user login --scopes %q --force-consent", scopeArg)
	}
	hint := fmt.Sprintf("Missing user OAuth scopes: %s.\nRe-authorize with:\n  %s", strings.Join(scopes, ", "), reloginCmd)
	return fmt.Errorf("%s\n%s", msg, hint)
}

func extractScopesFromErrorMessage(msg string) []string {
	matches := scopeBracketPattern.FindAllStringSubmatch(msg, -1)
	for _, match := range matches {
		if len(match) < 2 {
			continue
		}
		scopes := normalizeScopes(parseScopeList(match[1]))
		if len(scopes) == 0 {
			continue
		}
		if containsScopeToken(scopes) {
			return scopes
		}
	}
	return nil
}

func containsScopeToken(scopes []string) bool {
	for _, scope := range scopes {
		if strings.Contains(scope, ":") {
			return true
		}
	}
	return false
}

func shouldSuggestScopes(msg string) bool {
	// Prefer structured classification.
	// If we have a numeric OpenAPI error code, only suggest re-auth when it is the
	// well-known "insufficient OAuth scopes" error.
	if code := extractErrorCode(msg); code != 0 {
		return code == 99991679
	}

	// If the upstream error string does not include a code, avoid guessing based
	// on generic "permission denied" keywords. Those are often caused by Drive/Wiki
	// object-level ACLs (e.g. view-only documents), not missing OAuth scopes.
	lower := strings.ToLower(msg)
	if strings.Contains(lower, "scope") || strings.Contains(lower, "scopes") {
		return true
	}
	if strings.Contains(msg, "权限") && (strings.Contains(msg, "scope") || strings.Contains(msg, "权限范围")) {
		return true
	}
	return false
}

func extractErrorCode(msg string) int {
	match := errorCodePattern.FindStringSubmatch(msg)
	if len(match) != 2 {
		return 0
	}
	value, err := strconv.Atoi(match[1])
	if err != nil {
		return 0
	}
	return value
}

func selectPreferredScopes(scopes []string) []string {
	seen := map[string]struct{}{}
	out := make([]string, 0, len(scopes))
	readonly := false
	full := false
	for _, scope := range scopes {
		if scope == "drive:drive:readonly" {
			readonly = true
		}
		if scope == "drive:drive" {
			full = true
		}
	}
	for _, scope := range scopes {
		if scope == "drive:drive" && readonly {
			continue
		}
		if scope == "drive:drive:readonly" && full && !readonly {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	if readonly {
		if _, ok := seen["drive:drive:readonly"]; !ok {
			out = append(out, "drive:drive:readonly")
		}
	}
	return out
}
