package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"lark/internal/larkapi"
)

type listMailFoldersResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMailFoldersResponseData `json:"data"`
}

type listMailFoldersResponseData struct {
	Items []larkapi.MailFolder `json:"items"`
}

func (r *listMailFoldersResponse) Success() bool {
	return r.Code == 0
}

type MailAddress struct {
	MailAddress string `json:"mail_address"`
	Name        string `json:"name,omitempty"`
}

type MailMessage struct {
	MessageID    string        `json:"message_id"`
	ThreadID     string        `json:"thread_id"`
	Subject      string        `json:"subject"`
	Snippet      string        `json:"snippet"`
	FolderID     string        `json:"folder_id"`
	InternalDate string        `json:"internal_date"`
	From         MailAddress   `json:"from"`
	To           []MailAddress `json:"to"`
	CC           []MailAddress `json:"cc"`
	BCC          []MailAddress `json:"bcc"`
}

type ListMailMessagesRequest struct {
	MailboxID  string
	FolderID   string
	PageSize   int
	PageToken  string
	OnlyUnread bool
}

type ListMailMessagesResponse struct {
	Items     []MailMessage `json:"items"`
	HasMore   bool          `json:"has_more"`
	PageToken string        `json:"page_token"`
}

type listMailMessagesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMailMessagesResponseData `json:"data"`
}

type listMailMessagesResponseData struct {
	Items     []MailMessage `json:"items"`
	HasMore   bool          `json:"has_more"`
	PageToken string        `json:"page_token"`
}

func (r *listMailMessagesResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) ListMailFolders(ctx context.Context, token, mailboxID string) ([]larkapi.MailFolder, error) {
	if !c.available() || c.coreConfig == nil {
		return nil, ErrUnavailable
	}
	if mailboxID == "" {
		return nil, errors.New("mailbox id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/mail/v1/user_mailboxes/:mailbox_id/folders",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("mailbox_id", mailboxID)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if apiResp == nil {
		return nil, errors.New("list mail folders failed: empty response")
	}
	resp := &listMailFoldersResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, fmt.Errorf("list mail folders failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Items == nil {
		return nil, nil
	}
	return resp.Data.Items, nil
}

func (c *Client) ListMailMessages(ctx context.Context, token string, req ListMailMessagesRequest) (ListMailMessagesResponse, error) {
	if !c.available() || c.coreConfig == nil {
		return ListMailMessagesResponse{}, ErrUnavailable
	}
	if req.MailboxID == "" {
		return ListMailMessagesResponse{}, errors.New("mailbox id is required")
	}
	if req.PageSize <= 0 {
		return ListMailMessagesResponse{}, errors.New("page size must be greater than 0")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListMailMessagesResponse{}, errors.New("tenant access token is required")
	}

	reqParams := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	reqParams.PathParams.Set("mailbox_id", req.MailboxID)
	reqParams.QueryParams.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	if req.PageToken != "" {
		reqParams.QueryParams.Set("page_token", req.PageToken)
	}
	if req.FolderID != "" {
		reqParams.QueryParams.Set("folder_id", req.FolderID)
	}
	if req.OnlyUnread {
		reqParams.QueryParams.Set("only_unread", "true")
	}

	apiResp, err := larkcore.Request(ctx, reqParams, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	if apiResp == nil {
		return ListMailMessagesResponse{}, errors.New("list mail messages failed: empty response")
	}
	resp := &listMailMessagesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListMailMessagesResponse{}, err
	}
	if !resp.Success() {
		return ListMailMessagesResponse{}, fmt.Errorf("list mail messages failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListMailMessagesResponse{}, nil
	}
	return ListMailMessagesResponse{
		Items:     resp.Data.Items,
		HasMore:   resp.Data.HasMore,
		PageToken: resp.Data.PageToken,
	}, nil
}
