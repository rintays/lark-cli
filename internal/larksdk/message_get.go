package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) GetMessage(ctx context.Context, userAccessToken, messageID, userIDType string) (Message, error) {
	if !c.available() {
		return Message{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return Message{}, errors.New("user access token is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return Message{}, errors.New("message id is required")
	}

	builder := im.NewGetMessageReqBuilder().MessageId(messageID)
	if strings.TrimSpace(userIDType) != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Im.V1.Message.Get(ctx, builder.Build(), larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return Message{}, err
	}
	if resp == nil {
		return Message{}, errors.New("get message failed: empty response")
	}
	if !resp.Success() {
		return Message{}, fmt.Errorf("get message failed: %s", resp.Msg)
	}
	if resp.Data != nil && len(resp.Data.Items) > 0 {
		return mapMessage(resp.Data.Items[0]), nil
	}
	return Message{}, errors.New("get message failed: empty result")
}
