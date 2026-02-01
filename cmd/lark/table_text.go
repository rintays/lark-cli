package main

import (
	"strings"

	"lark/internal/output"
)

func tableText(headers []string, lines []string, emptyText string) string {
	if len(lines) == 0 {
		if strings.TrimSpace(emptyText) == "" {
			return emptyText
		}
		return output.Notice(output.NoticeInfo, emptyText, nil)
	}
	return output.TableTextFromLines(headers, lines)
}

func tableTextFromRows(headers []string, rows [][]string, emptyText string) string {
	if len(rows) == 0 {
		if strings.TrimSpace(emptyText) == "" {
			return emptyText
		}
		return output.Notice(output.NoticeInfo, emptyText, nil)
	}
	return output.TableText(headers, rows)
}

func tableTextRow(headers []string, row []string) string {
	return output.TableText(headers, [][]string{row})
}
