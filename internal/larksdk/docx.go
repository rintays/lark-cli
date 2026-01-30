package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"lark/internal/larkapi"
)

type createDocxDocumentResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createDocxDocumentResponseData `json:"data"`
}

type createDocxDocumentResponseData struct {
	Document *larkapi.DocxDocument `json:"document"`
}

func (r *createDocxDocumentResponse) Success() bool {
	return r.Code == 0
}

type getDocxDocumentResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getDocxDocumentResponseData `json:"data"`
}

type getDocxDocumentResponseData struct {
	Document *larkapi.DocxDocument `json:"document"`
}

func (r *getDocxDocumentResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) CreateDocxDocument(ctx context.Context, token string, req larkapi.CreateDocxDocumentRequest) (larkapi.DocxDocument, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.DocxDocument{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.DocxDocument{}, errors.New("tenant access token is required")
	}

	payload := map[string]any{
		"title": req.Title,
	}
	if req.FolderToken != "" {
		payload["folder_token"] = req.FolderToken
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/docx/v1/documents",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.Body = payload

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.DocxDocument{}, err
	}
	if apiResp == nil {
		return larkapi.DocxDocument{}, errors.New("create docx document failed: empty response")
	}
	resp := &createDocxDocumentResponse{ApiResp: apiResp}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return larkapi.DocxDocument{}, err
	}
	if !resp.Success() {
		return larkapi.DocxDocument{}, fmt.Errorf("create docx document failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return larkapi.DocxDocument{}, nil
	}
	return *resp.Data.Document, nil
}

func (c *Client) GetDocxDocument(ctx context.Context, token, documentID string) (larkapi.DocxDocument, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.DocxDocument{}, ErrUnavailable
	}
	if documentID == "" {
		return larkapi.DocxDocument{}, errors.New("document id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.DocxDocument{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/docx/v1/documents/:document_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("document_id", documentID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.DocxDocument{}, err
	}
	if apiResp == nil {
		return larkapi.DocxDocument{}, errors.New("get docx document failed: empty response")
	}
	resp := &getDocxDocumentResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.DocxDocument{}, err
	}
	if !resp.Success() {
		return larkapi.DocxDocument{}, fmt.Errorf("get docx document failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return larkapi.DocxDocument{}, nil
	}
	return *resp.Data.Document, nil
}
