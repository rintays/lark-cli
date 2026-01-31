package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) ListMessages(ctx context.Context, token string, req ListMessagesRequest) (ListMessagesResult, error) {
	if !c.available() {
		return ListMessagesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListMessagesResult{}, errors.New("tenant access token is required")
	}
	if strings.TrimSpace(req.ContainerIDType) == "" {
		return ListMessagesResult{}, errors.New("container id type is required")
	}
	if strings.TrimSpace(req.ContainerID) == "" {
		return ListMessagesResult{}, errors.New("container id is required")
	}

	builder := im.NewListMessageReqBuilder().
		ContainerIdType(req.ContainerIDType).
		ContainerId(req.ContainerID)
	if strings.TrimSpace(req.StartTime) != "" {
		builder.StartTime(req.StartTime)
	}
	if strings.TrimSpace(req.EndTime) != "" {
		builder.EndTime(req.EndTime)
	}
	if strings.TrimSpace(req.SortType) != "" {
		builder.SortType(req.SortType)
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if strings.TrimSpace(req.PageToken) != "" {
		builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Im.V1.Message.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListMessagesResult{}, err
	}
	if resp == nil {
		return ListMessagesResult{}, errors.New("list messages failed: empty response")
	}
	if !resp.Success() {
		return ListMessagesResult{}, fmt.Errorf("list messages failed: %s", resp.Msg)
	}

	result := ListMessagesResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]Message, 0, len(resp.Data.Items))
			for _, item := range resp.Data.Items {
				result.Items = append(result.Items, mapMessage(item))
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

func mapMessage(msg *im.Message) Message {
	if msg == nil {
		return Message{}
	}
	out := Message{}
	if msg.MessageId != nil {
		out.MessageID = *msg.MessageId
	}
	if msg.RootId != nil {
		out.RootID = *msg.RootId
	}
	if msg.ParentId != nil {
		out.ParentID = *msg.ParentId
	}
	if msg.ThreadId != nil {
		out.ThreadID = *msg.ThreadId
	}
	if msg.MsgType != nil {
		out.MsgType = *msg.MsgType
	}
	if msg.CreateTime != nil {
		out.CreateTime = *msg.CreateTime
	}
	if msg.UpdateTime != nil {
		out.UpdateTime = *msg.UpdateTime
	}
	if msg.Deleted != nil {
		out.Deleted = *msg.Deleted
	}
	if msg.Updated != nil {
		out.Updated = *msg.Updated
	}
	if msg.ChatId != nil {
		out.ChatID = *msg.ChatId
	}
	if msg.Sender != nil {
		out.Sender = mapMessageSender(msg.Sender)
	}
	if msg.Body != nil {
		out.Body = mapMessageBody(msg.Body)
	}
	if msg.Mentions != nil {
		out.Mentions = make([]MessageMention, 0, len(msg.Mentions))
		for _, mention := range msg.Mentions {
			out.Mentions = append(out.Mentions, mapMessageMention(mention))
		}
	}
	if msg.UpperMessageId != nil {
		out.UpperMessageID = *msg.UpperMessageId
	}
	return out
}

func mapMessageSender(sender *im.Sender) MessageSender {
	if sender == nil {
		return MessageSender{}
	}
	out := MessageSender{}
	if sender.Id != nil {
		out.ID = *sender.Id
	}
	if sender.IdType != nil {
		out.IDType = *sender.IdType
	}
	if sender.SenderType != nil {
		out.SenderType = *sender.SenderType
	}
	if sender.TenantKey != nil {
		out.TenantKey = *sender.TenantKey
	}
	return out
}

func mapMessageBody(body *im.MessageBody) MessageBody {
	if body == nil {
		return MessageBody{}
	}
	out := MessageBody{}
	if body.Content != nil {
		out.Content = *body.Content
	}
	return out
}

func mapMessageMention(mention *im.Mention) MessageMention {
	if mention == nil {
		return MessageMention{}
	}
	out := MessageMention{}
	if mention.Key != nil {
		out.Key = *mention.Key
	}
	if mention.Id != nil {
		out.ID = *mention.Id
	}
	if mention.IdType != nil {
		out.IDType = *mention.IdType
	}
	if mention.Name != nil {
		out.Name = *mention.Name
	}
	if mention.TenantKey != nil {
		out.TenantKey = *mention.TenantKey
	}
	return out
}
