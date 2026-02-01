package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type deleteBaseViewResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *deleteBaseViewResponseData `json:"data"`
}

type deleteBaseViewResponseData struct {
	ViewID  string `json:"view_id"`
	Deleted bool   `json:"deleted"`
}

func (r *deleteBaseViewResponse) Success() bool { return r.Code == 0 }

func (c *Client) DeleteBaseView(ctx context.Context, token, appToken, tableID, viewID string) (BaseViewDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseViewDeleteResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseViewDeleteResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseViewDeleteResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseViewDeleteResult{}, errors.New("table id is required")
	}
	if viewID == "" {
		return BaseViewDeleteResult{}, errors.New("view id is required")
	}

	if c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableView != nil {
		req := larkbitable.NewDeleteAppTableViewReqBuilder().AppToken(appToken).TableId(tableID).ViewId(viewID).Build()
		resp, err := c.sdk.Bitable.V1.AppTableView.Delete(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
		if err != nil {
			return BaseViewDeleteResult{}, err
		}
		if resp == nil {
			return BaseViewDeleteResult{}, errors.New("delete base view failed: empty response")
		}
		if !resp.Success() {
			return BaseViewDeleteResult{}, formatCodeError("delete base view failed", resp.CodeError, resp.ApiResp)
		}
		return BaseViewDeleteResult{ViewID: viewID, Deleted: true}, nil
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/views/:view_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("view_id", viewID)
	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseViewDeleteResult{}, err
	}
	if apiResp == nil {
		return BaseViewDeleteResult{}, errors.New("delete base view failed: empty response")
	}
	resp := &deleteBaseViewResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseViewDeleteResult{}, err
	}
	if !resp.Success() {
		return BaseViewDeleteResult{}, formatCodeError("delete base view failed", resp.CodeError, resp.ApiResp)
	}
	result := BaseViewDeleteResult{ViewID: viewID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.ViewID != "" {
			result.ViewID = resp.Data.ViewID
		}
		result.Deleted = resp.Data.Deleted
	}
	return result, nil
}
