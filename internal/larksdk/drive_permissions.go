package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type DrivePermissionMember struct {
	MemberType string `json:"member_type,omitempty"`
	MemberID   string `json:"member_id,omitempty"`
	Perm       string `json:"perm,omitempty"`
	PermType   string `json:"perm_type,omitempty"`
	Type       string `json:"type,omitempty"`
	Name       string `json:"name,omitempty"`
	Avatar     string `json:"avatar,omitempty"`
	// ExternalLabel indicates whether the member is marked as external.
	ExternalLabel *bool `json:"external_label,omitempty"`
}

type AddDrivePermissionMemberRequest struct {
	MemberType          string
	MemberID            string
	Perm                string
	PermType            string
	Type                string
	NeedNotification    bool
	NeedNotificationSet bool
}

type addDrivePermissionMemberResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *addDrivePermissionMemberResponseData `json:"data"`
}

type addDrivePermissionMemberResponseData struct {
	Member DrivePermissionMember `json:"member"`
}

func (r *addDrivePermissionMemberResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) AddDrivePermissionMember(ctx context.Context, token, fileToken, fileType string, req AddDrivePermissionMemberRequest) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DrivePermissionMember{}, errors.New("tenant access token is required")
	}
	return c.addDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) AddDrivePermissionMemberWithUserToken(ctx context.Context, userAccessToken, fileToken, fileType string, req AddDrivePermissionMemberRequest) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return DrivePermissionMember{}, errors.New("user access token is required")
	}
	return c.addDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) addDrivePermissionMember(ctx context.Context, fileToken, fileType string, req AddDrivePermissionMemberRequest, option larkcore.RequestOptionFunc) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	if fileToken == "" {
		return DrivePermissionMember{}, errors.New("file token is required")
	}
	if fileType == "" {
		return DrivePermissionMember{}, errors.New("file type is required")
	}
	if req.MemberType == "" {
		return DrivePermissionMember{}, errors.New("member type is required")
	}
	if req.MemberID == "" {
		return DrivePermissionMember{}, errors.New("member id is required")
	}
	if req.Perm == "" {
		return DrivePermissionMember{}, errors.New("perm is required")
	}

	payload := map[string]any{
		"member_type": req.MemberType,
		"member_id":   req.MemberID,
		"perm":        req.Perm,
	}
	if req.PermType != "" {
		payload["perm_type"] = req.PermType
	}
	if req.Type != "" {
		payload["type"] = req.Type
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/permissions/:token/members",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("token", fileToken)
	apiReq.QueryParams.Set("type", fileType)
	if req.NeedNotificationSet {
		apiReq.QueryParams.Set("need_notification", fmt.Sprintf("%t", req.NeedNotification))
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return DrivePermissionMember{}, err
	}
	if apiResp == nil {
		return DrivePermissionMember{}, errors.New("add drive permission member failed: empty response")
	}
	resp := &addDrivePermissionMemberResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DrivePermissionMember{}, err
	}
	if !resp.Success() {
		return DrivePermissionMember{}, fmt.Errorf("add drive permission member failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return DrivePermissionMember{}, nil
	}
	return resp.Data.Member, nil
}

type ListDrivePermissionMembersRequest struct {
	Fields   string
	PermType string
}

type ListDrivePermissionMembersResult struct {
	Items []DrivePermissionMember `json:"items"`
}

type listDrivePermissionMembersResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listDrivePermissionMembersResponseData `json:"data"`
}

type listDrivePermissionMembersResponseData struct {
	Items []DrivePermissionMember `json:"items"`
}

