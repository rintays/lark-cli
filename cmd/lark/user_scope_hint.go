package main

import (
	"fmt"
	"regexp"
	"strings"
)

var scopeBracketPattern = regexp.MustCompile(`\[(.*?)\]`)

func withUserScopeHint(err error) error {
	if err == nil {
		return nil
	}
	msg := err.Error()
	if strings.Contains(msg, "Re-authorize with:") {
		return err
	}
	scopes := extractScopesFromErrorMessage(msg)
	if len(scopes) == 0 {
		return err
	}
	scopes = ensureOfflineAccess(selectPreferredScopes(scopes))
	scopeArg := strings.Join(scopes, " ")
	hint := fmt.Sprintf("Missing user OAuth scopes: %s.\nRe-authorize with:\n  lark auth user login --scopes %q --force-consent", strings.Join(scopes, ", "), scopeArg)
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
