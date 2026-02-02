package larksdk

import (
	"context"
	"errors"
	"fmt"

	larksheets "github.com/larksuite/oapi-sdk-go/v3/service/sheets/v3"
)

// ListSpreadsheetSheets queries the sheets within a spreadsheet.
func (c *Client) ListSpreadsheetSheets(ctx context.Context, token string, tokenType AccessTokenType, spreadsheetToken string) ([]SpreadsheetSheet, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return nil, errors.New("spreadsheet token is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return nil, err
	}

	req := larksheets.NewQuerySpreadsheetSheetReqBuilder().SpreadsheetToken(spreadsheetToken).Build()
	resp, err := c.sdk.Sheets.V3.SpreadsheetSheet.Query(ctx, req, option)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("query spreadsheet sheets failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("query spreadsheet sheets failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Sheets == nil {
		return nil, nil
	}
	out := make([]SpreadsheetSheet, 0, len(resp.Data.Sheets))
	for _, s := range resp.Data.Sheets {
		if s == nil {
			continue
		}
		row := SpreadsheetSheet{}
		if s.SheetId != nil {
			row.SheetID = *s.SheetId
		}
		if s.Title != nil {
			row.Title = *s.Title
		}
		if s.Index != nil {
			row.Index = int(*s.Index)
		}
		if s.Hidden != nil {
			row.Hidden = *s.Hidden
		}
		if s.ResourceType != nil {
			row.ResourceType = *s.ResourceType
		}
		if s.GridProperties != nil {
			props := &SpreadsheetGridProperties{}
			if s.GridProperties.FrozenRowCount != nil {
				props.FrozenRowCount = *s.GridProperties.FrozenRowCount
			}
			if s.GridProperties.FrozenColumnCount != nil {
				props.FrozenColumnCount = *s.GridProperties.FrozenColumnCount
			}
			if s.GridProperties.RowCount != nil {
				props.RowCount = *s.GridProperties.RowCount
			}
			if s.GridProperties.ColumnCount != nil {
				props.ColumnCount = *s.GridProperties.ColumnCount
			}
			row.GridProperties = props
		}
		if len(s.Merges) > 0 {
			merges := make([]SpreadsheetMergeRange, 0, len(s.Merges))
			for _, merge := range s.Merges {
				if merge == nil {
					continue
				}
				item := SpreadsheetMergeRange{}
				if merge.StartRowIndex != nil {
					item.StartRowIndex = *merge.StartRowIndex
				}
				if merge.EndRowIndex != nil {
					item.EndRowIndex = *merge.EndRowIndex
				}
				if merge.StartColumnIndex != nil {
					item.StartColumnIndex = *merge.StartColumnIndex
				}
				if merge.EndColumnIndex != nil {
					item.EndColumnIndex = *merge.EndColumnIndex
				}
				merges = append(merges, item)
			}
			row.Merges = merges
		}
		out = append(out, row)
	}
	return out, nil
}
