package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) ListChatMembers(ctx context.Context, token string, req ListChatMembersRequest) (ListChatMembersResult, error) {
	if !c.available() {
		return ListChatMembersResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListChatMembersResult{}, errors.New("tenant access token is required")
	}
	chatID := strings.TrimSpace(req.ChatID)
	if chatID == "" {
		return ListChatMembersResult{}, errors.New("chat id is required")
	}

	builder := im.NewGetChatMembersReqBuilder().ChatId(chatID)
	if strings.TrimSpace(req.MemberIDType) != "" {
		builder.MemberIdType(strings.TrimSpace(req.MemberIDType))
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if strings.TrimSpace(req.PageToken) != "" {
		builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Im.V1.ChatMembers.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListChatMembersResult{}, err
	}
	if resp == nil {
		return ListChatMembersResult{}, errors.New("list chat members failed: empty response")
	}
	if !resp.Success() {
		return ListChatMembersResult{}, fmt.Errorf("list chat members failed: %s", resp.Msg)
	}

	result := ListChatMembersResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]ChatMember, 0, len(resp.Data.Items))
			for _, item := range resp.Data.Items {
				result.Items = append(result.Items, mapChatMember(item))
			}
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
		if resp.Data.MemberTotal != nil {
			result.MemberTotal = *resp.Data.MemberTotal
		}
	}
	return result, nil
}

func mapChatMember(member *im.ListMember) ChatMember {
	if member == nil {
		return ChatMember{}
	}
	out := ChatMember{}
	if member.MemberIdType != nil {
		out.MemberIDType = *member.MemberIdType
	}
	if member.MemberId != nil {
		out.MemberID = *member.MemberId
	}
	if member.Name != nil {
		out.Name = *member.Name
	}
	if member.TenantKey != nil {
		out.TenantKey = *member.TenantKey
	}
	return out
}
