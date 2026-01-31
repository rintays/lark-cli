package main

import (
	"strings"

	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

func docxBlocksText(blocks []*larkdocx.Block) string {
	if len(blocks) == 0 {
		return ""
	}
	lines := make([]string, 0, len(blocks))
	for _, block := range blocks {
		text := strings.TrimSpace(docxBlockText(block))
		if text != "" {
			lines = append(lines, text)
		}
	}
	return strings.Join(lines, "\n")
}

func docxBlockText(block *larkdocx.Block) string {
	if block == nil {
		return ""
	}
	if block.Text != nil {
		return docxTextValue(block.Text)
	}
	if block.Heading1 != nil {
		return docxTextValue(block.Heading1)
	}
	if block.Heading2 != nil {
		return docxTextValue(block.Heading2)
	}
	if block.Heading3 != nil {
		return docxTextValue(block.Heading3)
	}
	if block.Heading4 != nil {
		return docxTextValue(block.Heading4)
	}
	if block.Heading5 != nil {
		return docxTextValue(block.Heading5)
	}
	if block.Heading6 != nil {
		return docxTextValue(block.Heading6)
	}
	if block.Heading7 != nil {
		return docxTextValue(block.Heading7)
	}
	if block.Heading8 != nil {
		return docxTextValue(block.Heading8)
	}
	if block.Heading9 != nil {
		return docxTextValue(block.Heading9)
	}
	if block.Bullet != nil {
		return docxTextValue(block.Bullet)
	}
	if block.Ordered != nil {
		return docxTextValue(block.Ordered)
	}
	if block.Code != nil {
		return docxTextValue(block.Code)
	}
	if block.Quote != nil {
		return docxTextValue(block.Quote)
	}
	if block.Todo != nil {
		return docxTextValue(block.Todo)
	}
	if block.Page != nil {
		return docxTextValue(block.Page)
	}
	return ""
}

func docxTextValue(text *larkdocx.Text) string {
	if text == nil || len(text.Elements) == 0 {
		return ""
	}
	var b strings.Builder
	for _, el := range text.Elements {
		if el == nil {
			continue
		}
		switch {
		case el.TextRun != nil:
			if el.TextRun.Content != nil {
				b.WriteString(*el.TextRun.Content)
			}
		case el.MentionUser != nil:
			if el.MentionUser.UserId != nil && *el.MentionUser.UserId != "" {
				b.WriteString("@")
				b.WriteString(*el.MentionUser.UserId)
			} else {
				b.WriteString("@user")
			}
		case el.MentionDoc != nil:
			if el.MentionDoc.Title != nil && *el.MentionDoc.Title != "" {
				b.WriteString(*el.MentionDoc.Title)
			} else if el.MentionDoc.Url != nil && *el.MentionDoc.Url != "" {
				b.WriteString(*el.MentionDoc.Url)
			} else if el.MentionDoc.Token != nil && *el.MentionDoc.Token != "" {
				b.WriteString(*el.MentionDoc.Token)
			} else {
				b.WriteString("@doc")
			}
		case el.Reminder != nil:
			if el.Reminder.ExpireTime != nil && *el.Reminder.ExpireTime != "" {
				b.WriteString("[reminder:")
				b.WriteString(*el.Reminder.ExpireTime)
				b.WriteString("]")
			} else {
				b.WriteString("[reminder]")
			}
		case el.File != nil:
			if el.File.FileToken != nil && *el.File.FileToken != "" {
				b.WriteString("[file:")
				b.WriteString(*el.File.FileToken)
				b.WriteString("]")
			} else {
				b.WriteString("[file]")
			}
		case el.InlineBlock != nil:
			b.WriteString("[block]")
		case el.Equation != nil:
			if el.Equation.Content != nil && *el.Equation.Content != "" {
				b.WriteString(*el.Equation.Content)
			} else {
				b.WriteString("[equation]")
			}
		case el.LinkPreview != nil:
			if el.LinkPreview.Title != nil && *el.LinkPreview.Title != "" {
				b.WriteString(*el.LinkPreview.Title)
			} else if el.LinkPreview.Url != nil && *el.LinkPreview.Url != "" {
				b.WriteString(*el.LinkPreview.Url)
			} else {
				b.WriteString("[link]")
			}
		}
	}
	return b.String()
}
