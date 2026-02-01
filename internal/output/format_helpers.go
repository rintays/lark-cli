package output

import (
	"strings"
)

func styleNoticeLine(line string, theme Theme) string {
	trimmed := strings.TrimLeft(line, " \t")
	if trimmed == "" {
		return line
	}
	prefixes := []string{
		"WARNING:",
		"WARN:",
		"ERROR:",
		"HINT:",
		"NOTE:",
		"INFO:",
		"OK:",
		"SUCCESS:",
	}
	for _, prefix := range prefixes {
		if strings.HasPrefix(trimmed, prefix) {
			lead := line[:len(line)-len(trimmed)]
			rest := strings.TrimPrefix(trimmed, prefix)
			label := strings.TrimSuffix(prefix, ":")
			return lead + theme.RenderLabel(label) + ":" + rest
		}
	}
	return line
}

func isSeparatorRow(cols []string) bool {
	if len(cols) == 0 {
		return false
	}
	for _, col := range cols {
		trimmed := strings.TrimSpace(col)
		if trimmed == "" {
			continue
		}
		for _, r := range trimmed {
			if r != '-' {
				return false
			}
		}
	}
	return true
}
