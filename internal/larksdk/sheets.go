package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larksheets "github.com/larksuite/oapi-sdk-go/v3/service/sheets/v3"
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
	Data *SheetValueUpdate `json:"data"`
}

func (r *updateSheetRangeResponse) Success() bool {
	return r.Code == 0
}

type appendSheetRangeResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *SheetValueAppend `json:"data"`
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

type insertSheetRowsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *insertSheetRowsResponse) Success() bool {
	return r.Code == 0
}

type insertSheetColsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *insertSheetColsResponse) Success() bool {
	return r.Code == 0
}

type deleteSheetRowsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *deleteSheetRowsResponse) Success() bool {
	return r.Code == 0
}

type deleteSheetColsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *deleteSheetColsResponse) Success() bool {
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

func (c *Client) UpdateSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any) (SheetValueUpdate, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetValueUpdate{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetValueUpdate{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueUpdate{}, errors.New("range is required")
	}
	if len(values) == 0 {
		return SheetValueUpdate{}, errors.New("values are required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetValueUpdate{}, errors.New("tenant access token is required")
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
		"valueRange": SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetValueUpdate{}, err
	}
	if apiResp == nil {
		return SheetValueUpdate{}, errors.New("update sheet range failed: empty response")
	}
	resp := &updateSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetValueUpdate{}, err
	}
	if !resp.Success() {
		return SheetValueUpdate{}, fmt.Errorf("update sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return SheetValueUpdate{}, nil
	}
	return *resp.Data, nil
}

func (c *Client) AppendSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any, insertDataOption string) (SheetValueAppend, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetValueAppend{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetValueAppend{}, errors.New("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueAppend{}, errors.New("range is required")
	}
	if len(values) == 0 {
		return SheetValueAppend{}, errors.New("values are required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetValueAppend{}, errors.New("tenant access token is required")
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
		"valueRange": SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetValueAppend{}, err
	}
	if apiResp == nil {
		return SheetValueAppend{}, errors.New("append sheet range failed: empty response")
	}
	resp := &appendSheetRangeResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetValueAppend{}, err
	}
	if !resp.Success() {
		return SheetValueAppend{}, fmt.Errorf("append sheet range failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return SheetValueAppend{}, nil
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

func (c *Client) InsertSheetRows(ctx context.Context, token, spreadsheetToken, sheetID string, startIndex, count int) (SheetDimensionInsertResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetDimensionInsertResult{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetDimensionInsertResult{}, errors.New("spreadsheet token is required")
	}
	if sheetID == "" {
		return SheetDimensionInsertResult{}, errors.New("sheet id is required")
	}
	if startIndex < 0 {
		return SheetDimensionInsertResult{}, errors.New("start index must be >= 0")
	}
	if count <= 0 {
		return SheetDimensionInsertResult{}, errors.New("count must be greater than 0")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetDimensionInsertResult{}, errors.New("tenant access token is required")
	}
	endIndex := startIndex + count

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.PathParams.Set("sheet_id", sheetID)
	req.Body = map[string]any{
		"dimension_range": map[string]any{
			"major_dimension": "ROWS",
			"start_index":     startIndex,
			"end_index":       endIndex,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetDimensionInsertResult{}, err
	}
	if apiResp == nil {
		return SheetDimensionInsertResult{}, errors.New("insert sheet rows failed: empty response")
	}
	resp := &insertSheetRowsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetDimensionInsertResult{}, err
	}
	if !resp.Success() {
		return SheetDimensionInsertResult{}, fmt.Errorf("insert sheet rows failed: %s", resp.Msg)
	}
	return SheetDimensionInsertResult{StartIndex: startIndex, Count: count, EndIndex: endIndex}, nil
}

func (c *Client) DeleteSheetRows(ctx context.Context, token, spreadsheetToken, sheetID string, startIndex, count int) (SheetDimensionDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetDimensionDeleteResult{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetDimensionDeleteResult{}, errors.New("spreadsheet token is required")
	}
	if sheetID == "" {
		return SheetDimensionDeleteResult{}, errors.New("sheet id is required")
	}
	if startIndex < 0 {
		return SheetDimensionDeleteResult{}, errors.New("start index must be >= 0")
	}
	if count <= 0 {
		return SheetDimensionDeleteResult{}, errors.New("count must be greater than 0")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetDimensionDeleteResult{}, errors.New("tenant access token is required")
	}
	endIndex := startIndex + count

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/delete_dimension",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.PathParams.Set("sheet_id", sheetID)
	req.Body = map[string]any{
		"dimension_range": map[string]any{
			"major_dimension": "ROWS",
			"start_index":     startIndex,
			"end_index":       endIndex,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetDimensionDeleteResult{}, err
	}
	if apiResp == nil {
		return SheetDimensionDeleteResult{}, errors.New("delete sheet rows failed: empty response")
	}
	resp := &deleteSheetRowsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetDimensionDeleteResult{}, err
	}
	if !resp.Success() {
		return SheetDimensionDeleteResult{}, fmt.Errorf("delete sheet rows failed: %s", resp.Msg)
	}
	return SheetDimensionDeleteResult{StartIndex: startIndex, Count: count, EndIndex: endIndex}, nil
}

func (c *Client) DeleteSheetCols(ctx context.Context, token, spreadsheetToken, sheetID string, startIndex, count int) (SheetDimensionDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetDimensionDeleteResult{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetDimensionDeleteResult{}, errors.New("spreadsheet token is required")
	}
	if sheetID == "" {
		return SheetDimensionDeleteResult{}, errors.New("sheet id is required")
	}
	if startIndex < 0 {
		return SheetDimensionDeleteResult{}, errors.New("start index must be >= 0")
	}
	if count <= 0 {
		return SheetDimensionDeleteResult{}, errors.New("count must be greater than 0")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetDimensionDeleteResult{}, errors.New("tenant access token is required")
	}
	endIndex := startIndex + count

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/delete_dimension",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.PathParams.Set("sheet_id", sheetID)
	req.Body = map[string]any{
		"dimension_range": map[string]any{
			"major_dimension": "COLS",
			"start_index":     startIndex,
			"end_index":       endIndex,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetDimensionDeleteResult{}, err
	}
	if apiResp == nil {
		return SheetDimensionDeleteResult{}, errors.New("delete sheet cols failed: empty response")
	}
	resp := &deleteSheetColsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetDimensionDeleteResult{}, err
	}
	if !resp.Success() {
		return SheetDimensionDeleteResult{}, fmt.Errorf("delete sheet cols failed: %s", resp.Msg)
	}
	return SheetDimensionDeleteResult{StartIndex: startIndex, Count: count, EndIndex: endIndex}, nil
}

func (c *Client) InsertSheetCols(ctx context.Context, token, spreadsheetToken, sheetID string, startIndex, count int) (SheetDimensionInsertResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SheetDimensionInsertResult{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SheetDimensionInsertResult{}, errors.New("spreadsheet token is required")
	}
	if sheetID == "" {
		return SheetDimensionInsertResult{}, errors.New("sheet id is required")
	}
	if startIndex < 0 {
		return SheetDimensionInsertResult{}, errors.New("start index must be >= 0")
	}
	if count <= 0 {
		return SheetDimensionInsertResult{}, errors.New("count must be greater than 0")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SheetDimensionInsertResult{}, errors.New("tenant access token is required")
	}
	endIndex := startIndex + count

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("spreadsheet_token", spreadsheetToken)
	req.PathParams.Set("sheet_id", sheetID)
	req.Body = map[string]any{
		"dimension_range": map[string]any{
			"major_dimension": "COLS",
			"start_index":     startIndex,
			"end_index":       endIndex,
		},
	}

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SheetDimensionInsertResult{}, err
	}
	if apiResp == nil {
		return SheetDimensionInsertResult{}, errors.New("insert sheet cols failed: empty response")
	}
	resp := &insertSheetColsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SheetDimensionInsertResult{}, err
	}
	if !resp.Success() {
		return SheetDimensionInsertResult{}, fmt.Errorf("insert sheet cols failed: %s", resp.Msg)
	}
	return SheetDimensionInsertResult{StartIndex: startIndex, Count: count, EndIndex: endIndex}, nil
}

func (c *Client) GetSpreadsheetMetadata(ctx context.Context, token, spreadsheetToken string) (SpreadsheetMetadata, error) {
	if !c.available() {
		return SpreadsheetMetadata{}, ErrUnavailable
	}
	if spreadsheetToken == "" {
		return SpreadsheetMetadata{}, errors.New("spreadsheet token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SpreadsheetMetadata{}, errors.New("tenant access token is required")
	}

	req := larksheets.NewGetSpreadsheetReqBuilder().
		SpreadsheetToken(spreadsheetToken).
		Build()
	resp, err := c.sdk.Sheets.V3.Spreadsheet.Get(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SpreadsheetMetadata{}, err
	}
	if resp == nil {
		return SpreadsheetMetadata{}, errors.New("get spreadsheet metadata failed: empty response")
	}
	if !resp.Success() {
		return SpreadsheetMetadata{}, fmt.Errorf("get spreadsheet metadata failed: %s", resp.Msg)
	}
	metadata := SpreadsheetMetadata{}
	if resp.Data == nil || resp.Data.Spreadsheet == nil {
		return metadata, nil
	}
	if resp.Data.Spreadsheet.Title != nil {
		metadata.Properties.Title = *resp.Data.Spreadsheet.Title
	}
	return metadata, nil
}
