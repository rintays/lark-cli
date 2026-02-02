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
	theme := NewTheme(true)
	lines := strings.Split(text, "\n")
	hasTabs := false
	for _, line := range lines {
		if strings.Contains(line, "\t") {
			hasTabs = true
			break
		}
	}
	if !hasTabs {
		styled := make([]string, len(lines))
		for i, line := range lines {
			styled[i] = styleNoticeLine(line, theme)
		}
		return strings.Join(styled, "\n")
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
		styled := make([]string, len(lines))
		for i, line := range lines {
			styled[i] = styleNoticeLine(line, theme)
		}
		return strings.Join(styled, "\n")
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

	headerRow := -1
	separatorRow := -1
	if len(rows) >= 2 && isSeparatorRow(rows[1]) {
		headerRow = 0
		separatorRow = 1
	}

	out := make([]string, 0, len(rows))
	for rowIndex, row := range rows {
		if len(row) == 1 && row[0] == "" {
			out = append(out, "")
			continue
		}
		var b strings.Builder
		for colIndex, col := range row {
			if colIndex >= len(widths) {
				break
			}
			cell := col
			if rowIndex == headerRow {
				cell = theme.RenderHeader(cell)
			}
			if rowIndex == separatorRow {
				cell = theme.RenderSeparator(cell)
			}
			style := lipgloss.NewStyle().Width(widths[colIndex]).Align(lipgloss.Left)
			b.WriteString(style.Render(cell))
			if colIndex < len(row)-1 {
				b.WriteString("  ")
			}
		}
		out = append(out, strings.TrimRight(b.String(), " "))
	}
	return strings.Join(out, "\n")
}

func TableText(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}
	widths := make([]int, len(headers))
	for i, header := range headers {
		widths[i] = lipgloss.Width(header)
	}
	for _, row := range rows {
		for i, col := range row {
			if i >= len(widths) {
				break
			}
			if w := lipgloss.Width(col); w > widths[i] {
				widths[i] = w
			}
		}
	}

	lines := make([]string, 0, len(rows)+2)
	lines = append(lines, strings.Join(headers, "\t"))
	separator := make([]string, len(headers))
	for i, width := range widths {
		if width < 1 {
			width = 1
		}
		separator[i] = strings.Repeat("-", width)
	}
	lines = append(lines, strings.Join(separator, "\t"))
	for _, row := range rows {
		if len(row) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, row)
			row = padded
		}
		if len(row) > len(headers) {
			row = row[:len(headers)]
		}
		lines = append(lines, strings.Join(row, "\t"))
	}
	return strings.Join(lines, "\n")
}

func TableTSV(headers []string, rows [][]string) string {
	if len(headers) == 0 {
		return ""
	}
	lines := make([]string, 0, len(rows)+1)
	lines = append(lines, strings.Join(headers, "\t"))
	for _, row := range rows {
		if len(row) < len(headers) {
			padded := make([]string, len(headers))
			copy(padded, row)
			row = padded
		}
		if len(row) > len(headers) {
			row = row[:len(headers)]
		}
		lines = append(lines, strings.Join(row, "\t"))
	}
	return strings.Join(lines, "\n")
}

func TableTextFromLines(headers []string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return TableText(headers, rows)
}

func TableTSVFromLines(headers []string, lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	rows := make([][]string, 0, len(lines))
	for _, line := range lines {
		rows = append(rows, strings.Split(line, "\t"))
	}
	return TableTSV(headers, rows)
}
