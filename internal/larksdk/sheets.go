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
	ValueRange *SheetValueRange `json:"valueRange"`
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

type appendSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *larkapi.SheetValueAppend `json:"data"`
}

func (r *appendSheetRangeResponse) Success() bool {
	return r.Code == 0
}

type clearSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *clearSheetRangeData `json:"data"`
}

type clearSheetRangeData struct {
	ClearedRange       string `json:"clearedRange"`
	ClearedRangeLegacy string `json:"cleared_range"`
}

func (r *clearSheetRangeResponse) Success() bool {
	return r.Code == 0
}

type spreadsheetMetadataResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *larkapi.SpreadsheetMetadata `json:"data"`
}

func (r *spreadsheetMetadataResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) ReadSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (SheetValueRange, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetValueRange{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetValueRange{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueRange{}, errors.New("range is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetValueRange{}, errors.New("tenant access token is required")
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
		return SheetValueRange{}, err
	}
	if apiResp == nil {
		return SheetValueRange{}, errors.New("read sheet range failed: empty response")
	}
	resp := &readSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetValueRange{}, err
	}
	if !resp.Success() {
		return SheetValueRange{}, fmt.Errorf("read sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.ValueRange == nil {
		return SheetValueRange{}, nil
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

func (c *Client) AppendSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any, insertDataOption string) (larkapi.SheetValueAppend, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.SheetValueAppend{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return larkapi.SheetValueAppend{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return larkapi.SheetValueAppend{}, errors.New("range is required")
	}
	if len(values) == 0 {
		return larkapi.SheetValueAppend{}, errors.New("values are required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.SheetValueAppend{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values_append",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.QueryParams.Set("valueInputOption", "RAW")
	if insertDataOption != "" {
		req.QueryParams.Set("insertDataOption", insertDataOption)
	}
	req.Body = map[string]any{
		"valueRange": larkapi.SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.SheetValueAppend{}, err
	}
	if apiResp == nil {
		return larkapi.SheetValueAppend{}, errors.New("append sheet range failed: empty response")
	}
	resp := &appendSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.SheetValueAppend{}, err
	}
	if !resp.Success() {
		return larkapi.SheetValueAppend{}, fmt.Errorf("append sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return larkapi.SheetValueAppend{}, nil
	}
	return *resp.Data, nil
}

func (c *Client) ClearSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (ClearSheetRangeResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ClearSheetRangeResult{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return ClearSheetRangeResult{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return ClearSheetRangeResult{}, errors.New("range is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ClearSheetRangeResult{}, errors.New("tenant access token is required")
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
		return ClearSheetRangeResult{}, err
	}
	if apiResp == nil {
		return ClearSheetRangeResult{}, errors.New("clear sheet range failed: empty response")
	}
	resp := &clearSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ClearSheetRangeResult{}, err
	}
	if !resp.Success() {
		return ClearSheetRangeResult{}, fmt.Errorf("clear sheet range failed: %s", resp.Msg)
	}
	result := ClearSheetRangeResult{ClearedRange: sheetRange}
	if resp.Data != nil {
		if resp.Data.ClearedRange != "" {
			result.ClearedRange = resp.Data.ClearedRange
		} else if resp.Data.ClearedRangeLegacy != "" {
			result.ClearedRange = resp.Data.ClearedRangeLegacy
		}
	}
	return result, nil
}

func (c *Client) GetSpreadsheetMetadata(ctx context.Context, token, spreadsheetToken string) (larkapi.SpreadsheetMetadata, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.SpreadsheetMetadata{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return larkapi.SpreadsheetMetadata{}, errors.New("spreadsheet token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.SpreadsheetMetadata{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/metainfo",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.SpreadsheetMetadata{}, err
	}
	if apiResp == nil {
		return larkapi.SpreadsheetMetadata{}, errors.New("get spreadsheet metadata failed: empty response")
	}
	resp := &spreadsheetMetadataResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.SpreadsheetMetadata{}, err
	}
	if !resp.Success() {
		return larkapi.SpreadsheetMetadata{}, fmt.Errorf("get spreadsheet metadata failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return larkapi.SpreadsheetMetadata{}, nil
	}
	return *resp.Data, nil
}
