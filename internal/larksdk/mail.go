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

type MailAddressInput struct {
	MailAddress string `json:"mail_address"`
	Name        string `json:"name,omitempty"`
}

type MailAttachment struct {
	Body     string `json:"body"`
	Filename string `json:"filename"`
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

type getMailMessageResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getMailMessageResponseData `json:"data"`
}

type getMailMessageResponseData struct {
	Message *MailMessage `json:"message"`
}

func (r *getMailMessageResponse) Success() bool {
	return r.Code == 0
}

type SendMailRequest struct {
	Subject       string
	To            []MailAddressInput
	CC            []MailAddressInput
	BCC           []MailAddressInput
	HeadFromName  string
	BodyHTML      string
	BodyPlainText string
	Attachments   []MailAttachment
}

type sendMailResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *sendMailResponseData `json:"data"`
}

type sendMailResponseData struct {
	MessageID string `json:"message_id"`
}

func (r *sendMailResponse) Success() bool {
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

func (c *Client) GetMailMessage(ctx context.Context, token, mailboxID, messageID string) (MailMessage, error) {
	if !c.available() || c.coreConfig == nil {
		return MailMessage{}, ErrUnavailable
	}
	if mailboxID == "" {
		return MailMessage{}, errors.New("mailbox id is required")
	}
	if messageID == "" {
		return MailMessage{}, errors.New("message id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return MailMessage{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/:message_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	req.PathParams.Set("mailbox_id", mailboxID)
	req.PathParams.Set("message_id", messageID)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return MailMessage{}, err
	}
	if apiResp == nil {
		return MailMessage{}, errors.New("get mail message failed: empty response")
	}
	resp := &getMailMessageResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return MailMessage{}, err
	}
	if !resp.Success() {
		return MailMessage{}, fmt.Errorf("get mail message failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Message == nil {
		return MailMessage{}, nil
	}
	return *resp.Data.Message, nil
}

func (c *Client) SendMail(ctx context.Context, token, mailboxID string, req SendMailRequest) (string, error) {
	if !c.available() || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	if token == "" {
		return "", errors.New("user access token is required")
	}
	if mailboxID == "" {
		return "", errors.New("mailbox id is required")
	}
	if len(req.To) == 0 {
		return "", errors.New("to is required")
	}
	if req.Subject == "" {
		return "", errors.New("subject is required")
	}
	if req.BodyHTML == "" && req.BodyPlainText == "" {
		return "", errors.New("body_html or body_plain_text is required")
	}

	payload := map[string]any{
		"subject": req.Subject,
		"to":      req.To,
	}
	if len(req.CC) > 0 {
		payload["cc"] = req.CC
	}
	if len(req.BCC) > 0 {
		payload["bcc"] = req.BCC
	}
	if req.HeadFromName != "" {
		payload["head_from"] = map[string]any{"name": req.HeadFromName}
	}
	if req.BodyHTML != "" {
		payload["body_html"] = req.BodyHTML
	}
	if req.BodyPlainText != "" {
		payload["body_plain_text"] = req.BodyPlainText
	}
	if len(req.Attachments) > 0 {
		payload["attachments"] = req.Attachments
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/send",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("mailbox_id", mailboxID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithUserAccessToken(token))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("send mail failed: empty response")
	}
	resp := &sendMailResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("send mail failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return "", nil
	}
	return resp.Data.MessageID, nil
}
