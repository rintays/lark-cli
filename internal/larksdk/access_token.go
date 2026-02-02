package larksdk

import (
	"errors"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type AccessTokenType string

const (
	AccessTokenTenant AccessTokenType = "tenant"
	AccessTokenUser   AccessTokenType = "user"
)

func (c *Client) resolveAccessToken(token string, tokenType AccessTokenType) (string, error) {
	switch tokenType {
	case AccessTokenTenant:
		tenantToken := c.tenantToken(token)
		if tenantToken == "" {
			return "", errors.New("tenant access token is required")
		}
		return tenantToken, nil
	case AccessTokenUser:
		userToken := strings.TrimSpace(token)
		if userToken == "" {
			return "", errors.New("user access token is required")
		}
		return userToken, nil
	default:
		return "", errors.New("unsupported access token type")
	}
}

func (c *Client) accessTokenOption(token string, tokenType AccessTokenType) (larkcore.RequestOptionFunc, string, error) {
	resolved, err := c.resolveAccessToken(token, tokenType)
	if err != nil {
		return nil, "", err
	}
	switch tokenType {
	case AccessTokenTenant:
		return larkcore.WithTenantAccessToken(resolved), resolved, nil
	case AccessTokenUser:
		return larkcore.WithUserAccessToken(resolved), resolved, nil
	default:
		return nil, "", errors.New("unsupported access token type")
	}
}
