package output

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
	liptable "github.com/charmbracelet/lipgloss/table"
	"github.com/charmbracelet/x/term"
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
		text = FormatText(text, terminalWidth(p.Writer))
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

func FormatText(text string, width int) string {
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
	headers, dataRows := splitTableHeader(rows)
	return renderTable(theme, headers, dataRows, width)
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

	lines := make([]string, 0, len(rows))
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
	lines := make([]string, 0, len(rows))
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

func terminalWidth(w io.Writer) int {
	if w == nil {
		return 0
	}
	type fdWriter interface {
		Fd() uintptr
	}
	fw, ok := w.(fdWriter)
	if !ok {
		return 0
	}
	width, _, err := term.GetSize(fw.Fd())
	if err != nil || width <= 0 {
		return 0
	}
	return width
}

func splitTableHeader(rows [][]string) ([]string, [][]string) {
	if len(rows) >= 2 && isSeparatorRow(rows[1]) {
		if len(rows) > 2 {
			return rows[0], rows[2:]
		}
		return rows[0], nil
	}
	return nil, rows
}

func renderTable(theme Theme, headers []string, rows [][]string, width int) string {
	table := liptable.New().Wrap(true)
	if width > 0 {
		table.Width(width)
	}

	cellStyle := lipgloss.NewStyle().Padding(0, 1)
	headerStyle := cellStyle
	if theme.Styled {
		headerStyle = theme.headerStyle.Copy().Padding(0, 1)
	}

	table.StyleFunc(func(row, col int) lipgloss.Style {
		if row == liptable.HeaderRow {
			return headerStyle
		}
		return cellStyle
	})

	table.Border(lipgloss.NormalBorder()).
		BorderLeft(false).
		BorderRight(false).
		BorderTop(false).
		BorderBottom(false).
		BorderRow(false).
		BorderColumn(false)

	if len(headers) > 0 {
		table.Headers(headers...).BorderHeader(true)
		if theme.Styled {
			table.BorderStyle(theme.separatorStyle)
		}
	}
	if len(rows) > 0 {
		table.Rows(rows...)
	}
	return table.Render()
}
