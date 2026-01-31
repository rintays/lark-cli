package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

type createDocxDocumentResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createDocxDocumentResponseData `json:"data"`
}

type createDocxDocumentResponseData struct {
	Document *DocxDocument `json:"document"`
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
	Document *DocxDocument `json:"document"`
}

func (r *getDocxDocumentResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) CreateDocxDocument(ctx context.Context, token string, req CreateDocxDocumentRequest) (DocxDocument, error) {
	if !c.available() {
		return DocxDocument{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DocxDocument{}, errors.New("tenant access token is required")
	}
	if c.docxSDKAvailable() {
		return c.createDocxDocumentSDK(ctx, tenantToken, req)
	}
	if c.coreConfig == nil {
		return DocxDocument{}, ErrUnavailable
	}
	return c.createDocxDocumentCore(ctx, tenantToken, req)
}

func (c *Client) createDocxDocumentSDK(ctx context.Context, tenantToken string, req CreateDocxDocumentRequest) (DocxDocument, error) {
	bodyBuilder := larkdocx.NewCreateDocumentReqBodyBuilder().Title(req.Title)
	if req.FolderToken != "" {
		bodyBuilder = bodyBuilder.FolderToken(req.FolderToken)
	}
	body := bodyBuilder.Build()
	resp, err := c.sdk.Docx.V1.Document.Create(
		ctx,
		larkdocx.NewCreateDocumentReqBuilder().Body(body).Build(),
		larkcore.WithTenantAccessToken(tenantToken),
	)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp == nil {
		return DocxDocument{}, errors.New("create docx document failed: empty response")
	}
	if !resp.Success() {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", resp.Msg)
	}
	if resp.ApiResp != nil {
		raw := &createDocxDocumentResponse{ApiResp: resp.ApiResp}
		if err := json.Unmarshal(resp.ApiResp.RawBody, raw); err != nil {
			return DocxDocument{}, err
		}
		if raw.Data == nil || raw.Data.Document == nil {
			return DocxDocument{}, nil
		}
		return *raw.Data.Document, nil
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return DocxDocument{}, nil
	}
	return mapDocxDocument(resp.Data.Document), nil
}

func (c *Client) createDocxDocumentCore(ctx context.Context, tenantToken string, req CreateDocxDocumentRequest) (DocxDocument, error) {
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
		return DocxDocument{}, err
	}
	if apiResp == nil {
		return DocxDocument{}, errors.New("create docx document failed: empty response")
	}
	resp := &createDocxDocumentResponse{ApiResp: apiResp}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return DocxDocument{}, err
	}
	if !resp.Success() {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return DocxDocument{}, nil
	}
	return *resp.Data.Document, nil
}

func (c *Client) GetDocxDocument(ctx context.Context, token, documentID string) (DocxDocument, error) {
	if !c.available() {
		return DocxDocument{}, ErrUnavailable
	}
	if documentID == "" {
		return DocxDocument{}, errors.New("document id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DocxDocument{}, errors.New("tenant access token is required")
	}
	if c.docxSDKAvailable() {
		return c.getDocxDocumentSDK(ctx, tenantToken, documentID)
	}
	if c.coreConfig == nil {
		return DocxDocument{}, ErrUnavailable
	}
	return c.getDocxDocumentCore(ctx, tenantToken, documentID)
}

func (c *Client) getDocxDocumentSDK(ctx context.Context, tenantToken, documentID string) (DocxDocument, error) {
	resp, err := c.sdk.Docx.V1.Document.Get(
		ctx,
		larkdocx.NewGetDocumentReqBuilder().DocumentId(documentID).Build(),
		larkcore.WithTenantAccessToken(tenantToken),
	)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp == nil {
		return DocxDocument{}, errors.New("get docx document failed: empty response")
	}
	if !resp.Success() {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", resp.Msg)
	}
	if resp.ApiResp != nil {
		raw := &getDocxDocumentResponse{ApiResp: resp.ApiResp}
		if err := json.Unmarshal(resp.ApiResp.RawBody, raw); err != nil {
			return DocxDocument{}, err
		}
		if raw.Data == nil || raw.Data.Document == nil {
			return DocxDocument{}, nil
		}
		return *raw.Data.Document, nil
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return DocxDocument{}, nil
	}
	return mapDocxDocument(resp.Data.Document), nil
}

func (c *Client) getDocxDocumentCore(ctx context.Context, tenantToken, documentID string) (DocxDocument, error) {
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
		return DocxDocument{}, err
	}
	if apiResp == nil {
		return DocxDocument{}, errors.New("get docx document failed: empty response")
	}
	resp := &getDocxDocumentResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DocxDocument{}, err
	}
	if !resp.Success() {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Document == nil {
		return DocxDocument{}, nil
	}
	return *resp.Data.Document, nil
}

type rawContentDocxResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *rawContentDocxResponseData `json:"data"`
}

type rawContentDocxResponseData struct {
	Content string `json:"content"`
}

func (r *rawContentDocxResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) GetDocxRawContent(ctx context.Context, token, documentID string) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	if documentID == "" {
		return "", errors.New("document id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	if c.docxSDKAvailable() {
		return c.getDocxRawContentSDK(ctx, tenantToken, documentID)
	}
	if c.coreConfig == nil {
		return "", ErrUnavailable
	}
	return c.getDocxRawContentCore(ctx, tenantToken, documentID)
}

func (c *Client) getDocxRawContentSDK(ctx context.Context, tenantToken, documentID string) (string, error) {
	resp, err := c.sdk.Docx.V1.Document.RawContent(
		ctx,
		larkdocx.NewRawContentDocumentReqBuilder().DocumentId(documentID).Build(),
		larkcore.WithTenantAccessToken(tenantToken),
	)
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("get docx raw content failed: empty response")
	}
	if !resp.Success() {
		return "", fmt.Errorf("get docx raw content failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Content == nil {
		return "", nil
	}
	return *resp.Data.Content, nil
}

func (c *Client) getDocxRawContentCore(ctx context.Context, tenantToken, documentID string) (string, error) {
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/docx/v1/documents/:document_id/raw_content",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("document_id", documentID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("get docx raw content failed: empty response")
	}
	resp := &rawContentDocxResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("get docx raw content failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return "", nil
	}
	return resp.Data.Content, nil
}

func (c *Client) docxSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Docx != nil && c.sdk.Docx.V1 != nil && c.sdk.Docx.V1.Document != nil
}

func mapDocxDocument(document *larkdocx.Document) DocxDocument {
	if document == nil {
		return DocxDocument{}
	}
	doc := DocxDocument{}
	if document.DocumentId != nil {
		doc.DocumentID = *document.DocumentId
	}
	if document.Title != nil {
		doc.Title = *document.Title
	}
	if document.RevisionId != nil {
		doc.RevisionID = RevisionID(fmt.Sprintf("%d", *document.RevisionId))
	}
	return doc
}
