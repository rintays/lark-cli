package authregistry

import (
	"errors"
	"fmt"
	"sort"
	"strings"
)

// DefaultUserOAuthServices is the default service set used when the user enables
// the services-based OAuth mode but doesn't specify an explicit list.
var DefaultUserOAuthServices = []string{"drive"}

// UserOAuthServiceAliases maps user-facing aliases to service names.
var UserOAuthServiceAliases = map[string][]string{
	"all":  {"drive", "docx", "sheets"},
	"user": {"drive", "docx", "sheets"},
}

// ListUserOAuthServices returns services that can be used in services-based
// user OAuth flows.
//
// Services may exist in Registry without scopes yet; those are intentionally
// excluded to avoid suggesting incorrect scope strings.
func ListUserOAuthServices() []string {
	services := make([]string, 0, len(Registry))
	for name, def := range Registry {
		requiresUser := false
		for _, tt := range def.TokenTypes {
			if tt == TokenUser {
				requiresUser = true
				break
			}
		}
		if !requiresUser {
			continue
		}
		if len(def.UserScopes.Full) == 0 && len(def.UserScopes.Readonly) == 0 && len(def.RequiredUserScopes) == 0 {
			continue
		}
		services = append(services, name)
	}
	sort.Strings(services)
	return services
}

// ExpandUserOAuthServiceAliases expands aliases (like "all") into concrete service
// names. Input is normalized (lowercased, trimmed, de-duped).
func ExpandUserOAuthServiceAliases(services []string) []string {
	services = normalizeServices(services)
	expanded := make([]string, 0, len(services))
	for _, svc := range services {
		if alias, ok := UserOAuthServiceAliases[svc]; ok {
			expanded = append(expanded, alias...)
			continue
		}
		expanded = append(expanded, svc)
	}
	return normalizeServices(expanded)
}

// UserOAuthScopesFromServices returns the stable, de-duped union of OAuth scopes
// required for the given services.
//
// The resulting slice is deterministically sorted so that auth URL diffs are
// predictable and testable.
func UserOAuthScopesFromServices(services []string, readonly bool, driveScope string) ([]string, error) {
	driveScope = strings.ToLower(strings.TrimSpace(driveScope))
	if driveScope != "" {
		switch driveScope {
		case "full", "readonly":
		case "file":
			return nil, errors.New("drive-scope file is not supported; use full or readonly")
		default:
			return nil, fmt.Errorf("invalid drive-scope %q (use full or readonly)", driveScope)
		}
	}
	if readonly {
		if driveScope != "" {
			return nil, errors.New("drive-scope cannot be combined with --readonly")
		}
		driveScope = "readonly"
	}
	if driveScope == "" {
		driveScope = "full"
	}

	services = ExpandUserOAuthServiceAliases(services)
	if len(services) == 0 {
		services = DefaultUserOAuthServices
	}

	var scopes []string
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown service %q (use `lark auth user services` to list supported services)", name)
		}

		requiresUser := false
		for _, tt := range def.TokenTypes {
			if tt == TokenUser {
				requiresUser = true
				break
			}
		}
		if !requiresUser {
			return nil, fmt.Errorf("service %q does not require user OAuth", name)
		}

		// Prefer the requested variant when the service declares it; otherwise
		// fall back to any declared scopes.
		if driveScope == "readonly" {
			if len(def.UserScopes.Readonly) > 0 {
				scopes = append(scopes, def.UserScopes.Readonly...)
				continue
			}
			if len(def.UserScopes.Full) > 0 {
				scopes = append(scopes, def.UserScopes.Full...)
				continue
			}
			if len(def.RequiredUserScopes) > 0 {
				scopes = append(scopes, def.RequiredUserScopes...)
				continue
			}
			return nil, fmt.Errorf("service %q does not declare user OAuth scopes yet", name)
		}

		if len(def.UserScopes.Full) > 0 {
			scopes = append(scopes, def.UserScopes.Full...)
			continue
		}
		if len(def.UserScopes.Readonly) > 0 {
			scopes = append(scopes, def.UserScopes.Readonly...)
			continue
		}
		if len(def.RequiredUserScopes) > 0 {
			scopes = append(scopes, def.RequiredUserScopes...)
			continue
		}
		return nil, fmt.Errorf("service %q does not declare user OAuth scopes yet", name)
	}
	return uniqueSorted(scopes), nil
}

func normalizeServices(services []string) []string {
	seen := make(map[string]struct{}, len(services))
	out := make([]string, 0, len(services))
	for _, svc := range services {
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

func uniqueSorted(items []string) []string {
	items = normalizeStrings(items)
	sort.Strings(items)
	return items
}

func normalizeStrings(items []string) []string {
	seen := make(map[string]struct{}, len(items))
	out := make([]string, 0, len(items))
	for _, item := range items {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := seen[item]; ok {
			continue
		}
		seen[item] = struct{}{}
		out = append(out, item)
	}
	return out
}
