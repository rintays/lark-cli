package larksdk

import (
	"context"
	"errors"
	"fmt"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
)

type CreateWikiNodeRequest struct {
	SpaceID         string
	ObjType         string
	ObjToken        string
	ParentNodeToken string
	NodeType        string
	OriginNodeToken string
	OriginSpaceID   string
	Title           string
}

func (c *Client) CreateWikiNodeV2(ctx context.Context, token string, req CreateWikiNodeRequest) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiNode{}, errors.New("tenant access token is required")
	}
	return c.createWikiNodeV2(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) CreateWikiNodeV2WithUserToken(ctx context.Context, userAccessToken string, req CreateWikiNodeRequest) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return WikiNode{}, errors.New("user access token is required")
	}
	return c.createWikiNodeV2(ctx, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) createWikiNodeV2(ctx context.Context, req CreateWikiNodeRequest, option larkcore.RequestOptionFunc) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	spaceID := strings.TrimSpace(req.SpaceID)
	if spaceID == "" {
		return WikiNode{}, errors.New("space id is required")
	}
	objType := strings.TrimSpace(req.ObjType)
	if objType == "" {
		return WikiNode{}, errors.New("obj type is required")
	}
	objToken := strings.TrimSpace(req.ObjToken)
	originNodeToken := strings.TrimSpace(req.OriginNodeToken)
	if objToken == "" && originNodeToken == "" {
		return WikiNode{}, errors.New("obj token or origin node token is required")
	}

	node := &larkwiki.Node{}
	node.ObjType = &objType
	if objToken != "" {
		node.ObjToken = &objToken
	}
	if parent := strings.TrimSpace(req.ParentNodeToken); parent != "" {
		node.ParentNodeToken = &parent
	}
	if nodeType := strings.TrimSpace(req.NodeType); nodeType != "" {
		node.NodeType = &nodeType
	}
	if originNodeToken != "" {
		node.OriginNodeToken = &originNodeToken
	}
	if originSpaceID := strings.TrimSpace(req.OriginSpaceID); originSpaceID != "" {
		node.OriginSpaceId = &originSpaceID
	}
	if title := strings.TrimSpace(req.Title); title != "" {
		node.Title = &title
	}

	builder := larkwiki.NewCreateSpaceNodeReqBuilder().SpaceId(spaceID).Node(node)
	resp, err := c.sdk.Wiki.V2.SpaceNode.Create(ctx, builder.Build(), option)
	if err != nil {
		return WikiNode{}, err
	}
	if resp == nil {
		return WikiNode{}, errors.New("wiki node create failed: empty response")
	}
	if !resp.Success() {
		return WikiNode{}, fmt.Errorf("wiki node create failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Node == nil {
		return WikiNode{}, errors.New("wiki node create failed: missing node")
	}
	return convertWikiNode(resp.Data.Node), nil
}

type MoveWikiNodeRequest struct {
	SpaceID               string
	NodeToken             string
	TargetParentNodeToken string
	TargetSpaceID         string
}

func (c *Client) MoveWikiNodeV2(ctx context.Context, token string, req MoveWikiNodeRequest) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiNode{}, errors.New("tenant access token is required")
	}
	return c.moveWikiNodeV2(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) MoveWikiNodeV2WithUserToken(ctx context.Context, userAccessToken string, req MoveWikiNodeRequest) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return WikiNode{}, errors.New("user access token is required")
	}
	return c.moveWikiNodeV2(ctx, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) moveWikiNodeV2(ctx context.Context, req MoveWikiNodeRequest, option larkcore.RequestOptionFunc) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	spaceID := strings.TrimSpace(req.SpaceID)
	if spaceID == "" {
		return WikiNode{}, errors.New("space id is required")
	}
	nodeToken := strings.TrimSpace(req.NodeToken)
	if nodeToken == "" {
		return WikiNode{}, errors.New("node token is required")
	}
	targetParentToken := strings.TrimSpace(req.TargetParentNodeToken)
	targetSpaceID := strings.TrimSpace(req.TargetSpaceID)
	if targetParentToken == "" && targetSpaceID == "" {
		return WikiNode{}, errors.New("target parent node token or target space id is required")
	}

	body := &larkwiki.MoveSpaceNodeReqBody{}
	if targetParentToken != "" {
		body.TargetParentToken = &targetParentToken
	}
	if targetSpaceID != "" {
		body.TargetSpaceId = &targetSpaceID
	}

	builder := larkwiki.NewMoveSpaceNodeReqBuilder().SpaceId(spaceID).NodeToken(nodeToken).Body(body)
	resp, err := c.sdk.Wiki.V2.SpaceNode.Move(ctx, builder.Build(), option)
	if err != nil {
		return WikiNode{}, err
	}
	if resp == nil {
		return WikiNode{}, errors.New("wiki node move failed: empty response")
	}
	if !resp.Success() {
		return WikiNode{}, fmt.Errorf("wiki node move failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Node == nil {
		return WikiNode{}, errors.New("wiki node move failed: missing node")
	}
	return convertWikiNode(resp.Data.Node), nil
}

