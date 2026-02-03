package authregistry

import "strings"

// commandServiceMap defines stable mappings from canonical CLI command paths to
// service names in the registry.
//
// Keys are space-separated command paths. Values are service names.
var commandServiceMap = map[string][]string{
	"drive":                 {"drive"},
	"drive export":          {"drive-export"},
	"docs":                  {"docs"},
	"docs export":           {"drive-export"},
	"sheets":                {"sheets"},
	"mail":                  {"mail"},
	"mail send":             {"mail-send"},
	"mail public-mailboxes": {"mail-public"},
	"mail mailboxes":        {"mail-public"},
	"wiki":                  {"wiki"},
	"base":                  {"base"},
	"bases":                 {"base"},
	"calendar":              {"calendar"},
	"calendars":             {"calendar"},
	"tasks":                 {"task"},
	"tasks create":          {"task-write"},
	"tasks update":          {"task-write"},
	"tasks delete":          {"task-write"},
	"tasklists":             {"tasklist"},
	"tasklists create":      {"tasklist-write"},
	"tasklists update":      {"tasklist-write"},
	"tasklists delete":      {"tasklist-write"},
	"chats":                 {"im"},
	"messages":              {"im"},
	"msg":                   {"im"},
	"msg search":            {"search-message"},
	"messages search":       {"search-message"},
	"users search":          {"search-user"},

	// Internal aliases (not currently exposed as CLI roots).
	"im": {"im"},
}

var defaultCommandServiceMap = copyCommandServiceMap(commandServiceMap)

// SetCommandServiceMap merges command service overrides into the default map.
// This lets the CLI attach metadata close to command definitions without
// rewriting the entire registry at once.
func SetCommandServiceMap(overrides map[string][]string) {
	if len(overrides) == 0 {
		return
	}
	merged := copyCommandServiceMap(defaultCommandServiceMap)
	for key, services := range overrides {
		merged[key] = append([]string(nil), services...)
	}
	commandServiceMap = merged
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

func copyCommandServiceMap(in map[string][]string) map[string][]string {
	if in == nil {
		return nil
	}
	out := make(map[string][]string, len(in))
	for key, values := range in {
		out[key] = append([]string(nil), values...)
	}
	return out
}
