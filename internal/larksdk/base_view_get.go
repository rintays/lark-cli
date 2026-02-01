package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type getBaseViewResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getBaseViewResponseData `json:"data"`
}

type getBaseViewResponseData struct {
	View *BaseView `json:"view"`
}

func (r *getBaseViewResponse) Success() bool { return r.Code == 0 }

// GetBaseView returns a single view.
// SDK-first, with core fallback.
func (c *Client) GetBaseView(ctx context.Context, token, appToken, tableID, viewID string) (BaseView, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseView{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseView{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseView{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseView{}, errors.New("table id is required")
	}
	if viewID == "" {
		return BaseView{}, errors.New("view id is required")
	}

	if c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableView != nil {
		req := larkbitable.NewGetAppTableViewReqBuilder().AppToken(appToken).TableId(tableID).ViewId(viewID).Build()
		resp, err := c.sdk.Bitable.V1.AppTableView.Get(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
		if err == nil && resp != nil {
			if !resp.Success() {
				return BaseView{}, formatCodeError("get base view failed", resp.CodeError, resp.ApiResp)
			}
			if resp.Data == nil || resp.Data.View == nil {
				return BaseView{}, nil
			}
			return mapBaseViewFromSDK(resp.Data.View), nil
		}
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/views/:view_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("view_id", viewID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseView{}, err
	}
	if apiResp == nil {
		return BaseView{}, errors.New("get base view failed: empty response")
	}
	resp := &getBaseViewResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseView{}, err
	}
	if !resp.Success() {
		return BaseView{}, formatCodeError("get base view failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.View == nil {
		return BaseView{}, nil
	}
	return *resp.Data.View, nil
}
