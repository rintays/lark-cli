package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) GetChatAnnouncement(ctx context.Context, token string, chatID string, userIDType string) (ChatAnnouncement, error) {
	if !c.available() {
		return ChatAnnouncement{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ChatAnnouncement{}, errors.New("tenant access token is required")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return ChatAnnouncement{}, errors.New("chat id is required")
	}

	builder := im.NewGetChatAnnouncementReqBuilder().ChatId(chatID)
	if strings.TrimSpace(userIDType) != "" {
		builder.UserIdType(strings.TrimSpace(userIDType))
	}
	resp, err := c.sdk.Im.V1.ChatAnnouncement.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ChatAnnouncement{}, err
	}
	if resp == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement failed: empty response")
	}
	if !resp.Success() {
		return ChatAnnouncement{}, fmt.Errorf("get chat announcement failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ChatAnnouncement{}, errors.New("get chat announcement failed: empty data")
	}
	return mapChatAnnouncement(resp.Data), nil
}

func (c *Client) UpdateChatAnnouncement(ctx context.Context, token string, chatID string, revision string, requests []string) error {
	if !c.available() {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	chatID = strings.TrimSpace(chatID)
	if chatID == "" {
		return errors.New("chat id is required")
	}
	if strings.TrimSpace(revision) == "" {
		return errors.New("revision is required")
	}
	if len(requests) == 0 {
		return errors.New("requests are required")
	}

	body := im.NewPatchChatAnnouncementReqBodyBuilder().
		Revision(strings.TrimSpace(revision)).
		Requests(requests).
		Build()
	builder := im.NewPatchChatAnnouncementReqBuilder().
		ChatId(chatID).
		Body(body)

	resp, err := c.sdk.Im.V1.ChatAnnouncement.Patch(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("update chat announcement failed: empty response")
	}
	if !resp.Success() {
		return fmt.Errorf("update chat announcement failed: %s", resp.Msg)
	}
	return nil
}

func mapChatAnnouncement(data *im.GetChatAnnouncementRespData) ChatAnnouncement {
	if data == nil {
		return ChatAnnouncement{}
	}
	out := ChatAnnouncement{}
	if data.Content != nil {
		out.Content = *data.Content
	}
	if data.Revision != nil {
		out.Revision = *data.Revision
	}
	if data.CreateTime != nil {
		out.CreateTime = *data.CreateTime
	}
	if data.UpdateTime != nil {
		out.UpdateTime = *data.UpdateTime
	}
	if data.OwnerIdType != nil {
		out.OwnerIDType = *data.OwnerIdType
	}
	if data.OwnerId != nil {
		out.OwnerID = *data.OwnerId
	}
	if data.ModifierIdType != nil {
		out.ModifierIDType = *data.ModifierIdType
	}
	if data.ModifierId != nil {
		out.ModifierID = *data.ModifierId
	}
	return out
}
