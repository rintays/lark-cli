package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type listBaseTablesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listBaseTablesResponseData `json:"data"`
}

type listBaseTablesResponseData struct {
	Items     []BaseTable `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

func (r *listBaseTablesResponse) Success() bool { return r.Code == 0 }

func (c *Client) ListBaseTables(ctx context.Context, token, appToken string) (ListBaseTablesResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListBaseTablesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListBaseTablesResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return ListBaseTablesResult{}, errors.New("app token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListBaseTablesResult{}, err
	}
	if apiResp == nil {
		return ListBaseTablesResult{}, errors.New("list base tables failed: empty response")
	}
	resp := &listBaseTablesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListBaseTablesResult{}, err
	}
	if !resp.Success() {
		return ListBaseTablesResult{}, fmt.Errorf("list base tables failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListBaseTablesResult{}, nil
	}
	return ListBaseTablesResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}

type listBaseFieldsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listBaseFieldsResponseData `json:"data"`
}

type listBaseFieldsResponseData struct {
	Items     []BaseField `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

func (r *listBaseFieldsResponse) Success() bool { return r.Code == 0 }

func (c *Client) ListBaseFields(ctx context.Context, token, appToken, tableID string) (ListBaseFieldsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListBaseFieldsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListBaseFieldsResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return ListBaseFieldsResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return ListBaseFieldsResult{}, errors.New("table id is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListBaseFieldsResult{}, err
	}
	if apiResp == nil {
		return ListBaseFieldsResult{}, errors.New("list base fields failed: empty response")
	}
	resp := &listBaseFieldsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListBaseFieldsResult{}, err
	}
	if !resp.Success() {
		return ListBaseFieldsResult{}, fmt.Errorf("list base fields failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListBaseFieldsResult{}, nil
	}
	return ListBaseFieldsResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}
