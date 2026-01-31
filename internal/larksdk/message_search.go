package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type messageSearchResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *messageSearchResponseData `json:"data"`
}

type messageSearchResponseData struct {
	Items     []string `json:"items"`
	PageToken *string  `json:"page_token"`
	HasMore   *bool    `json:"has_more"`
}

func (r *messageSearchResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchMessages(ctx context.Context, userAccessToken string, req MessageSearchRequest) (MessageSearchResult, error) {
	if !c.available() || c.coreConfig == nil {
		return MessageSearchResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return MessageSearchResult{}, errors.New("user access token is required")
	}
	if strings.TrimSpace(req.Query) == "" {
		return MessageSearchResult{}, errors.New("query is required")
	}

	payload := map[string]any{
		"query": req.Query,
	}
	if len(req.FromIDs) > 0 {
		payload["from_ids"] = req.FromIDs
	}
	if len(req.ChatIDs) > 0 {
		payload["chat_ids"] = req.ChatIDs
	}
	if strings.TrimSpace(req.MessageType) != "" {
		payload["message_type"] = req.MessageType
	}
	if len(req.AtChatterIDs) > 0 {
		payload["at_chatter_ids"] = req.AtChatterIDs
	}
	if strings.TrimSpace(req.FromType) != "" {
		payload["from_type"] = req.FromType
	}
	if strings.TrimSpace(req.ChatType) != "" {
		payload["chat_type"] = req.ChatType
	}
	if strings.TrimSpace(req.StartTime) != "" {
		payload["start_time"] = req.StartTime
	}
	if strings.TrimSpace(req.EndTime) != "" {
		payload["end_time"] = req.EndTime
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/search/v2/message",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
		Body:                      payload,
	}
	if strings.TrimSpace(req.UserIDType) != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprint(req.PageSize))
	}
	if strings.TrimSpace(req.PageToken) != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return MessageSearchResult{}, err
	}
	if apiResp == nil {
		return MessageSearchResult{}, errors.New("message search failed: empty response")
	}
	resp := &messageSearchResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return MessageSearchResult{}, err
	}
	if !resp.Success() {
		return MessageSearchResult{}, fmt.Errorf("message search failed (code=%d): %s", resp.Code, resp.Msg)
	}

	result := MessageSearchResult{}
	if resp.Data != nil {
		result.Items = resp.Data.Items
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}
