package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type WikiNodeSearchV1Request struct {
	Query     string
	SpaceID   string
	PageSize  int
	PageToken string
}

type WikiNodeSearchV1Item struct {
	NodeToken string `json:"node_token"`
	Title     string `json:"title"`
	ObjType   string `json:"obj_type"`
	URL       string `json:"url"`
}

type WikiNodeSearchV1Result struct {
	Items     []WikiNodeSearchV1Item
	HasMore   bool
	PageToken string
}

type wikiNodeSearchV1Response struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *wikiNodeSearchV1ResponseData `json:"data"`
}

type wikiNodeSearchV1ResponseData struct {
	Items     []WikiNodeSearchV1Item `json:"items"`
	HasMore   *bool                  `json:"has_more"`
	PageToken *string                `json:"page_token"`
}

func (r *wikiNodeSearchV1Response) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchWikiNodesV1(ctx context.Context, userAccessToken string, req WikiNodeSearchV1Request) (WikiNodeSearchV1Result, error) {
	if !c.available() || c.coreConfig == nil {
		return WikiNodeSearchV1Result{}, ErrUnavailable
	}
	if userAccessToken == "" {
		return WikiNodeSearchV1Result{}, errors.New("user access token is required")
	}
	if req.Query == "" {
		return WikiNodeSearchV1Result{}, errors.New("query is required")
	}

	payload := map[string]any{
		"query": req.Query,
	}
	if req.SpaceID != "" {
		payload["space_id"] = req.SpaceID
	}
	if req.PageSize > 0 {
		payload["page_size"] = req.PageSize
	}
	if req.PageToken != "" {
		payload["page_token"] = req.PageToken
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/wiki/v1/nodes/search",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
		Body:                      payload,
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return WikiNodeSearchV1Result{}, err
	}
	if apiResp == nil {
		return WikiNodeSearchV1Result{}, errors.New("wiki node search failed: empty response")
	}
	resp := &wikiNodeSearchV1Response{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return WikiNodeSearchV1Result{}, err
	}
	if !resp.Success() {
		baseErr := fmt.Errorf("wiki node search failed: %s", resp.Msg)
		return WikiNodeSearchV1Result{}, withInsufficientScopeRemediation(baseErr, resp.Msg)
	}

	result := WikiNodeSearchV1Result{}
	if resp.Data != nil {
		result.Items = resp.Data.Items
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
	}
	return result, nil
}
