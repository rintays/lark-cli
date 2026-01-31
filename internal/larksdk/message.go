package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) SendMessage(ctx context.Context, token string, req MessageRequest) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	if strings.TrimSpace(req.ReceiveID) == "" {
		return "", errors.New("receive id is required")
	}
	msgType, content, err := normalizeMessageContent(req.MsgType, req.Content, req.Text)
	if err != nil {
		return "", err
	}

	receiveIDType := req.ReceiveIDType
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}

	bodyBuilder := im.NewCreateMessageReqBodyBuilder().
		ReceiveId(req.ReceiveID).
		MsgType(msgType).
		Content(content)
	if strings.TrimSpace(req.UUID) != "" {
		bodyBuilder.Uuid(strings.TrimSpace(req.UUID))
	}
	body := bodyBuilder.Build()
	builder := im.NewCreateMessageReqBuilder().
		ReceiveIdType(receiveIDType).
		Body(body)

	resp, err := c.sdk.Im.V1.Message.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("send message failed: empty response")
	}
	if !resp.Success() {
		return "", fmt.Errorf("send message failed: %s", resp.Msg)
	}
	if resp.Data != nil && resp.Data.MessageId != nil {
		return *resp.Data.MessageId, nil
	}
	return "", nil
}

type ReplyMessageRequest struct {
	MessageID     string
	MsgType       string
	Content       string
	Text          string
	ReplyInThread bool
	UUID          string
}

func (c *Client) ReplyMessage(ctx context.Context, token string, req ReplyMessageRequest) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	messageID := strings.TrimSpace(req.MessageID)
	if messageID == "" {
		return "", errors.New("message id is required")
	}
	msgType, content, err := normalizeMessageContent(req.MsgType, req.Content, req.Text)
	if err != nil {
		return "", err
	}

	bodyBuilder := im.NewReplyMessageReqBodyBuilder().
		MsgType(msgType).
		Content(content).
		ReplyInThread(req.ReplyInThread)
	if strings.TrimSpace(req.UUID) != "" {
		bodyBuilder.Uuid(strings.TrimSpace(req.UUID))
	}

	builder := im.NewReplyMessageReqBuilder().
		MessageId(messageID).
		Body(bodyBuilder.Build())

	resp, err := c.sdk.Im.V1.Message.Reply(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("reply message failed: empty response")
	}
	if !resp.Success() {
		return "", fmt.Errorf("reply message failed: %s", resp.Msg)
	}
	if resp.Data != nil && resp.Data.MessageId != nil {
		return *resp.Data.MessageId, nil
	}
	return "", nil
}

func normalizeMessageContent(msgType, content, text string) (string, string, error) {
	msgType = strings.TrimSpace(msgType)
	content = strings.TrimSpace(content)
	text = strings.TrimSpace(text)
	if msgType == "" && content != "" {
		return "", "", errors.New("msg_type is required when content is provided")
	}
	if msgType == "" {
		msgType = "text"
	}
	if content == "" && text == "" {
		return "", "", errors.New("content or text is required")
	}
	if content == "" {
		if msgType != "text" {
			return "", "", fmt.Errorf("content is required for msg_type %q", msgType)
		}
		raw, err := json.Marshal(map[string]string{"text": text})
		if err != nil {
			return "", "", err
		}
		content = string(raw)
	}
	return msgType, content, nil
}
