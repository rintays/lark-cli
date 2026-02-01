package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type SearchUsersRequest struct {
	Query     string
	PageSize  int
	PageToken string
}

type SearchUsersResult struct {
	Users     []User
	PageToken string
	HasMore   bool
}

type userSearchResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *userSearchResponseData `json:"data"`
}

type userSearchResponseData struct {
	Users     []User  `json:"users"`
	PageToken *string `json:"page_token"`
	HasMore   *bool   `json:"has_more"`
}

func (r *userSearchResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchUsers(ctx context.Context, userAccessToken string, req SearchUsersRequest) (SearchUsersResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SearchUsersResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return SearchUsersResult{}, errors.New("user access token is required")
	}
	query := strings.TrimSpace(req.Query)
	if query == "" {
		return SearchUsersResult{}, errors.New("query is required")
	}
	if req.PageSize < 0 {
		return SearchUsersResult{}, errors.New("page_size must be greater than or equal to 0")
	}
	if req.PageSize > 200 {
		return SearchUsersResult{}, errors.New("page_size must be less than or equal to 200")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/search/v1/user",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
	}
	apiReq.QueryParams.Set("query", query)
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprint(req.PageSize))
	}
	if strings.TrimSpace(req.PageToken) != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return SearchUsersResult{}, err
	}
	if apiResp == nil {
		return SearchUsersResult{}, errors.New("search users failed: empty response")
	}
	resp := &userSearchResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SearchUsersResult{}, err
	}
	if !resp.Success() {
		return SearchUsersResult{}, fmt.Errorf("search users failed (code=%d): %s", resp.Code, resp.Msg)
	}

	result := SearchUsersResult{}
	if resp.Data != nil {
		if resp.Data.Users != nil {
			result.Users = resp.Data.Users
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
