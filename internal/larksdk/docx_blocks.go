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
			// Prefer the root page block.
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
		BlockType(2). // 2 = text/paragraph block
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