func (r *listDrivePermissionMembersResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) ListDrivePermissionMembers(ctx context.Context, token, fileToken, fileType string, req ListDrivePermissionMembersRequest) (ListDrivePermissionMembersResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListDrivePermissionMembersResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListDrivePermissionMembersResult{}, errors.New("tenant access token is required")
	}
	return c.listDrivePermissionMembers(ctx, fileToken, fileType, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) ListDrivePermissionMembersWithUserToken(ctx context.Context, userAccessToken, fileToken, fileType string, req ListDrivePermissionMembersRequest) (ListDrivePermissionMembersResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListDrivePermissionMembersResult{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return ListDrivePermissionMembersResult{}, errors.New("user access token is required")
	}
	return c.listDrivePermissionMembers(ctx, fileToken, fileType, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) listDrivePermissionMembers(ctx context.Context, fileToken, fileType string, req ListDrivePermissionMembersRequest, option larkcore.RequestOptionFunc) (ListDrivePermissionMembersResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListDrivePermissionMembersResult{}, ErrUnavailable
	}
	if fileToken == "" {
		return ListDrivePermissionMembersResult{}, errors.New("file token is required")
	}
	if fileType == "" {
		return ListDrivePermissionMembersResult{}, errors.New("file type is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/permissions/:token/members",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("token", fileToken)
	apiReq.QueryParams.Set("type", fileType)
	if req.Fields != "" {
		apiReq.QueryParams.Set("fields", req.Fields)
	}
	if req.PermType != "" {
		apiReq.QueryParams.Set("perm_type", req.PermType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return ListDrivePermissionMembersResult{}, err
	}
	if apiResp == nil {
		return ListDrivePermissionMembersResult{}, errors.New("list drive permission members failed: empty response")
	}
	resp := &listDrivePermissionMembersResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListDrivePermissionMembersResult{}, err
	}
	if !resp.Success() {
		return ListDrivePermissionMembersResult{}, fmt.Errorf("list drive permission members failed: %s", resp.Msg)
	}
	result := ListDrivePermissionMembersResult{}
	if resp.Data != nil && resp.Data.Items != nil {
		result.Items = resp.Data.Items
	}
	return result, nil
}

type UpdateDrivePermissionMemberRequest struct {
	MemberType          string
	MemberID            string
	Perm                string
	PermType            string
	Type                string
	NeedNotification    bool
	NeedNotificationSet bool
}

type updateDrivePermissionMemberResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateDrivePermissionMemberResponseData `json:"data"`
}

type updateDrivePermissionMemberResponseData struct {
	Member DrivePermissionMember `json:"member"`
}

func (r *updateDrivePermissionMemberResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) UpdateDrivePermissionMember(ctx context.Context, token, fileToken, fileType string, req UpdateDrivePermissionMemberRequest) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DrivePermissionMember{}, errors.New("tenant access token is required")
	}
	return c.updateDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) UpdateDrivePermissionMemberWithUserToken(ctx context.Context, userAccessToken, fileToken, fileType string, req UpdateDrivePermissionMemberRequest) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return DrivePermissionMember{}, errors.New("user access token is required")
	}
	return c.updateDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) updateDrivePermissionMember(ctx context.Context, fileToken, fileType string, req UpdateDrivePermissionMemberRequest, option larkcore.RequestOptionFunc) (DrivePermissionMember, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionMember{}, ErrUnavailable
	}
	if fileToken == "" {
		return DrivePermissionMember{}, errors.New("file token is required")
	}
	if fileType == "" {
		return DrivePermissionMember{}, errors.New("file type is required")
	}
	if req.MemberType == "" {
		return DrivePermissionMember{}, errors.New("member type is required")
	}
	if req.MemberID == "" {
		return DrivePermissionMember{}, errors.New("member id is required")
	}
	if req.Perm == "" && req.PermType == "" && req.Type == "" {
		return DrivePermissionMember{}, errors.New("permission update requires at least one field")
	}

	payload := map[string]any{
		"member_type": req.MemberType,
		"member_id":   req.MemberID,
	}
	if req.Perm != "" {
		payload["perm"] = req.Perm
	}
	if req.PermType != "" {
		payload["perm_type"] = req.PermType
	}
	if req.Type != "" {
		payload["type"] = req.Type
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/permissions/:token/members/:member_id",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("token", fileToken)
	apiReq.PathParams.Set("member_id", req.MemberID)
	apiReq.QueryParams.Set("type", fileType)
	if req.NeedNotificationSet {
		apiReq.QueryParams.Set("need_notification", fmt.Sprintf("%t", req.NeedNotification))
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return DrivePermissionMember{}, err
	}
	if apiResp == nil {
		return DrivePermissionMember{}, errors.New("update drive permission member failed: empty response")
	}
	resp := &updateDrivePermissionMemberResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DrivePermissionMember{}, err
	}
	if !resp.Success() {
		return DrivePermissionMember{}, fmt.Errorf("update drive permission member failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return DrivePermissionMember{}, nil
	}
	return resp.Data.Member, nil
}

type DeleteDrivePermissionMemberRequest struct {
	MemberType string
	MemberID   string
	PermType   string
	Type       string
}

func (c *Client) DeleteDrivePermissionMember(ctx context.Context, token, fileToken, fileType string, req DeleteDrivePermissionMemberRequest) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	return c.deleteDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) DeleteDrivePermissionMemberWithUserToken(ctx context.Context, userAccessToken, fileToken, fileType string, req DeleteDrivePermissionMemberRequest) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return errors.New("user access token is required")
	}
	return c.deleteDrivePermissionMember(ctx, fileToken, fileType, req, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) deleteDrivePermissionMember(ctx context.Context, fileToken, fileType string, req DeleteDrivePermissionMemberRequest, option larkcore.RequestOptionFunc) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	if fileToken == "" {
		return errors.New("file token is required")
	}
	if fileType == "" {
		return errors.New("file type is required")
	}
	if req.MemberType == "" {
		return errors.New("member type is required")
	}
	if req.MemberID == "" {
		return errors.New("member id is required")
	}

	var payload map[string]any
	if req.PermType != "" || req.Type != "" {
		payload = map[string]any{}
		if req.PermType != "" {
			payload["perm_type"] = req.PermType
		}
		if req.Type != "" {
			payload["type"] = req.Type
		}
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/permissions/:token/members/:member_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("token", fileToken)
	apiReq.PathParams.Set("member_id", req.MemberID)
	apiReq.QueryParams.Set("type", fileType)
	apiReq.QueryParams.Set("member_type", req.MemberType)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("delete drive permission member failed: empty response")
	}
	resp := &larkcore.CodeError{}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return err
	}
	if resp.Code != 0 {
		return fmt.Errorf("delete drive permission member failed: %s", resp.Msg)
	}
	return nil
}
