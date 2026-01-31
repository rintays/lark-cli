package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkminutes "github.com/larksuite/oapi-sdk-go/v3/service/minutes/v1"
)

func (c *Client) GetMinute(ctx context.Context, token, minuteToken, userIDType string) (Minute, error) {
	if !c.available() {
		return Minute{}, ErrUnavailable
	}
	if minuteToken == "" {
		return Minute{}, errors.New("minute token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Minute{}, errors.New("tenant access token is required")
	}

	builder := larkminutes.NewGetMinuteReqBuilder().MinuteToken(minuteToken)
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Minutes.V1.Minute.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return Minute{}, err
	}
	if resp == nil {
		return Minute{}, errors.New("get minute failed: empty response")
	}
	if !resp.Success() {
		return Minute{}, fmt.Errorf("get minute failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Minute == nil {
		return Minute{}, nil
	}
	return mapMinute(resp.Data.Minute), nil
}

func mapMinute(minute *larkminutes.Minute) Minute {
	if minute == nil {
		return Minute{}
	}
	result := Minute{}
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
