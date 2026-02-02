package larksdk

import (
	"context"
	"errors"
	"fmt"
	"math"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

func (c *Client) UploadDriveMedia(ctx context.Context, token string, req UploadDriveMediaRequest) (DriveMediaUploadResult, error) {
	if !c.available() {
		return DriveMediaUploadResult{}, ErrUnavailable
	}
	if req.File == nil {
		return DriveMediaUploadResult{}, fmt.Errorf("file is required")
	}
	if req.FileName == "" {
		return DriveMediaUploadResult{}, fmt.Errorf("file name is required")
	}
	if req.ParentType == "" {
		return DriveMediaUploadResult{}, fmt.Errorf("parent type is required")
	}
	if req.ParentNode == "" {
		return DriveMediaUploadResult{}, fmt.Errorf("parent node is required")
	}
	if req.Size < 0 {
		return DriveMediaUploadResult{}, fmt.Errorf("file size must be non-negative")
	}
	if req.Size > math.MaxInt {
		return DriveMediaUploadResult{}, fmt.Errorf("file size exceeds platform limits")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return DriveMediaUploadResult{}, errors.New("tenant access token is required")
	}

	builder := larkdrive.NewUploadAllMediaReqBodyBuilder().
		FileName(req.FileName).
		ParentType(req.ParentType).
		ParentNode(req.ParentNode).
		Size(int(req.Size)).
		File(req.File)
	if req.Checksum != "" {
		builder.Checksum(req.Checksum)
	}
	if req.Extra != "" {
		builder.Extra(req.Extra)
	}
	body := builder.Build()

	resp, err := c.sdk.Drive.V1.Media.UploadAll(ctx, larkdrive.NewUploadAllMediaReqBuilder().Body(body).Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return DriveMediaUploadResult{}, err
	}
	if resp == nil {
		return DriveMediaUploadResult{}, errors.New("drive media upload failed: empty response")
	}
	if !resp.Success() {
		return DriveMediaUploadResult{}, fmt.Errorf("drive media upload failed: %s", resp.Msg)
	}
	result := DriveMediaUploadResult{}
	if resp.Data != nil && resp.Data.FileToken != nil {
		result.FileToken = *resp.Data.FileToken
	}
	if result.FileToken == "" {
		return DriveMediaUploadResult{}, fmt.Errorf("drive media upload response missing file token")
	}
	return result, nil
}
