package larksdk

import (
	"context"
	"errors"
	"fmt"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkwiki "github.com/larksuite/oapi-sdk-go/v3/service/wiki/v2"
)

type WikiSpace struct {
	SpaceID    string `json:"space_id"`
	Name       string `json:"name"`
	SpaceType  string `json:"space_type,omitempty"`
	Visibility string `json:"visibility,omitempty"`
}

type ListWikiSpacesRequest struct {
	PageSize  int
	PageToken string
}

type ListWikiSpacesResult struct {
	Items     []WikiSpace `json:"items"`
	HasMore   bool        `json:"has_more"`
	PageToken string      `json:"page_token"`
}

type GetWikiSpaceRequest struct {
	SpaceID string
}

func (c *Client) GetWikiSpaceV2(ctx context.Context, token string, req GetWikiSpaceRequest) (WikiSpace, error) {
	if !c.available() {
		return WikiSpace{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiSpace{}, errors.New("tenant access token is required")
	}
	if req.SpaceID == "" {
		return WikiSpace{}, errors.New("space id is required")
	}
	builder := larkwiki.NewGetSpaceReqBuilder().SpaceId(req.SpaceID)
	resp, err := c.sdk.Wiki.V2.Space.Get(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return WikiSpace{}, err
	}
	if resp == nil {
		return WikiSpace{}, errors.New("wiki space get failed: empty response")
	}
	if !resp.Success() {
		return WikiSpace{}, fmt.Errorf("wiki space get failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Space == nil {
		return WikiSpace{}, errors.New("wiki space get failed: missing space")
	}
	return convertWikiSpace(resp.Data.Space), nil
}

func (c *Client) ListWikiSpacesV2(ctx context.Context, token string, req ListWikiSpacesRequest) (ListWikiSpacesResult, error) {
	if !c.available() {
		return ListWikiSpacesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListWikiSpacesResult{}, errors.New("tenant access token is required")
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}
	builder := larkwiki.NewListSpaceReqBuilder().PageSize(req.PageSize)
	if req.PageToken != "" {
		builder = builder.PageToken(req.PageToken)
	}
	resp, err := c.sdk.Wiki.V2.Space.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListWikiSpacesResult{}, err
	}
	if resp == nil {
		return ListWikiSpacesResult{}, errors.New("wiki space list failed: empty response")
	}
	if !resp.Success() {
		return ListWikiSpacesResult{}, fmt.Errorf("wiki space list failed: %s", resp.Msg)
	}
	out := ListWikiSpacesResult{}
	if resp.Data == nil {
		return out, nil
	}
	if resp.Data.HasMore != nil {
		out.HasMore = *resp.Data.HasMore
	}
	if resp.Data.PageToken != nil {
		out.PageToken = *resp.Data.PageToken
	}
	if resp.Data.Items == nil {
		return out, nil
	}
	out.Items = make([]WikiSpace, 0, len(resp.Data.Items))
	for _, s := range resp.Data.Items {
		if s == nil {
			continue
		}
		out.Items = append(out.Items, convertWikiSpace(s))
	}
	return out, nil
}

func convertWikiSpace(s *larkwiki.Space) WikiSpace {
	ws := WikiSpace{}
	if s.SpaceId != nil {
		ws.SpaceID = *s.SpaceId
	}
	if s.Name != nil {
		ws.Name = *s.Name
	}
	if s.SpaceType != nil {
		ws.SpaceType = *s.SpaceType
	}
	if s.Visibility != nil {
		ws.Visibility = *s.Visibility
	}
	return ws
}

type WikiNode struct {
	SpaceID         string `json:"space_id,omitempty"`
	NodeToken       string `json:"node_token"`
	ObjToken        string `json:"obj_token,omitempty"`
	ObjType         string `json:"obj_type"`
	ParentNodeToken string `json:"parent_node_token,omitempty"`
	NodeType        string `json:"node_type,omitempty"`
	Title           string `json:"title,omitempty"`
	HasChild        bool   `json:"has_child,omitempty"`
}

type GetWikiNodeRequest struct {
	NodeToken string
	ObjType   string
}

func (c *Client) GetWikiNodeV2(ctx context.Context, token string, req GetWikiNodeRequest) (WikiNode, error) {
	if !c.available() {
		return WikiNode{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiNode{}, errors.New("tenant access token is required")
	}
	if req.NodeToken == "" {
		return WikiNode{}, errors.New("node token is required")
	}
	if req.ObjType == "" {
		return WikiNode{}, errors.New("obj type is required")
	}

	builder := larkwiki.NewGetNodeSpaceReqBuilder().Token(req.NodeToken).ObjType(req.ObjType)
	resp, err := c.sdk.Wiki.V2.Space.GetNode(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return WikiNode{}, err
	}
	if resp == nil {
		return WikiNode{}, errors.New("wiki node get failed: empty response")
	}
	if !resp.Success() {
		return WikiNode{}, fmt.Errorf("wiki node get failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Node == nil {
		return WikiNode{}, errors.New("wiki node get failed: missing node")
	}
	return convertWikiNode(resp.Data.Node), nil
}

type ListWikiNodesRequest struct {
	SpaceID         string
	ParentNodeToken string
	PageSize        int
	PageToken       string
}

type ListWikiNodesResult struct {
	Items     []WikiNode `json:"items"`
	HasMore   bool       `json:"has_more"`
	PageToken string     `json:"page_token"`
}

func (c *Client) ListWikiNodesV2(ctx context.Context, token string, req ListWikiNodesRequest) (ListWikiNodesResult, error) {
	if !c.available() {
		return ListWikiNodesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListWikiNodesResult{}, errors.New("tenant access token is required")
	}
	if req.SpaceID == "" {
		return ListWikiNodesResult{}, errors.New("space id is required")
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	builder := larkwiki.NewListSpaceNodeReqBuilder().SpaceId(req.SpaceID).PageSize(req.PageSize)
	if req.PageToken != "" {
		builder = builder.PageToken(req.PageToken)
	}
	if req.ParentNodeToken != "" {
		builder = builder.ParentNodeToken(req.ParentNodeToken)
	}

	resp, err := c.sdk.Wiki.V2.SpaceNode.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListWikiNodesResult{}, err
	}
	if resp == nil {
		return ListWikiNodesResult{}, errors.New("wiki node list failed: empty response")
	}
	if !resp.Success() {
		return ListWikiNodesResult{}, fmt.Errorf("wiki node list failed: %s", resp.Msg)
	}
	out := ListWikiNodesResult{}
	if resp.Data == nil {
		return out, nil
	}
	if resp.Data.HasMore != nil {
		out.HasMore = *resp.Data.HasMore
	}
	if resp.Data.PageToken != nil {
		out.PageToken = *resp.Data.PageToken
	}
	if resp.Data.Items == nil {
		return out, nil
	}
	out.Items = make([]WikiNode, 0, len(resp.Data.Items))
	for _, n := range resp.Data.Items {
		if n == nil {
			continue
		}
		out.Items = append(out.Items, convertWikiNode(n))
	}
	return out, nil
}

func convertWikiNode(n *larkwiki.Node) WikiNode {
	out := WikiNode{}
	if n.SpaceId != nil {
		out.SpaceID = *n.SpaceId
	}
	if n.NodeToken != nil {
		out.NodeToken = *n.NodeToken
	}
	if n.ObjToken != nil {
		out.ObjToken = *n.ObjToken
	}
	if n.ObjType != nil {
		out.ObjType = *n.ObjType
	}
	if n.ParentNodeToken != nil {
		out.ParentNodeToken = *n.ParentNodeToken
	}
	if n.NodeType != nil {
		out.NodeType = *n.NodeType
	}
	if n.Title != nil {
		out.Title = *n.Title
	}
	if n.HasChild != nil {
		out.HasChild = *n.HasChild
	}
	return out
}

type WikiSpaceMember struct {
	MemberType string `json:"member_type,omitempty"`
	MemberID   string `json:"member_id,omitempty"`
	MemberRole string `json:"member_role,omitempty"`
	Type       string `json:"type,omitempty"`
}

type ListWikiSpaceMembersRequest struct {
	SpaceID   string
	PageSize  int
	PageToken string
}

type ListWikiSpaceMembersResult struct {
	Members   []WikiSpaceMember `json:"members"`
	HasMore   bool              `json:"has_more"`
	PageToken string            `json:"page_token"`
}

func (c *Client) ListWikiSpaceMembersV2(ctx context.Context, token string, req ListWikiSpaceMembersRequest) (ListWikiSpaceMembersResult, error) {
	if !c.available() {
		return ListWikiSpaceMembersResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListWikiSpaceMembersResult{}, errors.New("tenant access token is required")
	}
	if req.SpaceID == "" {
		return ListWikiSpaceMembersResult{}, errors.New("space id is required")
	}
	if req.PageSize <= 0 {
		req.PageSize = 50
	}

	builder := larkwiki.NewListSpaceMemberReqBuilder().SpaceId(req.SpaceID).PageSize(req.PageSize)
	if req.PageToken != "" {
		builder = builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Wiki.V2.SpaceMember.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListWikiSpaceMembersResult{}, err
	}
	if resp == nil {
		return ListWikiSpaceMembersResult{}, errors.New("wiki member list failed: empty response")
	}
	if !resp.Success() {
		return ListWikiSpaceMembersResult{}, fmt.Errorf("wiki member list failed: %s", resp.Msg)
	}

	out := ListWikiSpaceMembersResult{}
	if resp.Data == nil {
		return out, nil
	}
	if resp.Data.HasMore != nil {
		out.HasMore = *resp.Data.HasMore
	}
	if resp.Data.PageToken != nil {
		out.PageToken = *resp.Data.PageToken
	}
	if resp.Data.Members == nil {
		return out, nil
	}
	out.Members = make([]WikiSpaceMember, 0, len(resp.Data.Members))
	for _, m := range resp.Data.Members {
		if m == nil {
			continue
		}
		out.Members = append(out.Members, convertWikiSpaceMember(m))
	}
	return out, nil
}

type CreateWikiSpaceMemberRequest struct {
	SpaceID             string
	MemberType          string
	MemberID            string
	MemberRole          string
	NeedNotification    bool
	NeedNotificationSet bool
}

func (c *Client) CreateWikiSpaceMemberV2(ctx context.Context, token string, req CreateWikiSpaceMemberRequest) (WikiSpaceMember, error) {
	if !c.available() {
		return WikiSpaceMember{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiSpaceMember{}, errors.New("tenant access token is required")
	}
	if req.SpaceID == "" {
		return WikiSpaceMember{}, errors.New("space id is required")
	}
	if req.MemberType == "" {
		return WikiSpaceMember{}, errors.New("member type is required")
	}
	if req.MemberID == "" {
		return WikiSpaceMember{}, errors.New("member id is required")
	}
	if req.MemberRole == "" {
		return WikiSpaceMember{}, errors.New("member role is required")
	}

	member := larkwiki.NewMemberBuilder().MemberType(req.MemberType).MemberId(req.MemberID).MemberRole(req.MemberRole).Build()
	builder := larkwiki.NewCreateSpaceMemberReqBuilder().SpaceId(req.SpaceID).Member(member)
	if req.NeedNotificationSet {
		builder = builder.NeedNotification(req.NeedNotification)
	}

	resp, err := c.sdk.Wiki.V2.SpaceMember.Create(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return WikiSpaceMember{}, err
	}
	if resp == nil {
		return WikiSpaceMember{}, errors.New("wiki member create failed: empty response")
	}
	if !resp.Success() {
		return WikiSpaceMember{}, fmt.Errorf("wiki member create failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Member == nil {
		return WikiSpaceMember{}, nil
	}
	return convertWikiSpaceMember(resp.Data.Member), nil
}

type DeleteWikiSpaceMemberRequest struct {
	SpaceID    string
	MemberType string
	MemberID   string
}

func (c *Client) DeleteWikiSpaceMemberV2(ctx context.Context, token string, req DeleteWikiSpaceMemberRequest) (WikiSpaceMember, error) {
	if !c.available() {
		return WikiSpaceMember{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return WikiSpaceMember{}, errors.New("tenant access token is required")
	}
	if req.SpaceID == "" {
		return WikiSpaceMember{}, errors.New("space id is required")
	}
	if req.MemberType == "" {
		return WikiSpaceMember{}, errors.New("member type is required")
	}
	if req.MemberID == "" {
		return WikiSpaceMember{}, errors.New("member id is required")
	}

	member := larkwiki.NewMemberBuilder().MemberType(req.MemberType).MemberId(req.MemberID).Build()
	builder := larkwiki.NewDeleteSpaceMemberReqBuilder().SpaceId(req.SpaceID).MemberId(req.MemberID).Member(member)
	resp, err := c.sdk.Wiki.V2.SpaceMember.Delete(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return WikiSpaceMember{}, err
	}
	if resp == nil {
		return WikiSpaceMember{}, errors.New("wiki member delete failed: empty response")
	}
	if !resp.Success() {
		return WikiSpaceMember{}, fmt.Errorf("wiki member delete failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Member == nil {
		return WikiSpaceMember{}, nil
	}
	return convertWikiSpaceMember(resp.Data.Member), nil
}

func convertWikiSpaceMember(m *larkwiki.Member) WikiSpaceMember {
	out := WikiSpaceMember{}
	if m.MemberType != nil {
		out.MemberType = *m.MemberType
	}
	if m.MemberId != nil {
		out.MemberID = *m.MemberId
	}
	if m.MemberRole != nil {
		out.MemberRole = *m.MemberRole
	}
	if m.Type != nil {
		out.Type = *m.Type
	}
	return out
}
