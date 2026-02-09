package larksdk

import (
	"context"
	"errors"

	larkdrive "github.com/larksuite/oapi-sdk-go/v3/service/drive/v1"
)

type DriveFileComment = larkdrive.FileComment

type DriveFileCommentReply = larkdrive.FileCommentReply

type CreateDriveFileCommentRequest struct {
	FileToken  string
	FileType   string
	UserIDType string
	Comment    *larkdrive.FileComment
}

type ListDriveFileCommentsRequest struct {
	FileToken  string
	FileType   string
	UserIDType string
	PageSize   int
	PageToken  string
}

type ListDriveFileCommentsResult struct {
	Items     []*larkdrive.FileComment
	PageToken string
	HasMore   bool
}

type GetDriveFileCommentRequest struct {
	FileToken  string
	CommentID  string
	FileType   string
	UserIDType string
}

type PatchDriveFileCommentRequest struct {
	FileToken  string
	CommentID  string
	FileType   string
	IsSolved   bool
	UserIDType string
}

type ListDriveFileCommentRepliesRequest struct {
	FileToken  string
	CommentID  string
	FileType   string
	UserIDType string
	PageSize   int
	PageToken  string
}

type ListDriveFileCommentRepliesResult struct {
	Items     []*larkdrive.FileCommentReply
	PageToken string
	HasMore   bool
}

type UpdateDriveFileCommentReplyRequest struct {
	FileToken  string
	CommentID  string
	ReplyID    string
	FileType   string
	UserIDType string
	Content    *larkdrive.ReplyContent
}

func (c *Client) CreateDriveFileComment(ctx context.Context, token string, tokenType AccessTokenType, req CreateDriveFileCommentRequest) (*larkdrive.FileComment, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	if req.FileToken == "" {
		return nil, errors.New("file token is required")
	}
	if req.FileType == "" {
		return nil, errors.New("file type is required")
	}
	if req.Comment == nil {
		return nil, errors.New("comment is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return nil, err
	}

	builder := larkdrive.NewCreateFileCommentReqBuilder().
		FileToken(req.FileToken).
		FileType(req.FileType).
		FileComment(req.Comment)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Drive.V1.FileComment.Create(ctx, builder.Build(), option)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("create drive file comment failed: empty response")
	}
	if !resp.Success() {
		return nil, apiError("create drive file comment", resp.Code, resp.Msg)
	}
	return mapCreateFileCommentData(resp.Data), nil
}

func mapCreateFileCommentData(data *larkdrive.CreateFileCommentRespData) *larkdrive.FileComment {
	if data == nil {
		return nil
	}
	return &larkdrive.FileComment{
		CommentId:    data.CommentId,
		UserId:       data.UserId,
		CreateTime:   data.CreateTime,
		UpdateTime:   data.UpdateTime,
		IsSolved:     data.IsSolved,
		SolvedTime:   data.SolvedTime,
		SolverUserId: data.SolverUserId,
		HasMore:      data.HasMore,
		PageToken:    data.PageToken,
		IsWhole:      data.IsWhole,
		Quote:        data.Quote,
		ReplyList:    data.ReplyList,
	}
}

