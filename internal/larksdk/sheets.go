package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"lark/internal/larkapi"
)

type readSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *readSheetRangeData `json:"data"`
}

type readSheetRangeData struct {
	ValueRange *larkapi.SheetValueRange `json:"valueRange"`
}

func (r *readSheetRangeResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) ReadSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (larkapi.SheetValueRange, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.SheetValueRange{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return larkapi.SheetValueRange{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return larkapi.SheetValueRange{}, errors.New("range is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.SheetValueRange{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values/:range",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.PathParams.Set("range", sheetRange)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.SheetValueRange{}, err
	}
	if apiResp == nil {
		return larkapi.SheetValueRange{}, errors.New("read sheet range failed: empty response")
	}
	resp := &readSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.SheetValueRange{}, err
	}
	if !resp.Success() {
		return larkapi.SheetValueRange{}, fmt.Errorf("read sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.ValueRange == nil {
		return larkapi.SheetValueRange{}, nil
	}
	return *resp.Data.ValueRange, nil
}
