package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"
)

func (c *Client) PinMessage(ctx context.Context, token string, messageID string) (Pin, error) {
	if !c.available() {
		return Pin{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Pin{}, errors.New("tenant access token is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return Pin{}, errors.New("message id is required")
	}

	body := im.NewCreatePinReqBodyBuilder().MessageId(messageID).Build()
	builder := im.NewCreatePinReqBuilder().Body(body)
	resp, err := c.sdk.Im.V1.Pin.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return Pin{}, err
	}
	if resp == nil {
		return Pin{}, errors.New("pin message failed: empty response")
	}
	if !resp.Success() {
		return Pin{}, fmt.Errorf("pin message failed: %s", resp.Msg)
	}
	if resp.Data != nil && resp.Data.Pin != nil {
		return mapPin(resp.Data.Pin), nil
	}
	return Pin{}, nil
}

func (c *Client) UnpinMessage(ctx context.Context, token string, messageID string) error {
	if !c.available() {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	messageID = strings.TrimSpace(messageID)
	if messageID == "" {
		return errors.New("message id is required")
	}

	builder := im.NewDeletePinReqBuilder().MessageId(messageID)
	resp, err := c.sdk.Im.V1.Pin.Delete(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("unpin message failed: empty response")
	}
	if !resp.Success() {
		return fmt.Errorf("unpin message failed: %s", resp.Msg)
	}
	return nil
}

func mapPin(pin *im.Pin) Pin {
	if pin == nil {
		return Pin{}
	}
	out := Pin{}
	if pin.MessageId != nil {
		out.MessageID = *pin.MessageId
	}
	if pin.ChatId != nil {
		out.ChatID = *pin.ChatId
	}
	if pin.OperatorId != nil {
		out.OperatorID = *pin.OperatorId
	}
	if pin.OperatorIdType != nil {
		out.OperatorIDType = *pin.OperatorIdType
	}
	if pin.CreateTime != nil {
		out.CreateTime = *pin.CreateTime
	}
	return out
}
