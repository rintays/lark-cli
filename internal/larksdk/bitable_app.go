package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type BitableApp struct {
	AppToken string `json:"app_token"`
	Name     string `json:"name"`
	URL      string `json:"url"`
}

type createBitableAppResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createBitableAppResponseData `json:"data"`
}

type createBitableAppResponseData struct {
	App *BitableApp `json:"app"`
}

func (r *createBitableAppResponse) Success() bool { return r.Code == 0 }

func (c *Client) CreateBitableApp(ctx context.Context, token string, name string, folderToken string) (BitableApp, error) {
	if !c.available() || c.coreConfig == nil {
		return BitableApp{}, ErrUnavailable
	}
	if name == "" {
		return BitableApp{}, errors.New("bitable app name is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BitableApp{}, errors.New("tenant access token is required")
	}

	payload := map[string]any{"name": name}
	if folderToken != "" {
		payload["folder_token"] = folderToken
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BitableApp{}, err
	}
	if apiResp == nil {
		return BitableApp{}, errors.New("create bitable app failed: empty response")
	}
	resp := &createBitableAppResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BitableApp{}, err
	}
	if !resp.Success() {
		return BitableApp{}, fmt.Errorf("create bitable app failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.App == nil {
		return BitableApp{}, errors.New("create bitable app failed: missing app")
	}
	if resp.Data.App.AppToken == "" {
		return BitableApp{}, errors.New("create bitable app failed: missing app_token")
	}
	return *resp.Data.App, nil
}
