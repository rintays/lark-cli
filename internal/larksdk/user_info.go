package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkauthen "github.com/larksuite/oapi-sdk-go/v3/service/authen/v1"
)

func (c *Client) UserInfo(ctx context.Context, token string) (UserInfo, error) {
	if !c.available() {
		return UserInfo{}, ErrUnavailable
	}
	userAccessToken := strings.TrimSpace(token)
	if userAccessToken == "" {
		return UserInfo{}, errors.New("user access token is required")
	}

	resp, err := c.sdk.Authen.V1.UserInfo.Get(ctx, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return UserInfo{}, err
	}
	if resp == nil {
		return UserInfo{}, errors.New("user info failed: empty response")
	}
	if !resp.Success() {
		return UserInfo{}, fmt.Errorf("user info failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return UserInfo{}, errors.New("user info response missing data")
	}
	return mapUserInfo(resp.Data), nil
}

func mapUserInfo(data *larkauthen.GetUserInfoRespData) UserInfo {
	if data == nil {
		return UserInfo{}
	}
	result := UserInfo{}
	if data.Name != nil {
		result.Name = *data.Name
	}
	if data.EnName != nil {
		result.EnName = *data.EnName
	}
	if data.AvatarUrl != nil {
		result.AvatarURL = *data.AvatarUrl
	}
	if data.AvatarThumb != nil {
		result.AvatarThumb = *data.AvatarThumb
	}
	if data.AvatarMiddle != nil {
		result.AvatarMiddle = *data.AvatarMiddle
	}
	if data.AvatarBig != nil {
		result.AvatarBig = *data.AvatarBig
	}
	if data.OpenId != nil {
		result.OpenID = *data.OpenId
	}
	if data.UnionId != nil {
		result.UnionID = *data.UnionId
	}
	if data.Email != nil {
		result.Email = *data.Email
	}
	if data.EnterpriseEmail != nil {
		result.EnterpriseEmail = *data.EnterpriseEmail
	}
	if data.UserId != nil {
		result.UserID = *data.UserId
	}
	if data.Mobile != nil {
		result.Mobile = *data.Mobile
	}
	if data.TenantKey != nil {
		result.TenantKey = *data.TenantKey
	}
	if data.EmployeeNo != nil {
		result.EmployeeNo = *data.EmployeeNo
	}
	return result
}
