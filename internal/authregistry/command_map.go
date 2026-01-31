package authregistry

import "strings"

// commandServiceMap defines stable mappings from canonical CLI command paths to
// service names in the registry. Keys are space-separated command paths.
var commandServiceMap = map[string][]string{
	"drive":    {"drive"},
	"docs":     {"docs"},
	"sheets":   {"sheets"},
	"mail":     {"mail"},
	"wiki":     {"wiki"},
	"base":     {"base"},
	"calendar": {"calendar"},
	"chats":    {"im"},
	"msg":      {"im"},
}

// ServicesForCommandPath returns the service names for a canonical command path.
// The returned list is stable-sorted and de-duped.
func ServicesForCommandPath(path []string) ([]string, bool) {
	key := normalizeCommandPath(path)
	if key == "" {
		return nil, false
	}
	services, ok := commandServiceMap[key]
	if !ok {
		return nil, false
	}
	return uniqueSorted(services), true
}

// ServicesForCommand returns the service names for a space-separated command
// path string (for example, "drive" or "chats list"). The returned list is
// stable-sorted and de-duped.
func ServicesForCommand(command string) ([]string, bool) {
	return ServicesForCommandPath(strings.Fields(command))
}

func normalizeCommandPath(path []string) string {
	if len(path) == 0 {
		return ""
	}
	parts := make([]string, 0, len(path))
	for _, part := range path {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			return ""
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, " ")
}
