package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
)

type UpdateWikiSpaceSettingRequest struct {
	SpaceID            string
	CreateSetting      string
	SecuritySetting    string
	CommentSetting     string
	CreateSettingSet   bool
	SecuritySettingSet bool
	CommentSettingSet  bool
}

type WikiSpaceSetting struct {
	CreateSetting   string `json:"create_setting,omitempty"`
	SecuritySetting string `json:"security_setting,omitempty"`
	CommentSetting  string `json:"comment_setting,omitempty"`
}

func (c *Client) UpdateWikiSpaceSettingV2(ctx context.Context, token string, req UpdateWikiSpaceSettingRequest) (WikiSpaceSetting, error) {
	if !c.available() {
		return WikiSpaceSetting{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiSpaceSetting{}, errors.New("tenant access token is required")
	}
	return c.updateWikiSpaceSettingV2(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) UpdateWikiSpaceSettingV2WithUserToken(ctx context.Context, userAccessToken string, req UpdateWikiSpaceSettingRequest) (WikiSpaceSetting, error) {
	if !c.available() {
		return WikiSpaceSetting{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return WikiSpaceSetting{}, errors.New("user access token is required")
	}
	return c.updateWikiSpaceSettingV2(ctx, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) updateWikiSpaceSettingV2(ctx context.Context, req UpdateWikiSpaceSettingRequest, option larkcore.RequestOptionFunc) (WikiSpaceSetting, error) {
	if !c.available() {
		return WikiSpaceSetting{}, ErrUnavailable
	}
	spaceID := strings.TrimSpace(req.SpaceID)
	if spaceID == "" {
		return WikiSpaceSetting{}, errors.New("space id is required")
	}
	if !req.CreateSettingSet && !req.SecuritySettingSet && !req.CommentSettingSet {
		return WikiSpaceSetting{}, errors.New("at least one setting is required")
	}

	builder := larkwiki.NewSettingBuilder()
	if req.CreateSettingSet {
		value := strings.TrimSpace(req.CreateSetting)
		if value == "" {
			return WikiSpaceSetting{}, errors.New("create setting is required")
		}
		builder = builder.CreateSetting(value)
	}
	if req.SecuritySettingSet {
		value := strings.TrimSpace(req.SecuritySetting)
		if value == "" {
			return WikiSpaceSetting{}, errors.New("security setting is required")
		}
		builder = builder.SecuritySetting(value)
	}
	if req.CommentSettingSet {
		value := strings.TrimSpace(req.CommentSetting)
		if value == "" {
			return WikiSpaceSetting{}, errors.New("comment setting is required")
		}
		builder = builder.CommentSetting(value)
	}

	setting := builder.Build()
	reqBuilder := larkwiki.NewUpdateSpaceSettingReqBuilder().SpaceId(spaceID).Setting(setting)
	resp, err := c.sdk.Wiki.V2.SpaceSetting.Update(ctx, reqBuilder.Build(), option)
	if err != nil {
		return WikiSpaceSetting{}, err
	}
	if resp == nil {
		return WikiSpaceSetting{}, errors.New("wiki space setting update failed: empty response")
	}
	if !resp.Success() {
		return WikiSpaceSetting{}, fmt.Errorf("wiki space setting update failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Setting == nil {
		return WikiSpaceSetting{}, nil
	}
	return convertWikiSpaceSetting(resp.Data.Setting), nil
}

func convertWikiSpaceSetting(setting *larkwiki.Setting) WikiSpaceSetting {
	out := WikiSpaceSetting{}
	if setting == nil {
		return out
	}
	if setting.CreateSetting != nil {
		out.CreateSetting = *setting.CreateSetting
	}
	if setting.SecuritySetting != nil {
		out.SecuritySetting = *setting.SecuritySetting
	}
	if setting.CommentSetting != nil {
		out.CommentSetting = *setting.CommentSetting
	}
	return out
}
