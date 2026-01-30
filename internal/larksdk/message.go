package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"lark/internal/larkapi"
)

func (c *Client) SendMessage(ctx context.Context, token string, req larkapi.MessageRequest) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	content, err := json.Marshal(map[string]string{"text": req.Text})
	if err != nil {
		return "", err
	}

	receiveIDType := req.ReceiveIDType
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}

	body := im.NewCreateMessageReqBodyBuilder().
		ReceiveId(req.ReceiveID).
		MsgType("text").
		Content(string(content)).
		Build()
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
