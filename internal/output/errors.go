package output

import (
	"errors"
	"strings"
)

type UsageError struct {
	Message string
	Usage   string
	Hint    string
}

func (e UsageError) Error() string {
	if strings.TrimSpace(e.Message) == "" {
		return "invalid usage"
	}
	return e.Message
}

func FormatError(err error, styled bool) string {
	if err == nil {
		return ""
	}
	theme := NewTheme(styled)
	var usageErr UsageError
	if errors.As(err, &usageErr) {
		return formatUsageError(usageErr, theme)
	}
	msg := strings.TrimSpace(err.Error())
	if msg == "" {
		return ""
	}
	if !styled {
		return msg
	}
	return theme.RenderLabel("ERROR") + " " + msg
}

func formatUsageError(err UsageError, theme Theme) string {
	var b strings.Builder
	message := strings.TrimSpace(err.Message)
	if message == "" {
		message = "invalid usage"
	}
	if theme.Styled {
		b.WriteString(theme.RenderLabel("ERROR"))
		b.WriteString(" ")
		b.WriteString(message)
	} else {
		b.WriteString(message)
	}
	if usage := strings.TrimSpace(err.Usage); usage != "" {
		b.WriteString("\n\n")
		if theme.Styled {
			b.WriteString(theme.RenderSectionTitle("Usage"))
		} else {
			b.WriteString("Usage")
		}
		b.WriteString("\n")
		b.WriteString(indentBlock(usage, "  "))
	}
	if hint := strings.TrimSpace(err.Hint); hint != "" {
		b.WriteString("\n\n")
		if theme.Styled {
			b.WriteString(theme.RenderSectionTitle("Hint"))
		} else {
			b.WriteString("Hint")
		}
		b.WriteString("\n")
		b.WriteString(indentBlock(hint, "  "))
	}
	return b.String()
}

func indentBlock(text, prefix string) string {
	lines := strings.Split(text, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = prefix + line
	}
	return strings.Join(lines, "\n")
}
