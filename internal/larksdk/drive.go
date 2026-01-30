package larksdk

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"

	"lark/internal/larkapi"
)

func (c *Client) ListDriveFiles(ctx context.Context, token string, req larkapi.ListDriveFilesRequest) (larkapi.ListDriveFilesResult, error) {
	if !c.available() {
		return larkapi.ListDriveFilesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.ListDriveFilesResult{}, errors.New("tenant access token is required")
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
		return larkapi.ListDriveFilesResult{}, err
	}
	if resp == nil {
		return larkapi.ListDriveFilesResult{}, errors.New("list drive files failed: empty response")
	}
	if !resp.Success() {
		return larkapi.ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", resp.Msg)
	}

	result := larkapi.ListDriveFilesResult{}
	if resp.Data != nil {
		if resp.Data.Files != nil {
			result.Files = make([]larkapi.DriveFile, 0, len(resp.Data.Files))
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

func (c *Client) DownloadDriveFile(ctx context.Context, token, fileToken string) (io.ReadCloser, error) {
	if !c.available() || c.coreConfig == nil {
		return nil, ErrUnavailable
	}
	if fileToken == "" {
		return nil, fmt.Errorf("file token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/" + url.PathEscape(fileToken) + "/download")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+tenantToken)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("drive download failed: %s", resp.Status)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("drive download failed: %s", resp.Status)
		}
		return nil, fmt.Errorf("drive download failed: %s: %s", resp.Status, string(bytes.TrimSpace(data)))
	}
	return resp.Body, nil
}

func mapDriveFile(file *larkdrive.File) larkapi.DriveFile {
	if file == nil {
		return larkapi.DriveFile{}
	}
	result := larkapi.DriveFile{}
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
