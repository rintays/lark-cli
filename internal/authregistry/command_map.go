package authregistry

import "strings"

// commandServiceMap defines stable mappings from canonical CLI command paths to
// service names in the registry.
//
// Keys are space-separated command paths. Values are service names.
var commandServiceMap = map[string][]string{
	"drive":    {"drive"},
	"docs":     {"docs"},
	"sheets":   {"sheets"},
	"mail":     {"mail"},
	"wiki":     {"wiki"},
	"base":     {"base"},
	"calendar":   {"calendar"},
	"calendars":  {"calendar"},
	"chats":    {"im"},
	"messages": {"im"},
	"msg":      {"im"},

	// Internal aliases (not currently exposed as CLI roots).
	"im": {"im"},
}

// ServicesForCommandPath returns the service names for a canonical command path.
//
// Matching is longest-prefix, so e.g. ["drive", "list"] maps via the "drive"
// entry unless a more specific mapping exists.
//
// The returned list is stable-sorted and de-duped.
func ServicesForCommandPath(path []string) ([]string, bool) {
	parts := normalizeCommandParts(path)
	if len(parts) == 0 {
		return nil, false
	}
	for i := len(parts); i > 0; i-- {
		key := strings.Join(parts[:i], " ")
		services, ok := commandServiceMap[key]
		if !ok {
			continue
		}
		return uniqueSorted(services), true
	}
	return nil, false
}

// ServicesForCommand returns the service names for a space-separated command
// path string (for example, "drive" or "chats list"). The returned list is
// stable-sorted and de-duped.
func ServicesForCommand(command string) ([]string, bool) {
	return ServicesForCommandPath(strings.Fields(command))
}

func normalizeCommandParts(path []string) []string {
	if len(path) == 0 {
		return nil
	}
	parts := make([]string, 0, len(path))
	for _, part := range path {
		part = strings.ToLower(strings.TrimSpace(part))
		if part == "" {
			return nil
		}
		parts = append(parts, part)
	}
	if len(parts) == 0 {
		return nil
	}
	return parts
}
