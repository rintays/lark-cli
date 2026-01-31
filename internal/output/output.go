package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

type Printer struct {
	Writer io.Writer
	JSON   bool
	Styled bool
}

func (p Printer) Print(v any, text string) error {
	if p.JSON {
		enc := json.NewEncoder(p.Writer)
		enc.SetIndent("", "  ")
		return enc.Encode(v)
	}
	if p.Styled {
		text = FormatText(text)
	}
	_, err := fmt.Fprintln(p.Writer, text)
	return err
}

func AutoStyle(w io.Writer) bool {
	type fdWriter interface {
		Fd() uintptr
	}
	if w == nil {
		return false
	}
	fw, ok := w.(fdWriter)
	if !ok {
		return false
	}
	fd := fw.Fd()
	return isatty.IsTerminal(fd) || isatty.IsCygwinTerminal(fd)
}

func FormatText(text string) string {
	if text == "" {
		return text
	}
	lines := strings.Split(text, "\n")
	hasTabs := false
	for _, line := range lines {
		if strings.Contains(line, "\t") {
			hasTabs = true
			break
		}
	}
	if !hasTabs {
		return text
	}

	rows := make([][]string, len(lines))
	maxCols := 0
	for i, line := range lines {
		cols := strings.Split(line, "\t")
		rows[i] = cols
		if len(cols) > maxCols {
			maxCols = len(cols)
		}
	}
	if maxCols == 0 {
		return text
	}

	widths := make([]int, maxCols)
	for _, row := range rows {
		for i, col := range row {
			w := lipgloss.Width(col)
			if w > widths[i] {
				widths[i] = w
			}
		}
	}

	out := make([]string, 0, len(rows))
	for _, row := range rows {
		if len(row) == 1 && row[0] == "" {
			out = append(out, "")
			continue
		}
		var b strings.Builder
		for i, col := range row {
			if i >= len(widths) {
				break
			}
			style := lipgloss.NewStyle().Width(widths[i]).Align(lipgloss.Left)
			b.WriteString(style.Render(col))
			if i < len(row)-1 {
				b.WriteString("  ")
			}
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return strings.Join(out, "\n")
}