type UpdateWikiNodeTitleRequest struct {
	SpaceID   string
	NodeToken string
	Title     string
}

func (c *Client) UpdateWikiNodeTitleV2(ctx context.Context, token string, req UpdateWikiNodeTitleRequest) error {
	if !c.available() {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	return c.updateWikiNodeTitleV2(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) UpdateWikiNodeTitleV2WithUserToken(ctx context.Context, userAccessToken string, req UpdateWikiNodeTitleRequest) error {
	if !c.available() {
		return ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return errors.New("user access token is required")
	}
	return c.updateWikiNodeTitleV2(ctx, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) updateWikiNodeTitleV2(ctx context.Context, req UpdateWikiNodeTitleRequest, option larkcore.RequestOptionFunc) error {
	if !c.available() {
		return ErrUnavailable
	}
	spaceID := strings.TrimSpace(req.SpaceID)
	if spaceID == "" {
		return errors.New("space id is required")
	}
	nodeToken := strings.TrimSpace(req.NodeToken)
	if nodeToken == "" {
		return errors.New("node token is required")
	}
	title := strings.TrimSpace(req.Title)
	if title == "" {
		return errors.New("title is required")
	}

	body := larkwiki.NewUpdateTitleSpaceNodeReqBodyBuilder().Title(title).Build()
	builder := larkwiki.NewUpdateTitleSpaceNodeReqBuilder().SpaceId(spaceID).NodeToken(nodeToken).Body(body)
	resp, err := c.sdk.Wiki.V2.SpaceNode.UpdateTitle(ctx, builder.Build(), option)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("wiki node update title failed: empty response")
	}
	if !resp.Success() {
		return fmt.Errorf("wiki node update title failed: %s", resp.Msg)
	}
	return nil
}

type MoveDocsToWikiRequest struct {
	SpaceID         string
	ParentNodeToken string
	ObjType         string
	ObjToken        string
	Apply           bool
	ApplySet        bool
}

type MoveDocsToWikiResult struct {
	WikiToken string `json:"wiki_token,omitempty"`
	TaskID    string `json:"task_id,omitempty"`
	Applied   *bool  `json:"applied,omitempty"`
}

func (c *Client) MoveDocsToWikiV2(ctx context.Context, token string, req MoveDocsToWikiRequest) (MoveDocsToWikiResult, error) {
	if !c.available() {
		return MoveDocsToWikiResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return MoveDocsToWikiResult{}, errors.New("tenant access token is required")
	}
	return c.moveDocsToWikiV2(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) MoveDocsToWikiV2WithUserToken(ctx context.Context, userAccessToken string, req MoveDocsToWikiRequest) (MoveDocsToWikiResult, error) {
	if !c.available() {
		return MoveDocsToWikiResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return MoveDocsToWikiResult{}, errors.New("user access token is required")
	}
	return c.moveDocsToWikiV2(ctx, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) moveDocsToWikiV2(ctx context.Context, req MoveDocsToWikiRequest, option larkcore.RequestOptionFunc) (MoveDocsToWikiResult, error) {
	if !c.available() {
		return MoveDocsToWikiResult{}, ErrUnavailable
	}
	spaceID := strings.TrimSpace(req.SpaceID)
	if spaceID == "" {
		return MoveDocsToWikiResult{}, errors.New("space id is required")
	}
	objType := strings.TrimSpace(req.ObjType)
	if objType == "" {
		return MoveDocsToWikiResult{}, errors.New("obj type is required")
	}
	objToken := strings.TrimSpace(req.ObjToken)
	if objToken == "" {
		return MoveDocsToWikiResult{}, errors.New("obj token is required")
	}

	bodyBuilder := larkwiki.NewMoveDocsToWikiSpaceNodeReqBodyBuilder().ObjType(objType).ObjToken(objToken)
	if parent := strings.TrimSpace(req.ParentNodeToken); parent != "" {
		bodyBuilder = bodyBuilder.ParentWikiToken(parent)
	}
	if req.ApplySet {
		bodyBuilder = bodyBuilder.Apply(req.Apply)
	}
	body := bodyBuilder.Build()

	builder := larkwiki.NewMoveDocsToWikiSpaceNodeReqBuilder().SpaceId(spaceID).Body(body)
	resp, err := c.sdk.Wiki.V2.SpaceNode.MoveDocsToWiki(ctx, builder.Build(), option)
	if err != nil {
		return MoveDocsToWikiResult{}, err
	}
	if resp == nil {
		return MoveDocsToWikiResult{}, errors.New("wiki move docs to wiki failed: empty response")
	}
	if !resp.Success() {
		return MoveDocsToWikiResult{}, fmt.Errorf("wiki move docs to wiki failed: %s", resp.Msg)
	}
	out := MoveDocsToWikiResult{}
	if resp.Data == nil {
		return out, nil
	}
	if resp.Data.WikiToken != nil {
		out.WikiToken = *resp.Data.WikiToken
	}
	if resp.Data.TaskId != nil {
		out.TaskID = *resp.Data.TaskId
	}
	if resp.Data.Applied != nil {
		out.Applied = resp.Data.Applied
	}
	return out, nil
}
