package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type BitableApp struct {
	AppToken    string `json:"app_token"`
	Name        string `json:"name"`
	URL         string `json:"url,omitempty"`
	FolderToken string `json:"folder_token,omitempty"`
	TimeZone    string `json:"time_zone,omitempty"`
	IsAdvanced  *bool  `json:"is_advanced,omitempty"`
}

type BitableAppCreateOptions struct {
	FolderToken      string
	TimeZone         string
	CustomizedConfig *bool
	SourceAppToken   string
	CopyTypes        []string
	ApiType          string
}

type BitableAppUpdateOptions struct {
	Name       string
	IsAdvanced *bool
}

type BitableAppCopyOptions struct {
	Name           string
	FolderToken    string
	WithoutContent *bool
	TimeZone       string
}

func (c *Client) bitableAppGetSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.App != nil
}

func mapBitableAppFromSDKDisplay(app *larkbitable.DisplayApp) BitableApp {
	var out BitableApp
	if app == nil {
		return out
	}
	if app.AppToken != nil {
		out.AppToken = *app.AppToken
	}
	if app.Name != nil {
		out.Name = *app.Name
	}
	if app.TimeZone != nil {
		out.TimeZone = *app.TimeZone
	}
	if app.IsAdvanced != nil {
		out.IsAdvanced = app.IsAdvanced
	}
	return out
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

type getBitableAppResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getBitableAppResponseData `json:"data"`
}

type getBitableAppResponseData struct {
	App *BitableApp `json:"app"`
}

func (r *getBitableAppResponse) Success() bool { return r.Code == 0 }

type updateBitableAppResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateBitableAppResponseData `json:"data"`
}

type updateBitableAppResponseData struct {
	App *BitableApp `json:"app"`
}

func (r *updateBitableAppResponse) Success() bool { return r.Code == 0 }

type copyBitableAppResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *copyBitableAppResponseData `json:"data"`
}

type copyBitableAppResponseData struct {
	App *BitableApp `json:"app"`
}

func (r *copyBitableAppResponse) Success() bool { return r.Code == 0 }

func (c *Client) CreateBitableApp(ctx context.Context, token string, name string, opts BitableAppCreateOptions) (BitableApp, error) {
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

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	if opts.CustomizedConfig != nil {
		apiReq.QueryParams.Set("customized_config", fmt.Sprint(*opts.CustomizedConfig))
	}
	if opts.SourceAppToken != "" {
		apiReq.QueryParams.Set("source_app_token", opts.SourceAppToken)
	}
	for _, value := range opts.CopyTypes {
		apiReq.QueryParams.Add("copy_types", value)
	}
	if opts.ApiType != "" {
		apiReq.QueryParams.Set("api_type", opts.ApiType)
	}

	payload := map[string]any{"name": name}
	if opts.FolderToken != "" {
		payload["folder_token"] = opts.FolderToken
	}
	if opts.TimeZone != "" {
		payload["time_zone"] = opts.TimeZone
	}
	apiReq.Body = payload

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

func (c *Client) GetBitableApp(ctx context.Context, token string, appToken string) (BitableApp, error) {
	if !c.available() || c.coreConfig == nil {
		return BitableApp{}, ErrUnavailable
	}
	if appToken == "" {
		return BitableApp{}, errors.New("app token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BitableApp{}, errors.New("tenant access token is required")
	}

	if c.bitableAppGetSDKAvailable() {
		req := larkbitable.NewGetAppReqBuilder().AppToken(appToken).Build()
		resp, err := c.sdk.Bitable.V1.App.Get(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
		if err != nil {
			return BitableApp{}, err
		}
		if resp == nil {
			return BitableApp{}, errors.New("get bitable app failed: empty response")
		}
		if !resp.Success() {
			return BitableApp{}, fmt.Errorf("get bitable app failed: %s", resp.Msg)
		}
		if resp.Data == nil || resp.Data.App == nil {
			return BitableApp{}, errors.New("get bitable app failed: missing app")
		}
		app := mapBitableAppFromSDKDisplay(resp.Data.App)
		if app.AppToken == "" {
			return BitableApp{}, errors.New("get bitable app failed: missing app_token")
		}
		return app, nil
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BitableApp{}, err
	}
	if apiResp == nil {
		return BitableApp{}, errors.New("get bitable app failed: empty response")
	}
	fallbackResp := &getBitableAppResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(fallbackResp, c.coreConfig); err != nil {
		return BitableApp{}, err
	}
	if !fallbackResp.Success() {
		return BitableApp{}, fmt.Errorf("get bitable app failed: %s", fallbackResp.Msg)
	}
	if fallbackResp.Data == nil || fallbackResp.Data.App == nil {
		return BitableApp{}, errors.New("get bitable app failed: missing app")
	}
	if fallbackResp.Data.App.AppToken == "" {
		return BitableApp{}, errors.New("get bitable app failed: missing app_token")
	}
	return *fallbackResp.Data.App, nil
}

func (c *Client) UpdateBitableApp(ctx context.Context, token string, appToken string, opts BitableAppUpdateOptions) (BitableApp, error) {
	if !c.available() || c.coreConfig == nil {
		return BitableApp{}, ErrUnavailable
	}
	if appToken == "" {
		return BitableApp{}, errors.New("app token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BitableApp{}, errors.New("tenant access token is required")
	}

	payload := map[string]any{}
	if opts.Name != "" {
		payload["name"] = opts.Name
	}
	if opts.IsAdvanced != nil {
		payload["is_advanced"] = *opts.IsAdvanced
	}
	if len(payload) == 0 {
		return BitableApp{}, errors.New("one of name or is_advanced is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BitableApp{}, err
	}
	if apiResp == nil {
		return BitableApp{}, errors.New("update bitable app failed: empty response")
	}
	resp := &updateBitableAppResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BitableApp{}, err
	}
	if !resp.Success() {
		return BitableApp{}, fmt.Errorf("update bitable app failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.App == nil {
		return BitableApp{}, errors.New("update bitable app failed: missing app")
	}
	if resp.Data.App.AppToken == "" {
		return BitableApp{}, errors.New("update bitable app failed: missing app_token")
	}
	return *resp.Data.App, nil
}

func (c *Client) CopyBitableApp(ctx context.Context, token string, appToken string, opts BitableAppCopyOptions) (BitableApp, error) {
	if !c.available() || c.coreConfig == nil {
		return BitableApp{}, ErrUnavailable
	}
	if appToken == "" {
		return BitableApp{}, errors.New("app token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BitableApp{}, errors.New("tenant access token is required")
	}

	payload := map[string]any{}
	if opts.Name != "" {
		payload["name"] = opts.Name
	}
	if opts.FolderToken != "" {
		payload["folder_token"] = opts.FolderToken
	}
	if opts.WithoutContent != nil {
		payload["without_content"] = *opts.WithoutContent
	}
	if opts.TimeZone != "" {
		payload["time_zone"] = opts.TimeZone
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/copy",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BitableApp{}, err
	}
	if apiResp == nil {
		return BitableApp{}, errors.New("copy bitable app failed: empty response")
	}
	resp := &copyBitableAppResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BitableApp{}, err
	}
	if !resp.Success() {
		return BitableApp{}, fmt.Errorf("copy bitable app failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.App == nil {
		return BitableApp{}, errors.New("copy bitable app failed: missing app")
	}
	if resp.Data.App.AppToken == "" {
		return BitableApp{}, errors.New("copy bitable app failed: missing app_token")
	}
	return *resp.Data.App, nil
}
