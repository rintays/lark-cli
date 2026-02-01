package output

import "strings"

type NoticeKind string

const (
	NoticeSuccess NoticeKind = "OK"
	NoticeWarning NoticeKind = "WARNING"
	NoticeError   NoticeKind = "ERROR"
	NoticeHint    NoticeKind = "HINT"
	NoticeInfo    NoticeKind = "INFO"
)

func Notice(kind NoticeKind, title string, lines []string) string {
	label := strings.TrimSpace(string(kind))
	if label == "" {
		label = string(NoticeInfo)
	}
	var b strings.Builder
	if title != "" {
		b.WriteString(label)
		b.WriteString(": ")
		b.WriteString(title)
	} else {
		b.WriteString(label)
		b.WriteString(":")
	}
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		b.WriteString("\n  - ")
		b.WriteString(trimmed)
	}
	return b.String()
}

func JoinBlocks(blocks ...string) string {
	parts := make([]string, 0, len(blocks))
	for _, block := range blocks {
		if strings.TrimSpace(block) == "" {
			continue
		}
		parts = append(parts, strings.TrimRight(block, " \t\n"))
	}
	return strings.Join(parts, "\n\n")
}
