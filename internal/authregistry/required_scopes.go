package authregistry

import (
	"fmt"
)

// RequiredUserScopesFromServices returns the stable, de-duped union of
// RequiredUserScopes declared by the given services.
//
// Services with no RequiredUserScopes are allowed and simply contribute nothing
// to the union.
func RequiredUserScopesFromServices(services []string) ([]string, error) {
	services = normalizeServices(services)
	var scopes []string
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", name)
		}
		scopes = append(scopes, def.RequiredUserScopes...)
	}
	return uniqueSorted(scopes), nil
}

// RequiresOfflineFromServices reports whether any of the given services declares
// RequiresOffline.
func RequiresOfflineFromServices(services []string) (bool, error) {
	services = normalizeServices(services)
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return false, fmt.Errorf("unknown service %q", name)
		}
		if def.RequiresOffline {
			return true, nil
		}
	}
	return false, nil
}
