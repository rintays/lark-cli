package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type whoamiResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *whoamiResponseData `json:"data"`
}

type whoamiResponseData struct {
	Tenant *TenantInfo `json:"tenant"`
}

func (r *whoamiResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) WhoAmI(ctx context.Context, token string) (TenantInfo, error) {
	if !c.available() || c.coreConfig == nil {
		return TenantInfo{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return TenantInfo{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/tenant/v2/tenant/query",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return TenantInfo{}, err
	}
	if apiResp == nil {
		return TenantInfo{}, errors.New("whoami failed: empty response")
	}
	resp := &whoamiResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return TenantInfo{}, err
	}
	if !resp.Success() {
		return TenantInfo{}, fmt.Errorf("whoami failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Tenant == nil {
		return TenantInfo{}, errors.New("whoami response missing tenant")
	}
	return *resp.Data.Tenant, nil
}
