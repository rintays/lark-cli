package output

import "github.com/charmbracelet/lipgloss"

type Theme struct {
	Styled         bool
	headerStyle    lipgloss.Style
	separatorStyle lipgloss.Style
	warnLabel      lipgloss.Style
	errorLabel     lipgloss.Style
	hintLabel      lipgloss.Style
	infoLabel      lipgloss.Style
	successLabel   lipgloss.Style
	sectionTitle   lipgloss.Style
}

func NewTheme(styled bool) Theme {
	theme := Theme{Styled: styled}
	if !styled {
		return theme
	}
	theme.headerStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "62", Dark: "81"})
	theme.separatorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.AdaptiveColor{Light: "245", Dark: "240"})
	theme.warnLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "166", Dark: "214"})
	theme.errorLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "160", Dark: "196"})
	theme.hintLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "24", Dark: "39"})
	theme.infoLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "67", Dark: "75"})
	theme.successLabel = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "28", Dark: "42"})
	theme.sectionTitle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.AdaptiveColor{Light: "238", Dark: "248"})
	return theme
}

func (t Theme) RenderHeader(text string) string {
	if !t.Styled {
		return text
	}
	return t.headerStyle.Render(text)
}

func (t Theme) RenderSeparator(text string) string {
	if !t.Styled {
		return text
	}
	return t.separatorStyle.Render(text)
}

func (t Theme) RenderLabel(label string) string {
	if !t.Styled {
		return label
	}
	switch label {
	case "WARNING", "WARN":
		return t.warnLabel.Render(label)
	case "ERROR":
		return t.errorLabel.Render(label)
	case "HINT":
		return t.hintLabel.Render(label)
	case "INFO":
		return t.infoLabel.Render(label)
	case "OK", "SUCCESS":
		return t.successLabel.Render(label)
	default:
		return t.infoLabel.Render(label)
	}
}

func (t Theme) RenderSectionTitle(text string) string {
	if !t.Styled {
		return text
	}
	return t.sectionTitle.Render(text)
}
