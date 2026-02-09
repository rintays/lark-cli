package larksdk

import (
	"context"
	"errors"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) ListChats(ctx context.Context, token string, req ListChatsRequest) (ListChatsResult, error) {
	if !c.available() {
		return ListChatsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListChatsResult{}, errors.New("tenant access token is required")
	}

	builder := im.NewListChatReqBuilder()
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Im.V1.Chat.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListChatsResult{}, err
	}
	if resp == nil {
		return ListChatsResult{}, errors.New("list chats failed: empty response")
	}
	if !resp.Success() {
		return ListChatsResult{}, apiError("list chats", resp.Code, resp.Msg)
	}

	result := ListChatsResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]Chat, 0, len(resp.Data.Items))
			for _, item := range resp.Data.Items {
				result.Items = append(result.Items, mapChat(item))
			}
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

func mapChat(chat *im.ListChat) Chat {
	if chat == nil {
		return Chat{}
	}
	result := Chat{}
	if chat.ChatId != nil {
		result.ChatID = *chat.ChatId
	}
	if chat.Avatar != nil {
		result.Avatar = *chat.Avatar
	}
	if chat.Name != nil {
		result.Name = *chat.Name
	}
	if chat.Description != nil {
		result.Description = *chat.Description
	}
	if chat.OwnerId != nil {
		result.OwnerID = *chat.OwnerId
	}
	if chat.OwnerIdType != nil {
		result.OwnerIDType = *chat.OwnerIdType
	}
	if chat.External != nil {
		result.External = *chat.External
	}
	if chat.TenantKey != nil {
		result.TenantKey = *chat.TenantKey
	}
	return result
}
