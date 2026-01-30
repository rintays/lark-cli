package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkminutes "github.com/larksuite/oapi-sdk-go/v3/service/minutes/v1"

	"lark/internal/larkapi"
)

type listMinutesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMinutesResponseData `json:"data"`
}

type listMinutesResponseData struct {
	Items     []*larkminutes.Minute `json:"items"`
	PageToken *string               `json:"page_token"`
	HasMore   *bool                 `json:"has_more"`
}

func (r *listMinutesResponse) Success() bool {
	return r.Code == 0
}

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

func (c *Client) ListMinutes(ctx context.Context, token string, req larkapi.ListMinutesRequest) (larkapi.ListMinutesResult, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.ListMinutesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.ListMinutesResult{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/minutes/v1/minutes",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.ListMinutesResult{}, err
	}
	if apiResp == nil {
		return larkapi.ListMinutesResult{}, errors.New("list minutes failed: empty response")
	}
	resp := &listMinutesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.ListMinutesResult{}, err
	}
	if !resp.Success() {
		return larkapi.ListMinutesResult{}, fmt.Errorf("list minutes failed: %s", resp.Msg)
	}

	result := larkapi.ListMinutesResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]larkapi.Minute, 0, len(resp.Data.Items))
			for _, minute := range resp.Data.Items {
				result.Items = append(result.Items, mapMinute(minute))
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
