package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
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
