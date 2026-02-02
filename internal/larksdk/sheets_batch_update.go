package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type batchUpdateSpreadsheetResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data,omitempty"`
}

func (r *batchUpdateSpreadsheetResponse) Success() bool { return r.Code == 0 }

func (c *Client) UpdateSpreadsheetSheetTitle(ctx context.Context, token string, tokenType AccessTokenType, spreadsheetToken, sheetID, title string) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	if spreadsheetToken == "" {
		return errors.New("spreadsheet token is required")
	}
	if sheetID == "" {
		return errors.New("sheet id is required")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return errors.New("sheet title is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return err
	}
	payload := map[string]any{
		"requests": []map[string]any{
			{
				"updateSheet": map[string]any{
					"properties": map[string]any{
						"sheetId": sheetID,
						"title":   title,
					},
				},
			},
		},
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/sheets_batch_update",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, option)
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("update sheet title failed: empty response")
	}
	resp := &batchUpdateSpreadsheetResponse{}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("update sheet title failed: %s", resp.Msg)
	}
	return nil
}
