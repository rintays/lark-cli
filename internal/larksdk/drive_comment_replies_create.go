package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

type CreateDriveFileCommentReplyRequest struct {
	FileToken  string
	CommentID  string
	FileType   string
	UserIDType string
	Content    *larkdrive.ReplyContent
}

type createDriveFileCommentReplyReqBody struct {
	Content *larkdrive.ReplyContent `json:"content,omitempty"`
}

type createDriveFileCommentReplyRespData struct {
	Reply *larkdrive.FileCommentReply `json:"reply,omitempty"`
}

type createDriveFileCommentReplyResp struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createDriveFileCommentReplyRespData `json:"data,omitempty"`
}

func (resp *createDriveFileCommentReplyResp) Success() bool { return resp.Code == 0 }

// CreateDriveFileCommentReply adds a reply to an existing comment thread.
//
// Note: the upstream SDK v3.5.3 doesn't include a typed wrapper for this endpoint,
// so we call the OpenAPI directly.
func (c *Client) CreateDriveFileCommentReply(ctx context.Context, token string, tokenType AccessTokenType, req CreateDriveFileCommentReplyRequest) (*larkdrive.FileCommentReply, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if req.FileToken == "" {
		return nil, errors.New("file token is required")
	}
	if req.CommentID == "" {
		return nil, errors.New("comment id is required")
	}
	if req.FileType == "" {
		return nil, errors.New("file type is required")
	}
	if req.Content == nil {
		return nil, errors.New("content is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return nil, err
	}

	apiReq := &larkcore.ApiReq{
		PathParams:  larkcore.PathParams{},
		QueryParams: larkcore.QueryParams{},
		Body:        &createDriveFileCommentReplyReqBody{Content: req.Content},
	}
	apiReq.PathParams.Set("file_token", req.FileToken)
	apiReq.PathParams.Set("comment_id", req.CommentID)
	apiReq.QueryParams.Set("file_type", req.FileType)
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}
	apiReq.ApiPath = "/open-apis/drive/v1/files/:file_token/comments/:comment_id/replies"
	apiReq.HttpMethod = http.MethodPost
	apiReq.SupportedAccessTokenTypes = []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return nil, err
	}
	resp := &createDriveFileCommentReplyResp{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, apiError("create drive file comment reply", resp.Code, resp.Msg)
	}
	if resp.Data != nil {
		return resp.Data.Reply, nil
	}
	return nil, nil
}
