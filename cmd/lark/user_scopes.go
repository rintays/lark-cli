package main

import (
	"errors"
	"sort"
	"strings"

	"lark/internal/authregistry"
)

type userOAuthScopeOptions struct {
	Scopes        string
	ScopesSet     bool
	Services      []string
	ServicesSet   bool
	Readonly      bool
	DriveScope    string
	DriveScopeSet bool
}

func resolveUserOAuthScopes(state *appState, opts userOAuthScopeOptions) ([]string, string, error) {
	if opts.ScopesSet {
		scopes := normalizeScopes(parseScopeList(opts.Scopes))
		if len(scopes) == 0 {
			return nil, "", errors.New("scopes must not be empty")
		}
		return canonicalizeUserOAuthScopes(scopes), "flag", nil
	}

	if opts.ServicesSet || opts.Readonly || opts.DriveScopeSet {
		services := opts.Services
		if len(services) == 0 {
			services = authregistry.DefaultUserOAuthServices
		}
		scopes, err := authregistry.UserOAuthScopesFromServices(services, opts.Readonly, opts.DriveScope)
		if err != nil {
			return nil, "", err
		}
		return canonicalizeUserOAuthScopes(scopes), "services", nil
	}

	if state != nil && state.Config != nil {
		account := resolveUserAccountName(state)
		if acct, ok := loadUserAccount(state.Config, account); ok && len(acct.UserScopes) > 0 {
			scopes := normalizeScopes(acct.UserScopes)
			return canonicalizeUserOAuthScopes(scopes), "account", nil
		}
		if len(state.Config.UserScopes) > 0 {
			scopes := normalizeScopes(state.Config.UserScopes)
			return canonicalizeUserOAuthScopes(scopes), "config", nil
		}
	}

	return []string{defaultUserOAuthScope}, "default", nil
}

func canonicalizeUserOAuthScopes(scopes []string) []string {
	scopes = normalizeScopes(scopes)

	rest := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if scope == defaultUserOAuthScope {
			continue
		}
		rest = append(rest, scope)
	}
	sort.Strings(rest)

	out := make([]string, 0, len(rest)+1)
	out = append(out, defaultUserOAuthScope)
	out = append(out, rest...)
	return out
}

func parseScopeList(raw string) []string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return nil
	}
	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == '\n' || r == '\t' || r == ' '
	})
	return normalizeScopes(fields)
}

func normalizeScopes(scopes []string) []string {
	seen := make(map[string]struct{}, len(scopes))
	out := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		scope = strings.TrimSpace(scope)
		if scope == "" {
			continue
		}
		if _, ok := seen[scope]; ok {
			continue
		}
		seen[scope] = struct{}{}
		out = append(out, scope)
	}
	return out
}

func ensureOfflineAccess(scopes []string) []string {
	return canonicalizeUserOAuthScopes(scopes)
}

func requestedUserOAuthScopes(scopes []string, grantedScope string, incremental bool) []string {
	scopes = normalizeScopes(scopes)
	if !incremental || strings.TrimSpace(grantedScope) == "" {
		return canonicalizeUserOAuthScopes(scopes)
	}
	granted := normalizeScopes(parseScopeList(grantedScope))
	grantedSet := make(map[string]struct{}, len(granted))
	for _, scope := range granted {
		grantedSet[scope] = struct{}{}
	}
	delta := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		if scope == "" {
			continue
		}
		if scope != defaultUserOAuthScope {
			if _, ok := grantedSet[scope]; ok {
				continue
			}
		}
		delta = append(delta, scope)
	}
	return canonicalizeUserOAuthScopes(delta)
}

func joinScopes(scopes []string) string {
	return strings.Join(canonicalizeUserOAuthScopes(scopes), " ")
}

func parseServicesList(raw []string) []string {
	parts := make([]string, 0, len(raw))
	for _, entry := range raw {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		for _, part := range strings.FieldsFunc(entry, func(r rune) bool { return r == ',' || r == ' ' || r == '\t' || r == '\n' }) {
			part = strings.ToLower(strings.TrimSpace(part))
			if part != "" {
				parts = append(parts, part)
			}
		}
	}
	return normalizeServices(parts)
}

func normalizeServices(services []string) []string {
	seen := make(map[string]struct{}, len(services))
	out := make([]string, 0, len(services))
	for _, svc := range services {
		if svc == "" {
			continue
		}
		svc = strings.ToLower(strings.TrimSpace(svc))
		if svc == "" {
			continue
		}
		if _, ok := seen[svc]; ok {
			continue
		}
		seen[svc] = struct{}{}
		out = append(out, svc)
	}
	return out
}
