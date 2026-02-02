package larksdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type RefreshAccessTokenError struct {
	Code int
	Msg  string
}

func (e *RefreshAccessTokenError) Error() string {
	if e == nil {
		return "refresh access token failed"
	}
	if e.Msg == "" {
		return fmt.Sprintf("refresh access token failed (code=%d)", e.Code)
	}
	return fmt.Sprintf("refresh access token failed (code=%d): %s", e.Code, e.Msg)
}

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
	endpoint, err := c.oauthTokenURL()
	if err != nil {
		return "", "", 0, err
	}
	payload := map[string]string{
		"grant_type":    "refresh_token",
		"client_id":     c.coreConfig.AppId,
		"client_secret": c.coreConfig.AppSecret,
		"refresh_token": refreshToken,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", "", 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", "", 0, err
	}
	req.Header.Set("Content-Type", "application/json; charset=utf-8")

	httpClient := http.DefaultClient
	if c.coreConfig != nil && c.coreConfig.HttpClient != nil {
		httpClient = c.coreConfig.HttpClient
	}
	resp, err := httpClient.Do(req)
	if err != nil {
		return "", "", 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", "", 0, err
	}
	var parsed oauthTokenResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		if resp.StatusCode < 200 || resp.StatusCode > 299 {
			return "", "", 0, fmt.Errorf("refresh access token failed: %s", strings.TrimSpace(string(data)))
		}
		return "", "", 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		msg := parsed.ErrorDescription
		if msg == "" {
			msg = parsed.Error
		}
		if msg == "" {
			msg = parsed.Msg
		}
		if msg == "" {
			msg = strings.TrimSpace(string(data))
		}
		return "", "", 0, &RefreshAccessTokenError{Code: parsed.Code, Msg: msg}
	}
	if parsed.Code != 0 || parsed.Error != "" {
		msg := parsed.ErrorDescription
		if msg == "" {
			msg = parsed.Error
		}
		if msg == "" {
			msg = parsed.Msg
		}
		return "", "", 0, &RefreshAccessTokenError{Code: parsed.Code, Msg: msg}
	}
	if parsed.AccessToken == "" {
		return "", "", 0, errors.New("refresh access token failed: missing access_token")
	}
	if parsed.ExpiresIn <= 0 {
		return "", "", 0, errors.New("refresh access token failed: invalid expires_in")
	}
	return parsed.AccessToken, parsed.RefreshToken, parsed.ExpiresIn, nil
}

type oauthTokenResponse struct {
	Code             int    `json:"code"`
	Msg              string `json:"msg"`
	AccessToken      string `json:"access_token"`
	ExpiresIn        int64  `json:"expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Error            string `json:"error"`
	ErrorDescription string `json:"error_description"`
}

func (c *Client) oauthTokenURL() (string, error) {
	if c == nil || c.coreConfig == nil {
		return "", errors.New("sdk config is required")
	}
	base, err := url.Parse(c.coreConfig.BaseUrl)
	if err != nil {
		return "", err
	}
	base.Path = "/open-apis/authen/v2/oauth/token"
	return base.String(), nil
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
