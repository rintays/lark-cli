package main

import (
	"fmt"
	"net/url"
	"strconv"
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

type docxBlockIndex struct {
	blocks map[string]*larkdocx.Block
	order  []string
}

func newDocxBlockIndex(blocks []*larkdocx.Block) *docxBlockIndex {
	idx := &docxBlockIndex{
		blocks: make(map[string]*larkdocx.Block, len(blocks)),
		order:  make([]string, 0, len(blocks)),
	}
	for _, block := range blocks {
		id := docxBlockID(block)
		if id == "" {
			continue
		}
		if _, exists := idx.blocks[id]; !exists {
			idx.order = append(idx.order, id)
		}
		idx.blocks[id] = block
	}
	return idx
}

func docxBlocksMarkdown(documentID string, blocks []*larkdocx.Block) string {
	if len(blocks) == 0 {
		return ""
	}
	idx := newDocxBlockIndex(blocks)
	start, recursive := idx.startIDs(documentID)
	lines := make([]string, 0, len(blocks))
	if recursive {
		visited := map[string]bool{}
		for _, id := range start {
			idx.renderMarkdown(id, &lines, visited)
		}
	} else {
		for _, id := range start {
			block := idx.blocks[id]
			if block == nil {
				continue
			}
			if md, _ := docxBlockMarkdown(idx, block); md != "" {
				lines = append(lines, strings.TrimRight(md, "\n"))
			}
		}
	}
	return strings.Join(lines, "\n")
}

func (idx *docxBlockIndex) startIDs(documentID string) ([]string, bool) {
	if documentID != "" {
		if root, ok := idx.blocks[documentID]; ok && len(root.Children) > 0 {
			return root.Children, true
		}
	}
	if documentID != "" {
		ids := make([]string, 0)
		for _, id := range idx.order {
			block := idx.blocks[id]
			if block == nil || block.ParentId == nil {
				continue
			}
			if *block.ParentId == documentID {
				ids = append(ids, id)
			}
		}
		if len(ids) > 0 {
			return ids, true
		}
	}
	return idx.order, false
}

func (idx *docxBlockIndex) renderMarkdown(id string, lines *[]string, visited map[string]bool) {
	if id == "" || visited[id] {
		return
	}
	visited[id] = true
	block := idx.blocks[id]
	if block == nil {
		return
	}
	md, skipChildren := docxBlockMarkdown(idx, block)
	if md != "" {
		*lines = append(*lines, strings.TrimRight(md, "\n"))
	}
	if skipChildren {
		return
	}
	for _, childID := range block.Children {
		idx.renderMarkdown(childID, lines, visited)
	}
}

func docxBlockMarkdown(idx *docxBlockIndex, block *larkdocx.Block) (string, bool) {
	if block == nil {
		return "", true
	}
	switch {
	case block.Heading1 != nil:
		return docxHeadingMarkdown(1, block.Heading1), true
	case block.Heading2 != nil:
		return docxHeadingMarkdown(2, block.Heading2), true
	case block.Heading3 != nil:
		return docxHeadingMarkdown(3, block.Heading3), true
	case block.Heading4 != nil:
		return docxHeadingMarkdown(4, block.Heading4), true
	case block.Heading5 != nil:
		return docxHeadingMarkdown(5, block.Heading5), true
	case block.Heading6 != nil:
		return docxHeadingMarkdown(6, block.Heading6), true
	case block.Heading7 != nil:
		return docxHeadingMarkdown(7, block.Heading7), true
	case block.Heading8 != nil:
		return docxHeadingMarkdown(8, block.Heading8), true
	case block.Heading9 != nil:
		return docxHeadingMarkdown(9, block.Heading9), true
	case block.Text != nil:
		return strings.TrimSpace(docxTextMarkdown(block.Text)), true
	case block.Bullet != nil:
		return fmt.Sprintf("%s- %s", docxIndentPrefix(block.Bullet), docxTextMarkdown(block.Bullet)), true
	case block.Ordered != nil:
		return fmt.Sprintf("%s%s %s", docxIndentPrefix(block.Ordered), docxOrderedMarker(block.Ordered), docxTextMarkdown(block.Ordered)), true
	case block.Todo != nil:
		return fmt.Sprintf("%s%s %s", docxIndentPrefix(block.Todo), docxTodoMarker(block.Todo), docxTextMarkdown(block.Todo)), true
	case block.Quote != nil:
		return docxQuoteMarkdown(block.Quote), true
	case block.Code != nil:
		return docxCodeMarkdown(block.Code), true
	case block.Divider != nil:
		return "---", true
	case block.Image != nil:
		return docxImageMarkdown(block.Image), true
	case block.File != nil:
		return docxFileMarkdown(block.File), true
	case block.Table != nil:
		return docxTableMarkdown(idx, block.Table), true
	case block.TableCell != nil:
		return "", false
	case block.Page != nil || block.Grid != nil || block.GridColumn != nil || block.QuoteContainer != nil || block.Callout != nil || block.View != nil:
		return "", false
	}
	if text := strings.TrimSpace(docxBlockText(block)); text != "" {
		return text, true
	}
	return "", false
}

func docxHeadingMarkdown(level int, text *larkdocx.Text) string {
	if level < 1 {
		level = 1
	}
	if level > 6 {
		level = 6
	}
	return fmt.Sprintf("%s %s", strings.Repeat("#", level), docxTextMarkdown(text))
}

func docxQuoteMarkdown(text *larkdocx.Text) string {
	content := docxTextMarkdown(text)
	if content == "" {
		return ""
	}
	parts := strings.Split(content, "\n")
	for i, line := range parts {
		parts[i] = "> " + line
	}
	return strings.Join(parts, "\n")
}

func docxCodeMarkdown(text *larkdocx.Text) string {
	content := docxTextMarkdown(text)
	if content == "" {
		return "```\n```"
	}
	return "```\n" + content + "\n```"
}

func docxImageMarkdown(image *larkdocx.Image) string {
	if image == nil {
		return ""
	}
	if image.Token != nil && *image.Token != "" {
		return fmt.Sprintf("![image](%s)", *image.Token)
	}
	return "![image]"
}

func docxFileMarkdown(file *larkdocx.File) string {
	if file == nil {
		return ""
	}
	name := ""
	if file.Name != nil {
		name = strings.TrimSpace(*file.Name)
	}
	if name == "" && file.Token != nil {
		name = strings.TrimSpace(*file.Token)
	}
	if name == "" {
		name = "file"
	}
	return fmt.Sprintf("[%s]", name)
}

func docxTableMarkdown(idx *docxBlockIndex, table *larkdocx.Table) string {
	if table == nil || len(table.Cells) == 0 {
		return ""
	}
	rows, cols := docxTableSize(table)
	if rows <= 0 || cols <= 0 {
		return "[table]"
	}
	grid := make([][]string, rows)
	for r := 0; r < rows; r++ {
		grid[r] = make([]string, cols)
	}
	for i, cellID := range table.Cells {
		r := i / cols
		c := i % cols
		if r >= rows {
			break
		}
		grid[r][c] = docxTableCellText(idx, cellID)
	}
	lines := make([]string, 0, rows+2)
	lines = append(lines, "| "+strings.Join(grid[0], " | ")+" |")
	seps := make([]string, cols)
	for i := range seps {
		seps[i] = "---"
	}
	lines = append(lines, "| "+strings.Join(seps, " | ")+" |")
	for r := 1; r < rows; r++ {
		lines = append(lines, "| "+strings.Join(grid[r], " | ")+" |")
	}
	return strings.Join(lines, "\n")
}

func docxTableSize(table *larkdocx.Table) (int, int) {
	if table == nil || table.Property == nil {
		return 0, 0
	}
	rows := 0
	cols := 0
	if table.Property.RowSize != nil {
		rows = *table.Property.RowSize
	}
	if table.Property.ColumnSize != nil {
		cols = *table.Property.ColumnSize
	}
	if rows <= 0 || cols <= 0 {
		return 0, 0
	}
	return rows, cols
}

func docxTableCellText(idx *docxBlockIndex, cellID string) string {
	if idx == nil || cellID == "" {
		return ""
	}
	cell := idx.blocks[cellID]
	if cell == nil {
		return ""
	}
	if text := strings.TrimSpace(docxBlockText(cell)); text != "" {
		return text
	}
	if len(cell.Children) == 0 {
		return ""
	}
	parts := make([]string, 0, len(cell.Children))
	for _, childID := range cell.Children {
		child := idx.blocks[childID]
		if child == nil {
			continue
		}
		if text := strings.TrimSpace(docxBlockText(child)); text != "" {
			parts = append(parts, text)
		}
	}
	return strings.Join(parts, "<br>")
}

func docxTextMarkdown(text *larkdocx.Text) string {
	if text == nil || len(text.Elements) == 0 {
		return ""
	}
	var b strings.Builder
	for _, el := range text.Elements {
		if el == nil {
			continue
		}
		b.WriteString(docxTextElementMarkdown(el))
	}
	return b.String()
}

func docxTextElementMarkdown(el *larkdocx.TextElement) string {
	switch {
	case el.TextRun != nil:
		return docxApplyInlineStyle(valueOrEmpty(el.TextRun.Content), el.TextRun.TextElementStyle)
	case el.MentionUser != nil:
		content := "@user"
		if el.MentionUser.UserId != nil && *el.MentionUser.UserId != "" {
			content = "@" + *el.MentionUser.UserId
		}
		return docxApplyInlineStyle(content, el.MentionUser.TextElementStyle)
	case el.MentionDoc != nil:
		title := valueOrEmpty(el.MentionDoc.Title)
		if title == "" {
			title = valueOrEmpty(el.MentionDoc.Token)
		}
		if title == "" {
			title = "@doc"
		}
		if urlValue := docxDecodeURL(valueOrEmpty(el.MentionDoc.Url)); urlValue != "" {
			title = fmt.Sprintf("[%s](%s)", title, urlValue)
		}
		return docxApplyInlineStyle(title, el.MentionDoc.TextElementStyle)
	case el.Reminder != nil:
		if el.Reminder.ExpireTime != nil && *el.Reminder.ExpireTime != "" {
			return "[reminder:" + *el.Reminder.ExpireTime + "]"
		}
		return "[reminder]"
	case el.File != nil:
		if el.File.FileToken != nil && *el.File.FileToken != "" {
			return "[file:" + *el.File.FileToken + "]"
		}
		return "[file]"
	case el.InlineBlock != nil:
		return "[block]"
	case el.Equation != nil:
		if el.Equation.Content != nil && *el.Equation.Content != "" {
			return "$" + *el.Equation.Content + "$"
		}
		return "[equation]"
	case el.LinkPreview != nil:
		title := valueOrEmpty(el.LinkPreview.Title)
		if title == "" {
			title = valueOrEmpty(el.LinkPreview.Url)
		}
		if title == "" {
			return "[link]"
		}
		if urlValue := docxDecodeURL(valueOrEmpty(el.LinkPreview.Url)); urlValue != "" {
			return fmt.Sprintf("[%s](%s)", title, urlValue)
		}
		return title
	default:
		return ""
	}
}

func docxApplyInlineStyle(content string, style *larkdocx.TextElementStyle) string {
	if content == "" {
		return ""
	}
	if style == nil {
		return content
	}
	if style.InlineCode != nil && *style.InlineCode {
		return "`" + content + "`"
	}
	if style.Bold != nil && *style.Bold {
		content = "**" + content + "**"
	}
	if style.Italic != nil && *style.Italic {
		content = "*" + content + "*"
	}
	if style.Strikethrough != nil && *style.Strikethrough {
		content = "~~" + content + "~~"
	}
	if style.Underline != nil && *style.Underline {
		content = "<u>" + content + "</u>"
	}
	if style.Link != nil && style.Link.Url != nil && *style.Link.Url != "" {
		content = fmt.Sprintf("[%s](%s)", content, docxDecodeURL(*style.Link.Url))
	}
	return content
}

func docxIndentPrefix(text *larkdocx.Text) string {
	level := docxIndentationLevel(text)
	if level <= 0 {
		return ""
	}
	return strings.Repeat("  ", level)
}

func docxIndentationLevel(text *larkdocx.Text) int {
	if text == nil || text.Style == nil || text.Style.IndentationLevel == nil {
		return 0
	}
	raw := strings.TrimSpace(*text.Style.IndentationLevel)
	if raw == "" {
		return 0
	}
	level, err := strconv.Atoi(raw)
	if err != nil || level < 0 {
		return 0
	}
	return level
}

func docxOrderedMarker(text *larkdocx.Text) string {
	if text == nil || text.Style == nil || text.Style.Sequence == nil {
		return "1."
	}
	raw := strings.TrimSpace(*text.Style.Sequence)
	if raw == "" || raw == "auto" {
		return "1."
	}
	return raw + "."
}

func docxTodoMarker(text *larkdocx.Text) string {
	if text == nil || text.Style == nil || text.Style.Done == nil {
		return "- [ ]"
	}
	if *text.Style.Done {
		return "- [x]"
	}
	return "- [ ]"
}

func valueOrEmpty(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func docxDecodeURL(raw string) string {
	if raw == "" {
		return ""
	}
	decoded, err := url.QueryUnescape(raw)
	if err != nil {
		return raw
	}
	return decoded
}
