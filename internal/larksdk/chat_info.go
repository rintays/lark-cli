package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) CreateChatDetail(ctx context.Context, token string, req CreateChatRequest) (ChatInfo, error) {
	if !c.available() {
		return ChatInfo{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ChatInfo{}, errors.New("tenant access token is required")
	}

	body := &im.CreateChatReqBody{}
	if strings.TrimSpace(req.Avatar) != "" {
		body.Avatar = stringPtr(strings.TrimSpace(req.Avatar))
	}
	if strings.TrimSpace(req.Name) != "" {
		body.Name = stringPtr(strings.TrimSpace(req.Name))
	}
	if strings.TrimSpace(req.Description) != "" {
		body.Description = stringPtr(strings.TrimSpace(req.Description))
	}
	if req.I18nNames != nil {
		body.I18nNames = mapI18nNamesToSDK(req.I18nNames)
	}
	if strings.TrimSpace(req.OwnerID) != "" {
		body.OwnerId = stringPtr(strings.TrimSpace(req.OwnerID))
	}
	if len(req.UserIDList) > 0 {
		body.UserIdList = req.UserIDList
	}
	if len(req.BotIDList) > 0 {
		body.BotIdList = req.BotIDList
	}
	if strings.TrimSpace(req.GroupMessageType) != "" {
		body.GroupMessageType = stringPtr(strings.TrimSpace(req.GroupMessageType))
	}
	if strings.TrimSpace(req.ChatMode) != "" {
		body.ChatMode = stringPtr(strings.TrimSpace(req.ChatMode))
	}
	if strings.TrimSpace(req.ChatType) != "" {
		body.ChatType = stringPtr(strings.TrimSpace(req.ChatType))
	}
	if strings.TrimSpace(req.JoinMessageVisibility) != "" {
		body.JoinMessageVisibility = stringPtr(strings.TrimSpace(req.JoinMessageVisibility))
	}
	if strings.TrimSpace(req.LeaveMessageVisibility) != "" {
		body.LeaveMessageVisibility = stringPtr(strings.TrimSpace(req.LeaveMessageVisibility))
	}
	if strings.TrimSpace(req.MembershipApproval) != "" {
		body.MembershipApproval = stringPtr(strings.TrimSpace(req.MembershipApproval))
	}
	if req.RestrictedModeSetting != nil {
		body.RestrictedModeSetting = mapRestrictedModeSettingToSDK(req.RestrictedModeSetting)
	}
	if strings.TrimSpace(req.UrgentSetting) != "" {
		body.UrgentSetting = stringPtr(strings.TrimSpace(req.UrgentSetting))
	}
	if strings.TrimSpace(req.VideoConferenceSetting) != "" {
		body.VideoConferenceSetting = stringPtr(strings.TrimSpace(req.VideoConferenceSetting))
	}
	if strings.TrimSpace(req.EditPermission) != "" {
		body.EditPermission = stringPtr(strings.TrimSpace(req.EditPermission))
	}
	if strings.TrimSpace(req.HideMemberCountSetting) != "" {
		body.HideMemberCountSetting = stringPtr(strings.TrimSpace(req.HideMemberCountSetting))
	}
	if strings.TrimSpace(req.PinManageSetting) != "" {
		body.PinManageSetting = stringPtr(strings.TrimSpace(req.PinManageSetting))
	}
	if req.External != nil {
		body.External = req.External
	}

	builder := im.NewCreateChatReqBuilder().Body(body)
	if strings.TrimSpace(req.UserIDType) != "" {
		builder.UserIdType(strings.TrimSpace(req.UserIDType))
	}
	if req.SetBotManager != nil {
		builder.SetBotManager(*req.SetBotManager)
	}
	if strings.TrimSpace(req.UUID) != "" {
		builder.Uuid(strings.TrimSpace(req.UUID))
	}

	resp, err := c.sdk.Im.V1.Chat.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ChatInfo{}, err
	}
	if resp == nil {
		return ChatInfo{}, errors.New("create chat failed: empty response")
	}
	if !resp.Success() {
		return ChatInfo{}, fmt.Errorf("create chat failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ChatInfo{}, errors.New("create chat failed: empty data")
	}
	return mapChatInfoFromCreate(resp.Data), nil
}

func (c *Client) GetChatInfo(ctx context.Context, token string, req GetChatRequest) (ChatInfo, error) {
	if !c.available() {
		return ChatInfo{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ChatInfo{}, errors.New("tenant access token is required")
	}
	chatID := strings.TrimSpace(req.ChatID)
	if chatID == "" {
		return ChatInfo{}, errors.New("chat id is required")
	}

	builder := im.NewGetChatReqBuilder().ChatId(chatID)
	if strings.TrimSpace(req.UserIDType) != "" {
		builder.UserIdType(strings.TrimSpace(req.UserIDType))
	}
	resp, err := c.sdk.Im.V1.Chat.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ChatInfo{}, err
	}
	if resp == nil {
		return ChatInfo{}, errors.New("get chat failed: empty response")
	}
	if !resp.Success() {
		return ChatInfo{}, fmt.Errorf("get chat failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ChatInfo{}, errors.New("get chat failed: empty data")
	}
	chat := mapChatInfoFromGet(resp.Data)
	chat.ChatID = chatID
	return chat, nil
}

func (c *Client) UpdateChatInfo(ctx context.Context, token string, req UpdateChatRequest) error {
	if !c.available() {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	chatID := strings.TrimSpace(req.ChatID)
	if chatID == "" {
		return errors.New("chat id is required")
	}

	body := &im.UpdateChatReqBody{}
	if strings.TrimSpace(req.Avatar) != "" {
		body.Avatar = stringPtr(strings.TrimSpace(req.Avatar))
	}
	if strings.TrimSpace(req.Name) != "" {
		body.Name = stringPtr(strings.TrimSpace(req.Name))
	}
	if strings.TrimSpace(req.Description) != "" {
		body.Description = stringPtr(strings.TrimSpace(req.Description))
	}
	if req.I18nNames != nil {
		body.I18nNames = mapI18nNamesToSDK(req.I18nNames)
	}
	if strings.TrimSpace(req.AddMemberPermission) != "" {
		body.AddMemberPermission = stringPtr(strings.TrimSpace(req.AddMemberPermission))
	}
	if strings.TrimSpace(req.ShareCardPermission) != "" {
		body.ShareCardPermission = stringPtr(strings.TrimSpace(req.ShareCardPermission))
	}
	if strings.TrimSpace(req.AtAllPermission) != "" {
		body.AtAllPermission = stringPtr(strings.TrimSpace(req.AtAllPermission))
	}
	if strings.TrimSpace(req.EditPermission) != "" {
		body.EditPermission = stringPtr(strings.TrimSpace(req.EditPermission))
	}
	if strings.TrimSpace(req.OwnerID) != "" {
		body.OwnerId = stringPtr(strings.TrimSpace(req.OwnerID))
	}
	if strings.TrimSpace(req.JoinMessageVisibility) != "" {
		body.JoinMessageVisibility = stringPtr(strings.TrimSpace(req.JoinMessageVisibility))
	}
	if strings.TrimSpace(req.LeaveMessageVisibility) != "" {
		body.LeaveMessageVisibility = stringPtr(strings.TrimSpace(req.LeaveMessageVisibility))
	}
	if strings.TrimSpace(req.MembershipApproval) != "" {
		body.MembershipApproval = stringPtr(strings.TrimSpace(req.MembershipApproval))
	}
	if req.RestrictedModeSetting != nil {
		body.RestrictedModeSetting = mapRestrictedModeSettingToSDK(req.RestrictedModeSetting)
	}
	if strings.TrimSpace(req.ChatType) != "" {
		body.ChatType = stringPtr(strings.TrimSpace(req.ChatType))
	}
	if strings.TrimSpace(req.GroupMessageType) != "" {
		body.GroupMessageType = stringPtr(strings.TrimSpace(req.GroupMessageType))
	}
	if strings.TrimSpace(req.UrgentSetting) != "" {
		body.UrgentSetting = stringPtr(strings.TrimSpace(req.UrgentSetting))
	}
	if strings.TrimSpace(req.VideoConferenceSetting) != "" {
		body.VideoConferenceSetting = stringPtr(strings.TrimSpace(req.VideoConferenceSetting))
	}
	if strings.TrimSpace(req.PinManageSetting) != "" {
		body.PinManageSetting = stringPtr(strings.TrimSpace(req.PinManageSetting))
	}
	if strings.TrimSpace(req.HideMemberCountSetting) != "" {
		body.HideMemberCountSetting = stringPtr(strings.TrimSpace(req.HideMemberCountSetting))
	}

	builder := im.NewUpdateChatReqBuilder().
		ChatId(chatID).
		Body(body)
	if strings.TrimSpace(req.UserIDType) != "" {
		builder.UserIdType(strings.TrimSpace(req.UserIDType))
	}
	resp, err := c.sdk.Im.V1.Chat.Update(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("update chat failed: empty response")
	}
	if !resp.Success() {
		return fmt.Errorf("update chat failed: %s", resp.Msg)
	}
	return nil
}

func mapChatInfoFromCreate(chat *im.CreateChatRespData) ChatInfo {
	if chat == nil {
		return ChatInfo{}
	}
	out := ChatInfo{}
	if chat.ChatId != nil {
		out.ChatID = *chat.ChatId
	}
	if chat.Avatar != nil {
		out.Avatar = *chat.Avatar
	}
	if chat.Name != nil {
		out.Name = *chat.Name
	}
	if chat.Description != nil {
		out.Description = *chat.Description
	}
	if chat.I18nNames != nil {
		out.I18nNames = mapI18nNamesFromSDK(chat.I18nNames)
	}
	if chat.OwnerId != nil {
		out.OwnerID = *chat.OwnerId
	}
	if chat.OwnerIdType != nil {
		out.OwnerIDType = *chat.OwnerIdType
	}
	if chat.UrgentSetting != nil {
		out.UrgentSetting = *chat.UrgentSetting
	}
	if chat.VideoConferenceSetting != nil {
		out.VideoConferenceSetting = *chat.VideoConferenceSetting
	}
	if chat.PinManageSetting != nil {
		out.PinManageSetting = *chat.PinManageSetting
	}
	if chat.AddMemberPermission != nil {
		out.AddMemberPermission = *chat.AddMemberPermission
	}
	if chat.ShareCardPermission != nil {
		out.ShareCardPermission = *chat.ShareCardPermission
	}
	if chat.AtAllPermission != nil {
		out.AtAllPermission = *chat.AtAllPermission
	}
	if chat.EditPermission != nil {
		out.EditPermission = *chat.EditPermission
	}
	if chat.GroupMessageType != nil {
		out.GroupMessageType = *chat.GroupMessageType
	}
	if chat.ChatMode != nil {
		out.ChatMode = *chat.ChatMode
	}
	if chat.ChatType != nil {
		out.ChatType = *chat.ChatType
	}
	if chat.ChatTag != nil {
		out.ChatTag = *chat.ChatTag
	}
	if chat.External != nil {
		out.External = *chat.External
	}
	if chat.TenantKey != nil {
		out.TenantKey = *chat.TenantKey
	}
	if chat.JoinMessageVisibility != nil {
		out.JoinMessageVisibility = *chat.JoinMessageVisibility
	}
	if chat.LeaveMessageVisibility != nil {
		out.LeaveMessageVisibility = *chat.LeaveMessageVisibility
	}
	if chat.MembershipApproval != nil {
		out.MembershipApproval = *chat.MembershipApproval
	}
	if chat.ModerationPermission != nil {
		out.ModerationPermission = *chat.ModerationPermission
	}
	if chat.RestrictedModeSetting != nil {
		out.RestrictedModeSetting = mapRestrictedModeSettingFromSDK(chat.RestrictedModeSetting)
	}
	if chat.HideMemberCountSetting != nil {
		out.HideMemberCountSetting = *chat.HideMemberCountSetting
	}
	return out
}

func mapChatInfoFromGet(chat *im.GetChatRespData) ChatInfo {
	if chat == nil {
		return ChatInfo{}
	}
	out := ChatInfo{}
	if chat.Avatar != nil {
		out.Avatar = *chat.Avatar
	}
	if chat.Name != nil {
		out.Name = *chat.Name
	}
	if chat.Description != nil {
		out.Description = *chat.Description
	}
	if chat.I18nNames != nil {
		out.I18nNames = mapI18nNamesFromSDK(chat.I18nNames)
	}
	if chat.OwnerId != nil {
		out.OwnerID = *chat.OwnerId
	}
	if chat.OwnerIdType != nil {
		out.OwnerIDType = *chat.OwnerIdType
	}
	if len(chat.UserManagerIdList) > 0 {
		out.UserManagerIDList = chat.UserManagerIdList
	}
	if len(chat.BotManagerIdList) > 0 {
		out.BotManagerIDList = chat.BotManagerIdList
	}
	if chat.GroupMessageType != nil {
		out.GroupMessageType = *chat.GroupMessageType
	}
	if chat.ChatMode != nil {
		out.ChatMode = *chat.ChatMode
	}
	if chat.ChatType != nil {
		out.ChatType = *chat.ChatType
	}
	if chat.ChatTag != nil {
		out.ChatTag = *chat.ChatTag
	}
	if chat.JoinMessageVisibility != nil {
		out.JoinMessageVisibility = *chat.JoinMessageVisibility
	}
	if chat.LeaveMessageVisibility != nil {
		out.LeaveMessageVisibility = *chat.LeaveMessageVisibility
	}
	if chat.MembershipApproval != nil {
		out.MembershipApproval = *chat.MembershipApproval
	}
	if chat.ModerationPermission != nil {
		out.ModerationPermission = *chat.ModerationPermission
	}
	if chat.AddMemberPermission != nil {
		out.AddMemberPermission = *chat.AddMemberPermission
	}
	if chat.ShareCardPermission != nil {
		out.ShareCardPermission = *chat.ShareCardPermission
	}
	if chat.AtAllPermission != nil {
		out.AtAllPermission = *chat.AtAllPermission
	}
	if chat.EditPermission != nil {
		out.EditPermission = *chat.EditPermission
	}
	if chat.External != nil {
		out.External = *chat.External
	}
	if chat.TenantKey != nil {
		out.TenantKey = *chat.TenantKey
	}
	if chat.UserCount != nil {
		out.UserCount = *chat.UserCount
	}
	if chat.BotCount != nil {
		out.BotCount = *chat.BotCount
	}
	if chat.RestrictedModeSetting != nil {
		out.RestrictedModeSetting = mapRestrictedModeSettingFromSDK(chat.RestrictedModeSetting)
	}
	if chat.UrgentSetting != nil {
		out.UrgentSetting = *chat.UrgentSetting
	}
	if chat.VideoConferenceSetting != nil {
		out.VideoConferenceSetting = *chat.VideoConferenceSetting
	}
	if chat.PinManageSetting != nil {
		out.PinManageSetting = *chat.PinManageSetting
	}
	if chat.HideMemberCountSetting != nil {
		out.HideMemberCountSetting = *chat.HideMemberCountSetting
	}
	if chat.ChatStatus != nil {
		out.ChatStatus = *chat.ChatStatus
	}
	return out
}

func mapI18nNamesFromSDK(names *im.I18nNames) *I18nNames {
	if names == nil {
		return nil
	}
	out := &I18nNames{}
	if names.ZhCn != nil {
		out.ZhCn = *names.ZhCn
	}
	if names.EnUs != nil {
		out.EnUs = *names.EnUs
	}
	if names.JaJp != nil {
		out.JaJp = *names.JaJp
	}
	return out
}

func mapI18nNamesToSDK(names *I18nNames) *im.I18nNames {
	if names == nil {
		return nil
	}
	out := &im.I18nNames{}
	if strings.TrimSpace(names.ZhCn) != "" {
		out.ZhCn = stringPtr(strings.TrimSpace(names.ZhCn))
	}
	if strings.TrimSpace(names.EnUs) != "" {
		out.EnUs = stringPtr(strings.TrimSpace(names.EnUs))
	}
	if strings.TrimSpace(names.JaJp) != "" {
		out.JaJp = stringPtr(strings.TrimSpace(names.JaJp))
	}
	return out
}

func mapRestrictedModeSettingFromSDK(setting *im.RestrictedModeSetting) *RestrictedModeSetting {
	if setting == nil {
		return nil
	}
	out := &RestrictedModeSetting{}
	if setting.Status != nil {
		out.Status = *setting.Status
	}
	if setting.ScreenshotHasPermissionSetting != nil {
		out.ScreenshotHasPermissionSetting = *setting.ScreenshotHasPermissionSetting
	}
	if setting.DownloadHasPermissionSetting != nil {
		out.DownloadHasPermissionSetting = *setting.DownloadHasPermissionSetting
	}
	if setting.MessageHasPermissionSetting != nil {
		out.MessageHasPermissionSetting = *setting.MessageHasPermissionSetting
	}
	return out
}

func mapRestrictedModeSettingToSDK(setting *RestrictedModeSetting) *im.RestrictedModeSetting {
	if setting == nil {
		return nil
	}
	out := &im.RestrictedModeSetting{}
	out.Status = boolPtr(setting.Status)
	if strings.TrimSpace(setting.ScreenshotHasPermissionSetting) != "" {
		out.ScreenshotHasPermissionSetting = stringPtr(strings.TrimSpace(setting.ScreenshotHasPermissionSetting))
	}
	if strings.TrimSpace(setting.DownloadHasPermissionSetting) != "" {
		out.DownloadHasPermissionSetting = stringPtr(strings.TrimSpace(setting.DownloadHasPermissionSetting))
	}
	if strings.TrimSpace(setting.MessageHasPermissionSetting) != "" {
		out.MessageHasPermissionSetting = stringPtr(strings.TrimSpace(setting.MessageHasPermissionSetting))
	}
	return out
}

func stringPtr(value string) *string {
	if strings.TrimSpace(value) == "" {
		return nil
	}
	return &value
}

func boolPtr(value bool) *bool {
	return &value
}
