package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

func (c *Client) docxBlocksSDKAvailable() bool {
	return c != nil && c.sdk != nil &&
		c.sdk.Docx != nil && c.sdk.Docx.V1 != nil &&
		c.sdk.Docx.V1.DocumentBlock != nil &&
		c.sdk.Docx.V1.DocumentBlockChildren != nil
}

func (c *Client) docxBlockSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Docx != nil && c.sdk.Docx.V1 != nil && c.sdk.Docx.V1.DocumentBlock != nil
}

func (c *Client) docxBlockChildrenSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Docx != nil && c.sdk.Docx.V1 != nil && c.sdk.Docx.V1.DocumentBlockChildren != nil
}

func (c *Client) docxBlockDescendantSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Docx != nil && c.sdk.Docx.V1 != nil && c.sdk.Docx.V1.DocumentBlockDescendant != nil
}

// AppendDocxTextBlock appends a plain text paragraph into a Docx document.
//
// This is primarily used by integration tests to create a predictable piece of
// content that can be asserted via export/cat.
func (c *Client) AppendDocxTextBlock(ctx context.Context, token, documentID, text string) (string, error) {
	if !c.available() {
		return "", ErrUnavailable
	}
	if strings.TrimSpace(documentID) == "" {
		return "", errors.New("document id is required")
	}
	if strings.TrimSpace(text) == "" {
		return "", errors.New("text is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	if !c.docxBlocksSDKAvailable() {
		return "", ErrUnavailable
	}

	listResp, err := c.sdk.Docx.V1.DocumentBlock.List(
		ctx,
		larkdocx.NewListDocumentBlockReqBuilder().DocumentId(documentID).PageSize(200).Build(),
		larkcore.WithTenantAccessToken(tenantToken),
	)
	if err != nil {
		return "", err
	}
	if listResp == nil {
		return "", errors.New("list docx blocks failed: empty response")
	}
	if !listResp.Success() {
		return "", fmt.Errorf("list docx blocks failed: %s", listResp.Msg)
	}

	var parentBlockID string
	if listResp.Data != nil {
		for _, block := range listResp.Data.Items {
			if block == nil || block.BlockId == nil || *block.BlockId == "" {
				continue
			}
			if block.Page != nil {
				parentBlockID = *block.BlockId
				break
			}
		}
		if parentBlockID == "" {
			for _, block := range listResp.Data.Items {
				if block == nil || block.BlockId == nil || *block.BlockId == "" {
					continue
				}
				parentBlockID = *block.BlockId
				break
			}
		}
	}
	if parentBlockID == "" {
		return "", errors.New("unable to find a parent block to append content")
	}

	child := larkdocx.NewBlockBuilder().
		BlockType(2).
		Text(larkdocx.NewTextBuilder().
			Elements([]*larkdocx.TextElement{
				larkdocx.NewTextElementBuilder().
					TextRun(larkdocx.NewTextRunBuilder().Content(text).Build()).
					Build(),
			}).
			Build(),
		).
		Build()

	body := larkdocx.NewCreateDocumentBlockChildrenReqBodyBuilder().
		Children([]*larkdocx.Block{child}).
		Index(0).
		Build()

	createResp, err := c.sdk.Docx.V1.DocumentBlockChildren.Create(
		ctx,
		larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
			DocumentId(documentID).
			BlockId(parentBlockID).
			Body(body).
			Build(),
		larkcore.WithTenantAccessToken(tenantToken),
	)
	if err != nil {
		return "", err
	}
	if createResp == nil {
		return "", errors.New("create docx block children failed: empty response")
	}
	if !createResp.Success() {
		return "", fmt.Errorf("create docx block children failed: %s", createResp.Msg)
	}
	if createResp.Data == nil || len(createResp.Data.Children) == 0 || createResp.Data.Children[0] == nil {
		return "", nil
	}
	if createResp.Data.Children[0].BlockId == nil {
		return "", nil
	}
	return *createResp.Data.Children[0].BlockId, nil
}

func (c *Client) GetDocxBlock(ctx context.Context, token, documentID, blockID string, revisionID int, userIDType string) (*larkdocx.Block, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, errors.New("block id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockSDKAvailable() {
		return nil, errors.New("docx block sdk unavailable")
	}

	builder := larkdocx.NewGetDocumentBlockReqBuilder().DocumentId(documentID).BlockId(blockID)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlock.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("get docx block failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("get docx block failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return nil, nil
	}
	return resp.Data.Block, nil
}

func (c *Client) ListDocxBlocks(ctx context.Context, token, documentID string, pageSize int, pageToken string, revisionID int, userIDType string) ([]*larkdocx.Block, string, bool, error) {
	if !c.available() {
		return nil, "", false, ErrUnavailable
	}
	if documentID == "" {
		return nil, "", false, errors.New("document id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, "", false, errors.New("tenant access token is required")
	}
	if !c.docxBlockSDKAvailable() {
		return nil, "", false, errors.New("docx block sdk unavailable")
	}

	builder := larkdocx.NewListDocumentBlockReqBuilder().DocumentId(documentID)
	if pageSize > 0 {
		builder.PageSize(pageSize)
	}
	if pageToken != "" {
		builder.PageToken(pageToken)
	}
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlock.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, "", false, err
	}
	if resp == nil {
		return nil, "", false, errors.New("list docx blocks failed: empty response")
	}
	if !resp.Success() {
		return nil, "", false, fmt.Errorf("list docx blocks failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return nil, "", false, nil
	}
	nextToken := ""
	if resp.Data.PageToken != nil {
		nextToken = *resp.Data.PageToken
	}
	hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
	return resp.Data.Items, nextToken, hasMore, nil
}

func (c *Client) GetDocxBlockChildren(ctx context.Context, token, documentID, blockID string, pageSize int, pageToken string, revisionID int, withDescendants bool, userIDType string) ([]*larkdocx.Block, string, bool, error) {
	if !c.available() {
		return nil, "", false, ErrUnavailable
	}
	if documentID == "" {
		return nil, "", false, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, "", false, errors.New("block id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, "", false, errors.New("tenant access token is required")
	}
	if !c.docxBlockChildrenSDKAvailable() {
		return nil, "", false, errors.New("docx block children sdk unavailable")
	}

	builder := larkdocx.NewGetDocumentBlockChildrenReqBuilder().DocumentId(documentID).BlockId(blockID)
	if pageSize > 0 {
		builder.PageSize(pageSize)
	}
	if pageToken != "" {
		builder.PageToken(pageToken)
	}
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if withDescendants {
		builder.WithDescendants(withDescendants)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlockChildren.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, "", false, err
	}
	if resp == nil {
		return nil, "", false, errors.New("get docx block children failed: empty response")
	}
	if !resp.Success() {
		return nil, "", false, fmt.Errorf("get docx block children failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return nil, "", false, nil
	}
	nextToken := ""
	if resp.Data.PageToken != nil {
		nextToken = *resp.Data.PageToken
	}
	hasMore := resp.Data.HasMore != nil && *resp.Data.HasMore
	return resp.Data.Items, nextToken, hasMore, nil
}

func (c *Client) CreateDocxBlockChildren(ctx context.Context, token, documentID, blockID string, body *larkdocx.CreateDocumentBlockChildrenReqBody, revisionID int, clientToken, userIDType string) (*larkdocx.CreateDocumentBlockChildrenRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, errors.New("block id is required")
	}
	if body == nil {
		return nil, errors.New("request body is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockChildrenSDKAvailable() {
		return nil, errors.New("docx block children sdk unavailable")
	}

	builder := larkdocx.NewCreateDocumentBlockChildrenReqBuilder().
		DocumentId(documentID).
		BlockId(blockID).
		Body(body)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlockChildren.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("create docx block children failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("create docx block children failed: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *Client) CreateDocxBlockDescendant(ctx context.Context, token, documentID, blockID string, body *larkdocx.CreateDocumentBlockDescendantReqBody, revisionID int, clientToken, userIDType string) (*larkdocx.CreateDocumentBlockDescendantRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, errors.New("block id is required")
	}
	if body == nil {
		return nil, errors.New("request body is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockDescendantSDKAvailable() {
		return nil, errors.New("docx block descendant sdk unavailable")
	}

	builder := larkdocx.NewCreateDocumentBlockDescendantReqBuilder().
		DocumentId(documentID).
		BlockId(blockID).
		Body(body)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlockDescendant.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("create docx block descendant failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("create docx block descendant failed: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *Client) PatchDocxBlock(ctx context.Context, token, documentID, blockID string, update *larkdocx.UpdateBlockRequest, revisionID int, clientToken, userIDType string) (*larkdocx.PatchDocumentBlockRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, errors.New("block id is required")
	}
	if update == nil {
		return nil, errors.New("update request is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockSDKAvailable() {
		return nil, errors.New("docx block sdk unavailable")
	}

	builder := larkdocx.NewPatchDocumentBlockReqBuilder().
		DocumentId(documentID).
		BlockId(blockID).
		UpdateBlockRequest(update)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlock.Patch(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("patch docx block failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("patch docx block failed: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *Client) BatchUpdateDocxBlocks(ctx context.Context, token, documentID string, requests []*larkdocx.UpdateBlockRequest, revisionID int, clientToken, userIDType string) (*larkdocx.BatchUpdateDocumentBlockRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if len(requests) == 0 {
		return nil, errors.New("update requests are required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockSDKAvailable() {
		return nil, errors.New("docx block sdk unavailable")
	}

	body := larkdocx.NewBatchUpdateDocumentBlockReqBodyBuilder().Requests(requests).Build()
	builder := larkdocx.NewBatchUpdateDocumentBlockReqBuilder().
		DocumentId(documentID).
		Body(body)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}
	if userIDType != "" {
		builder.UserIdType(userIDType)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlock.BatchUpdate(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch update docx blocks failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("batch update docx blocks failed: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *Client) BatchDeleteDocxBlockChildren(ctx context.Context, token, documentID, blockID string, startIndex, endIndex int, revisionID int, clientToken string) (*larkdocx.BatchDeleteDocumentBlockChildrenRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if documentID == "" {
		return nil, errors.New("document id is required")
	}
	if blockID == "" {
		return nil, errors.New("block id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxBlockChildrenSDKAvailable() {
		return nil, errors.New("docx block children sdk unavailable")
	}

	body := larkdocx.NewBatchDeleteDocumentBlockChildrenReqBodyBuilder().
		StartIndex(startIndex).
		EndIndex(endIndex).
		Build()
	builder := larkdocx.NewBatchDeleteDocumentBlockChildrenReqBuilder().
		DocumentId(documentID).
		BlockId(blockID).
		Body(body)
	if revisionID != 0 {
		builder.DocumentRevisionId(revisionID)
	}
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}

	resp, err := c.sdk.Docx.V1.DocumentBlockChildren.BatchDelete(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch delete docx block children failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("batch delete docx block children failed: %s", resp.Msg)
	}
	return resp.Data, nil
}

func (c *Client) ConvertDocxContent(ctx context.Context, token, contentType, content string) (*larkdocx.ConvertDocumentRespData, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if contentType == "" {
		return nil, errors.New("content type is required")
	}
	if content == "" {
		return nil, errors.New("content is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if !c.docxSDKAvailable() {
		return nil, errors.New("docx sdk unavailable")
	}

	body := larkdocx.NewConvertDocumentReqBodyBuilder().
		ContentType(contentType).
		Content(content).
		Build()
	builder := larkdocx.NewConvertDocumentReqBuilder().Body(body)

	resp, err := c.sdk.Docx.V1.Document.Convert(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("convert docx content failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("convert docx content failed: %s", resp.Msg)
	}
	return resp.Data, nil
}
