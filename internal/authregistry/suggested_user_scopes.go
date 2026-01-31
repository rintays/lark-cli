package authregistry

import "fmt"

// SuggestedUserOAuthScopesFromServices returns a stable, de-duped union of user
// OAuth scopes for the given services.
//
// When a service declares UserScopes variants, the requested variant is used.
// When a service does not declare variants, this falls back to the service's
// RequiredUserScopes.
func SuggestedUserOAuthScopesFromServices(services []string, readonly bool) ([]string, error) {
	services = normalizeServices(services)
	var scopes []string
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", name)
		}

		if readonly {
			if len(def.UserScopes.Readonly) > 0 {
				scopes = append(scopes, def.UserScopes.Readonly...)
				continue
			}
		} else {
			if len(def.UserScopes.Full) > 0 {
				scopes = append(scopes, def.UserScopes.Full...)
				continue
			}
		}

		// Fallback when variants are unavailable (or not yet declared).
		scopes = append(scopes, def.RequiredUserScopes...)
	}
	return uniqueSorted(scopes), nil
}
