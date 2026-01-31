package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type DocsSearchRequest struct {
	Query    string
	DocTypes []string
	Count    int
	Offset   int
	OwnerIDs []string
	ChatIDs  []string
}

type DocsSearchEntity struct {
	DocsToken string `json:"docs_token"`
	DocsType  string `json:"docs_type"`
	Title     string `json:"title"`
	OwnerID   string `json:"owner_id"`
	URL       string `json:"url,omitempty"`
	OpenURL   string `json:"open_url,omitempty"`
}

type DocsSearchResult struct {
	Entities []DocsSearchEntity
	HasMore  bool
	Total    int
}

type docsSearchResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *docsSearchResponseData `json:"data"`
}

type docsSearchResponseData struct {
	Entities []DocsSearchEntity `json:"docs_entities"`
	HasMore  *bool              `json:"has_more"`
	Total    *int               `json:"total"`
}

func (r *docsSearchResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchDocsObjectsWithUserToken(ctx context.Context, userAccessToken string, req DocsSearchRequest) (DocsSearchResult, error) {
	if !c.available() || c.coreConfig == nil {
		return DocsSearchResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return DocsSearchResult{}, errors.New("user access token is required")
	}
	if strings.TrimSpace(req.Query) == "" {
		return DocsSearchResult{}, errors.New("query is required")
	}
	if req.Count <= 0 {
		return DocsSearchResult{}, errors.New("count must be greater than 0")
	}
	if req.Offset < 0 {
		return DocsSearchResult{}, errors.New("offset must be greater than or equal to 0")
	}

	payload := map[string]any{
		"search_key": req.Query,
		"count":      req.Count,
		"offset":     req.Offset,
	}
	if len(req.DocTypes) > 0 {
		payload["docs_types"] = req.DocTypes
	}
	if len(req.OwnerIDs) > 0 {
		payload["owner_ids"] = req.OwnerIDs
	}
	if len(req.ChatIDs) > 0 {
		payload["chat_ids"] = req.ChatIDs
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/suite/docs-api/search/object",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
		Body:                      payload,
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(userAccessToken))
	if err != nil {
		return DocsSearchResult{}, err
	}
	if apiResp == nil {
		return DocsSearchResult{}, errors.New("docs search failed: empty response")
	}
	resp := &docsSearchResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DocsSearchResult{}, err
	}
	if !resp.Success() {
		return DocsSearchResult{}, fmt.Errorf("docs search failed (code=%d): %s", resp.Code, resp.Msg)
	}

	result := DocsSearchResult{}
	if resp.Data != nil {
		result.Entities = resp.Data.Entities
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
		if resp.Data.Total != nil {
			result.Total = *resp.Data.Total
		}
	}
	return result, nil
}
