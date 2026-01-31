package authregistry

import "fmt"

// TokenTypesFromServices returns the stable, de-duped union of TokenTypes
// declared by the given services.
func TokenTypesFromServices(services []string) ([]TokenType, error) {
	services = normalizeServices(services)
	var types []string
	for _, name := range services {
		def, ok := Registry[name]
		if !ok {
			return nil, fmt.Errorf("unknown service %q", name)
		}
		for _, tt := range def.TokenTypes {
			types = append(types, string(tt))
		}
	}
	types = uniqueSorted(types)
	out := make([]TokenType, 0, len(types))
	for _, t := range types {
		out = append(out, TokenType(t))
	}
	return out, nil
}
