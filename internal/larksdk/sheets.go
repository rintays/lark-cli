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

type updateSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *larkapi.SheetValueUpdate `json:"data"`
}

func (r *updateSheetRangeResponse) Success() bool {
	return r.Code == 0
}

type clearSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data map[string]any `json:"data"`
}

func (r *clearSheetRangeResponse) Success() bool {
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

func (c *Client) UpdateSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any) (larkapi.SheetValueUpdate, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.SheetValueUpdate{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return larkapi.SheetValueUpdate{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return larkapi.SheetValueUpdate{}, errors.New("range is required")
	}
	if len(values) == 0 {
		return larkapi.SheetValueUpdate{}, errors.New("values are required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.SheetValueUpdate{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.QueryParams.Set("valueInputOption", "RAW")
	req.Body = map[string]any{
		"valueRange": larkapi.SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.SheetValueUpdate{}, err
	}
	if apiResp == nil {
		return larkapi.SheetValueUpdate{}, errors.New("update sheet range failed: empty response")
	}
	resp := &updateSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.SheetValueUpdate{}, err
	}
	if !resp.Success() {
		return larkapi.SheetValueUpdate{}, fmt.Errorf("update sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return larkapi.SheetValueUpdate{}, nil
	}
	return *resp.Data, nil
}

func (c *Client) ClearSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (string, error) {
	if !c.available() || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	if spreadsheetToken == "" {
		return "", errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return "", errors.New("range is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values_clear",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.QueryParams.Set("range", sheetRange)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("clear sheet range failed: empty response")
	}
	resp := &clearSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("clear sheet range failed: %s", resp.Msg)
	}
	clearedRange := sheetRange
	if resp.Data != nil {
		if value, ok := resp.Data["clearedRange"].(string); ok && value != "" {
			clearedRange = value
		} else if value, ok := resp.Data["cleared_range"].(string); ok && value != "" {
			clearedRange = value
		}
	}
	return clearedRange, nil
}
