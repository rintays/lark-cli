package larksdk

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

func (c *Client) ListDriveFiles(ctx context.Context, token string, req ListDriveFilesRequest) (ListDriveFilesResult, error) {
	if !c.available() {
		return ListDriveFilesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListDriveFilesResult{}, errors.New("tenant access token is required")
	}

	builder := larkdrive.NewListFileReqBuilder()
	if req.FolderToken != "" {
		builder.FolderToken(req.FolderToken)
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Drive.V1.File.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	if resp == nil {
		return ListDriveFilesResult{}, errors.New("list drive files failed: empty response")
	}
	if !resp.Success() {
		return ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", resp.Msg)
	}

	result := ListDriveFilesResult{}
	if resp.Data != nil {
		if resp.Data.Files != nil {
			result.Files = make([]DriveFile, 0, len(resp.Data.Files))
			for _, file := range resp.Data.Files {
				result.Files = append(result.Files, mapDriveFile(file))
			}
		}
		if resp.Data.NextPageToken != nil {
			result.PageToken = *resp.Data.NextPageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

type searchDriveFilesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *searchDriveFilesResponseData `json:"data"`
}

type searchDriveFilesResponseData struct {
	Files     []*larkdrive.File `json:"files"`
	PageToken *string           `json:"page_token"`
	HasMore   *bool             `json:"has_more"`
}

func (r *searchDriveFilesResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) SearchDriveFiles(ctx context.Context, token string, req SearchDriveFilesRequest) (SearchDriveFilesResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SearchDriveFilesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SearchDriveFilesResult{}, errors.New("tenant access token is required")
	}

	payload := map[string]any{
		"query": req.Query,
	}
	if req.PageSize > 0 {
		payload["page_size"] = req.PageSize
	}
	if req.PageToken != "" {
		payload["page_token"] = req.PageToken
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/files/search",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
		Body:                      payload,
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	if apiResp == nil {
		return SearchDriveFilesResult{}, errors.New("search drive files failed: empty response")
	}
	resp := &searchDriveFilesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SearchDriveFilesResult{}, err
	}
	if !resp.Success() {
		return SearchDriveFilesResult{}, fmt.Errorf("search drive files failed: %s", resp.Msg)
	}
	result := SearchDriveFilesResult{}
	if resp.Data != nil {
		if resp.Data.Files != nil {
			result.Files = make([]DriveFile, 0, len(resp.Data.Files))
			for _, file := range resp.Data.Files {
				result.Files = append(result.Files, mapDriveFile(file))
			}
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

type getDriveFileResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getDriveFileResponseData `json:"data"`
}

type getDriveFileResponseData struct {
	File *larkdrive.File `json:"file"`
}

func (r *getDriveFileResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) GetDriveFileMetadata(ctx context.Context, token string, req GetDriveFileRequest) (DriveFile, error) {
	if !c.available() || c.coreConfig == nil {
		return DriveFile{}, ErrUnavailable
	}
	if req.FileToken == "" {
		return DriveFile{}, errors.New("file token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DriveFile{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/files/:file_token",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("file_token", req.FileToken)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return DriveFile{}, err
	}
	if apiResp == nil {
		return DriveFile{}, errors.New("get drive file failed: empty response")
	}
	resp := &getDriveFileResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DriveFile{}, err
	}
	if !resp.Success() {
		return DriveFile{}, fmt.Errorf("get drive file failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.File == nil {
		return DriveFile{}, nil
	}
	return mapDriveFile(resp.Data.File), nil
}

type DrivePermissionPublic struct {
	ExternalAccess  bool   `json:"external_access"`
	SecurityEntity  string `json:"security_entity"`
	CommentEntity   string `json:"comment_entity"`
	ShareEntity     string `json:"share_entity"`
	LinkShareEntity string `json:"link_share_entity"`
	InviteExternal  bool   `json:"invite_external"`
	LockSwitch      bool   `json:"lock_switch"`
}

type UpdateDrivePermissionPublicRequest struct {
	ExternalAccess  *bool  `json:"external_access,omitempty"`
	SecurityEntity  string `json:"security_entity,omitempty"`
	CommentEntity   string `json:"comment_entity,omitempty"`
	ShareEntity     string `json:"share_entity,omitempty"`
	LinkShareEntity string `json:"link_share_entity,omitempty"`
	InviteExternal  *bool  `json:"invite_external,omitempty"`
}

type updateDrivePermissionPublicResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateDrivePermissionPublicResponseData `json:"data"`
}

type updateDrivePermissionPublicResponseData struct {
	Permission DrivePermissionPublic `json:"permission_public"`
}

func (r *updateDrivePermissionPublicResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) UpdateDrivePermissionPublic(ctx context.Context, token, fileToken, fileType string, req UpdateDrivePermissionPublicRequest) (DrivePermissionPublic, error) {
	if !c.available() || c.coreConfig == nil {
		return DrivePermissionPublic{}, ErrUnavailable
	}
	if fileToken == "" {
		return DrivePermissionPublic{}, fmt.Errorf("file token is required")
	}
	if fileType == "" {
		return DrivePermissionPublic{}, fmt.Errorf("file type is required")
	}
	if !hasDrivePermissionPublicUpdate(req) {
		return DrivePermissionPublic{}, fmt.Errorf("permission update requires at least one field")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DrivePermissionPublic{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/permissions/:file_token/public",
		HttpMethod:                http.MethodPatch,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      req,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("file_token", fileToken)
	apiReq.QueryParams.Set("type", fileType)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	if apiResp == nil {
		return DrivePermissionPublic{}, errors.New("update drive permission failed: empty response")
	}
	resp := &updateDrivePermissionPublicResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DrivePermissionPublic{}, err
	}
	if !resp.Success() {
		return DrivePermissionPublic{}, fmt.Errorf("update drive permission failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return DrivePermissionPublic{}, nil
	}
	return resp.Data.Permission, nil
}

func hasDrivePermissionPublicUpdate(req UpdateDrivePermissionPublicRequest) bool {
	if req.ExternalAccess != nil || req.InviteExternal != nil {
		return true
	}
	if req.SecurityEntity != "" || req.CommentEntity != "" || req.ShareEntity != "" || req.LinkShareEntity != "" {
		return true
	}
	return false
}

func (c *Client) UploadDriveFile(ctx context.Context, token string, req UploadDriveFileRequest) (DriveUploadResult, error) {
	if !c.available() {
		return DriveUploadResult{}, ErrUnavailable
	}
	if req.File == nil {
		return DriveUploadResult{}, fmt.Errorf("file is required")
	}
	if req.FileName == "" {
		return DriveUploadResult{}, fmt.Errorf("file name is required")
	}
	if req.Size < 0 {
		return DriveUploadResult{}, fmt.Errorf("file size must be non-negative")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DriveUploadResult{}, errors.New("tenant access token is required")
	}

	parentNode := req.FolderToken
	if parentNode == "" {
		parentNode = "root"
	}

	body := larkdrive.NewUploadAllFileReqBodyBuilder().
		FileName(req.FileName).
		ParentType("explorer").
		ParentNode(parentNode).
		Size(int(req.Size)).
		File(req.File).
		Build()
	builder := larkdrive.NewUploadAllFileReqBuilder().Body(body)

	resp, err := c.sdk.Drive.V1.File.UploadAll(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return DriveUploadResult{}, err
	}
	if resp == nil {
		return DriveUploadResult{}, errors.New("drive upload failed: empty response")
	}
	if !resp.Success() {
		return DriveUploadResult{}, fmt.Errorf("drive upload failed: %s", resp.Msg)
	}

	result := DriveUploadResult{}
	if resp.Data != nil && resp.Data.FileToken != nil {
		result.FileToken = *resp.Data.FileToken
	}
	if result.FileToken == "" {
		return DriveUploadResult{}, fmt.Errorf("drive upload response missing file token")
	}
	return result, nil
}

func (c *Client) DownloadDriveFile(ctx context.Context, token, fileToken string) (io.ReadCloser, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if fileToken == "" {
		return nil, fmt.Errorf("file token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	builder := larkdrive.NewDownloadFileReqBuilder().FileToken(fileToken)
	resp, err := c.sdk.Drive.V1.File.Download(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("drive download failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("drive download failed: %s", resp.Msg)
	}
	if resp.File == nil {
		return nil, errors.New("drive download failed: empty file")
	}
	return io.NopCloser(resp.File), nil
}

func mapDriveFile(file *larkdrive.File) DriveFile {
	if file == nil {
		return DriveFile{}
	}
	result := DriveFile{}
	if file.Token != nil {
		result.Token = *file.Token
	}
	if file.Name != nil {
		result.Name = *file.Name
	}
	if file.Type != nil {
		result.FileType = *file.Type
	}
	if file.Url != nil {
		result.URL = *file.Url
	}
	if file.ParentToken != nil {
		result.ParentID = *file.ParentToken
	}
	if file.OwnerId != nil {
		result.OwnerID = *file.OwnerId
	}
	return result
}
