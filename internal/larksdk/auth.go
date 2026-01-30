package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
)

func (c *Client) TenantAccessToken(ctx context.Context) (string, int64, error) {
	if !c.available() {
		return "", 0, ErrUnavailable
	}
	req := &larkcore.SelfBuiltTenantAccessTokenReq{
		AppID:     c.coreConfig.AppId,
		AppSecret: c.coreConfig.AppSecret,
	}
	resp, err := c.sdk.GetTenantAccessTokenBySelfBuiltApp(ctx, req)
	if err != nil {
		return "", 0, err
	}
	if resp == nil {
		return "", 0, errors.New("tenant access token failed: empty response")
	}
	if !resp.Success() {
		return "", 0, fmt.Errorf("tenant access token failed: %s", resp.Msg)
	}
	if resp.TenantAccessToken == "" {
		return "", 0, errors.New("tenant access token missing from response")
	}
	return resp.TenantAccessToken, int64(resp.Expire), nil
}

func (c *Client) RefreshUserAccessToken(ctx context.Context, refreshToken string) (string, string, int64, error) {
	if !c.available() {
		return "", "", 0, ErrUnavailable
	}
	if refreshToken == "" {
		return "", "", 0, errors.New("refresh token is required")
	}
	appAccessToken, err := c.appAccessToken(ctx)
	if err != nil {
		return "", "", 0, err
	}
	body := larkauthen.NewCreateRefreshAccessTokenReqBodyBuilder().
		GrantType("refresh_token").
		RefreshToken(refreshToken).
		Build()
	req := larkauthen.NewCreateRefreshAccessTokenReqBuilder().Body(body).Build()
	resp, err := c.sdk.Authen.V1.RefreshAccessToken.Create(ctx, req, func(option *larkcore.RequestOption) {
		option.AppAccessToken = appAccessToken
	})
	if err != nil {
		return "", "", 0, err
	}
	if resp == nil {
		return "", "", 0, errors.New("refresh access token failed: empty response")
	}
	if !resp.Success() {
		return "", "", 0, fmt.Errorf("refresh access token failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return "", "", 0, errors.New("refresh access token failed: missing data")
	}
	accessToken := ""
	if resp.Data.AccessToken != nil {
		accessToken = *resp.Data.AccessToken
	}
	if accessToken == "" {
		return "", "", 0, errors.New("refresh access token failed: missing access_token")
	}
	expiresIn := int64(0)
	if resp.Data.ExpiresIn != nil {
		expiresIn = int64(*resp.Data.ExpiresIn)
	}
	if expiresIn <= 0 {
		return "", "", 0, errors.New("refresh access token failed: invalid expires_in")
	}
	newRefreshToken := ""
	if resp.Data.RefreshToken != nil {
		newRefreshToken = *resp.Data.RefreshToken
	}
	return accessToken, newRefreshToken, expiresIn, nil
}

func (c *Client) appAccessToken(ctx context.Context) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	resp, err := larkcore.Request(ctx, &larkcore.ApiReq{
		HttpMethod: http.MethodPost,
		ApiPath:    larkcore.AppAccessTokenInternalUrlPath,
		Body: &larkcore.SelfBuiltAppAccessTokenReq{
			AppID:     c.coreConfig.AppId,
			AppSecret: c.coreConfig.AppSecret,
		},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeNone},
	}, c.coreConfig)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("app access token failed: empty response")
	}
	var parsed larkcore.AppAccessTokenResp
	if err := json.Unmarshal(resp.RawBody, &parsed); err != nil {
		return "", err
	}
	if !parsed.Success() {
		return "", fmt.Errorf("app access token failed: %s", parsed.Msg)
	}
	if parsed.AppAccessToken == "" {
		return "", errors.New("app access token failed: missing app_access_token")
	}
	return parsed.AppAccessToken, nil
}
