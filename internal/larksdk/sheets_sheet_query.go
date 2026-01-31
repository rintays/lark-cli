package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larksheets "github.com/larksuite/oapi-sdk-go/v3/service/sheets/v3"
)

// ListSpreadsheetSheets queries the sheets within a spreadsheet.
func (c *Client) ListSpreadsheetSheets(ctx context.Context, token, spreadsheetToken string) ([]SpreadsheetSheet, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return nil, errors.New("spreadsheet token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}

	req := larksheets.NewQuerySpreadsheetSheetReqBuilder().SpreadsheetToken(spreadsheetToken).Build()
	resp, err := c.sdk.Sheets.V3.SpreadsheetSheet.Query(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
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
		out = append(out, row)
	}
	return out, nil
}
