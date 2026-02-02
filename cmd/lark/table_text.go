package main

import (
	"strings"

	"lark/internal/output"
)

var tablePlain bool

func tableText(headers []string, lines []string, emptyText string) string {
	if len(lines) == 0 {
		if strings.TrimSpace(emptyText) == "" {
			return emptyText
		}
		return output.Notice(output.NoticeInfo, emptyText, nil)
	}
	if tablePlain {
		return output.TableTSVFromLines(headers, lines)
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
	if tablePlain {
		return output.TableTSV(headers, rows)
	}
	return output.TableText(headers, rows)
}

func tableTextRow(headers []string, row []string) string {
	if tablePlain {
		return output.TableTSV(headers, [][]string{row})
	}
	return output.TableText(headers, [][]string{row})
}
