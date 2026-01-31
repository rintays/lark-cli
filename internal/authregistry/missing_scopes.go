package authregistry

import "fmt"

// ServicesMissingRequiredUserScopes returns the subset of services that:
// - exist in the Registry
// - require a user token
// - but do not declare any RequiredUserScopes yet (nil slice)
//
// This is used to avoid pretending we know the correct scope strings when we
// don’t, and to drive remediation UX later ("please re-login with ...").
func ServicesMissingRequiredUserScopes(services []string) ([]string, error) {
	services = normalizeServices(services)
	missing := make([]string, 0, len(services))
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", name)
		}
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
		if def.RequiredUserScopes == nil {
			missing = append(missing, name)
		}
	}
	return uniqueSorted(missing), nil
}