func (c *Client) ListDriveFileComments(ctx context.Context, token string, tokenType AccessTokenType, req ListDriveFileCommentsRequest) (ListDriveFileCommentsResult, error) {
	if !c.available() {
		return ListDriveFileCommentsResult{}, ErrUnavailable
	}
	if req.FileToken == "" {
		return ListDriveFileCommentsResult{}, errors.New("file token is required")
	}
	if req.FileType == "" {
		return ListDriveFileCommentsResult{}, errors.New("file type is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return ListDriveFileCommentsResult{}, err
	}

	builder := larkdrive.NewListFileCommentReqBuilder().
		FileToken(req.FileToken).
		FileType(req.FileType)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Drive.V1.FileComment.List(ctx, builder.Build(), option)
	if err != nil {
		return ListDriveFileCommentsResult{}, err
	}
	if resp == nil {
		return ListDriveFileCommentsResult{}, errors.New("list drive file comments failed: empty response")
	}
	if !resp.Success() {
		return ListDriveFileCommentsResult{}, apiError("list drive file comments", resp.Code, resp.Msg)
	}
	result := ListDriveFileCommentsResult{}
	if resp.Data != nil {
		result.Items = resp.Data.Items
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

func (c *Client) GetDriveFileComment(ctx context.Context, token string, tokenType AccessTokenType, req GetDriveFileCommentRequest) (*larkdrive.FileComment, error) {
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

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return nil, err
	}

	builder := larkdrive.NewGetFileCommentReqBuilder().
		FileToken(req.FileToken).
		CommentId(req.CommentID).
		FileType(req.FileType)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Drive.V1.FileComment.Get(ctx, builder.Build(), option)
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("get drive file comment failed: empty response")
	}
	if !resp.Success() {
		return nil, apiError("get drive file comment", resp.Code, resp.Msg)
	}
	return mapGetFileCommentData(resp.Data), nil
}

func mapGetFileCommentData(data *larkdrive.GetFileCommentRespData) *larkdrive.FileComment {
	if data == nil {
		return nil
	}
	return &larkdrive.FileComment{
		CommentId:    data.CommentId,
		UserId:       data.UserId,
		CreateTime:   data.CreateTime,
		UpdateTime:   data.UpdateTime,
		IsSolved:     data.IsSolved,
		SolvedTime:   data.SolvedTime,
		SolverUserId: data.SolverUserId,
		HasMore:      data.HasMore,
		PageToken:    data.PageToken,
		IsWhole:      data.IsWhole,
		Quote:        data.Quote,
		ReplyList:    data.ReplyList,
	}
}

func (c *Client) PatchDriveFileComment(ctx context.Context, token string, tokenType AccessTokenType, req PatchDriveFileCommentRequest) error {
	if !c.available() {
		return ErrUnavailable
	}
	if req.FileToken == "" {
		return errors.New("file token is required")
	}
	if req.CommentID == "" {
		return errors.New("comment id is required")
	}
	if req.FileType == "" {
		return errors.New("file type is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return err
	}

	body := &larkdrive.PatchFileCommentReqBody{IsSolved: &req.IsSolved}
	builder := larkdrive.NewPatchFileCommentReqBuilder().
		FileToken(req.FileToken).
		CommentId(req.CommentID).
		FileType(req.FileType).
		Body(body)

	resp, err := c.sdk.Drive.V1.FileComment.Patch(ctx, builder.Build(), option)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("patch drive file comment failed: empty response")
	}
	if !resp.Success() {
		return apiError("patch drive file comment", resp.Code, resp.Msg)
	}
	return nil
}

func (c *Client) ListDriveFileCommentReplies(ctx context.Context, token string, tokenType AccessTokenType, req ListDriveFileCommentRepliesRequest) (ListDriveFileCommentRepliesResult, error) {
	if !c.available() {
		return ListDriveFileCommentRepliesResult{}, ErrUnavailable
	}
	if req.FileToken == "" {
		return ListDriveFileCommentRepliesResult{}, errors.New("file token is required")
	}
	if req.CommentID == "" {
		return ListDriveFileCommentRepliesResult{}, errors.New("comment id is required")
	}
	if req.FileType == "" {
		return ListDriveFileCommentRepliesResult{}, errors.New("file type is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return ListDriveFileCommentRepliesResult{}, err
	}

	builder := larkdrive.NewListFileCommentReplyReqBuilder().
		FileToken(req.FileToken).
		CommentId(req.CommentID).
		FileType(req.FileType)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}

	resp, err := c.sdk.Drive.V1.FileCommentReply.List(ctx, builder.Build(), option)
	if err != nil {
		return ListDriveFileCommentRepliesResult{}, err
	}
	if resp == nil {
		return ListDriveFileCommentRepliesResult{}, errors.New("list drive file comment replies failed: empty response")
	}
	if !resp.Success() {
		return ListDriveFileCommentRepliesResult{}, apiError("list drive file comment replies", resp.Code, resp.Msg)
	}
	result := ListDriveFileCommentRepliesResult{}
	if resp.Data != nil {
		result.Items = resp.Data.Items
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

func (c *Client) UpdateDriveFileCommentReply(ctx context.Context, token string, tokenType AccessTokenType, req UpdateDriveFileCommentReplyRequest) error {
	if !c.available() {
		return ErrUnavailable
	}
	if req.FileToken == "" {
		return errors.New("file token is required")
	}
	if req.CommentID == "" {
		return errors.New("comment id is required")
	}
	if req.ReplyID == "" {
		return errors.New("reply id is required")
	}
	if req.FileType == "" {
		return errors.New("file type is required")
	}
	if req.Content == nil {
		return errors.New("content is required")
	}

	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return err
	}

	body := &larkdrive.UpdateFileCommentReplyReqBody{Content: req.Content}
	builder := larkdrive.NewUpdateFileCommentReplyReqBuilder().
		FileToken(req.FileToken).
		CommentId(req.CommentID).
		ReplyId(req.ReplyID).
		FileType(req.FileType).
		Body(body)
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Drive.V1.FileCommentReply.Update(ctx, builder.Build(), option)
	if err != nil {
		return err
	}
	if resp == nil {
		return errors.New("update drive file comment reply failed: empty response")
	}
	if !resp.Success() {
		return apiError("update drive file comment reply", resp.Code, resp.Msg)
	}
	return nil
}
