package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkminutes "github.com/larksuite/oapi-sdk-go/v3/service/minutes/v1"

	"lark/internal/larkapi"
)

func (c *Client) GetMinute(ctx context.Context, token, minuteToken, userIDType string) (larkapi.Minute, error) {
	if !c.available() {
		return larkapi.Minute{}, ErrUnavailable
	}
	if minuteToken == "" {
		return larkapi.Minute{}, errors.New("minute token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.Minute{}, errors.New("tenant access token is required")
	}

	builder := larkminutes.NewGetMinuteReqBuilder().MinuteToken(minuteToken)
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Minutes.V1.Minute.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.Minute{}, err
	}
	if resp == nil {
		return larkapi.Minute{}, errors.New("get minute failed: empty response")
	}
	if !resp.Success() {
		return larkapi.Minute{}, fmt.Errorf("get minute failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Minute == nil {
		return larkapi.Minute{}, nil
	}
	return mapMinute(resp.Data.Minute), nil
}

func mapMinute(minute *larkminutes.Minute) larkapi.Minute {
	if minute == nil {
		return larkapi.Minute{}
	}
	result := larkapi.Minute{}
	if minute.Token != nil {
		result.Token = *minute.Token
	}
	if minute.OwnerId != nil {
		result.OwnerID = *minute.OwnerId
	}
	if minute.CreateTime != nil {
		result.CreateTime = *minute.CreateTime
	}
	if minute.Title != nil {
		result.Title = *minute.Title
	}
	if minute.Cover != nil {
		result.Cover = *minute.Cover
	}
	if minute.Duration != nil {
		result.Duration = *minute.Duration
	}
	if minute.Url != nil {
		result.URL = *minute.Url
	}
	return result
}
