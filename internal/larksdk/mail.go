package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkmail "github.com/larksuite/oapi-sdk-go/v3/service/mail/v1"
)

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func optionalStringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

type listMailFoldersResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMailFoldersResponseData `json:"data"`
}

type listMailFoldersResponseData struct {
	Items []MailFolder `json:"items"`
}

func (r *listMailFoldersResponse) Success() bool {
	return r.Code == 0
}

type Mailbox struct {
	MailboxID     string `json:"mailbox_id"`
	Name          string `json:"name,omitempty"`
	DisplayName   string `json:"display_name,omitempty"`
	MailAddress   string `json:"mail_address,omitempty"`
	PrimaryEmail  string `json:"primary_email,omitempty"`
	Email         string `json:"email,omitempty"`
	UserID        string `json:"user_id,omitempty"`
	MailboxStatus string `json:"mailbox_status,omitempty"`
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

type listMailboxesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMailboxesResponseData `json:"data"`
}

type listMailboxesResponseData struct {
	Items []Mailbox `json:"items"`
}

func (r *listMailboxesResponse) Success() bool {
	return r.Code == 0
}

type getMailboxResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getMailboxResponseData `json:"data"`
}

type getMailboxResponseData struct {
	Mailbox     *Mailbox `json:"mailbox"`
	UserMailbox *Mailbox `json:"user_mailbox"`
}

func (r *getMailboxResponse) Success() bool {
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

func (c *Client) ListMailFolders(ctx context.Context, token, mailboxID string) ([]MailFolder, error) {
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

func (c *Client) ListPublicMailboxes(ctx context.Context, token string) ([]Mailbox, error) {
	if !c.available() {
		return nil, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}

	req := larkmail.NewListPublicMailboxReqBuilder().Build()
	resp, err := c.sdk.Mail.V1.PublicMailbox.List(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("list public mailboxes failed: empty response")
	}
	if !resp.Success() {
		return nil, fmt.Errorf("list public mailboxes failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Items == nil {
		return nil, nil
	}

	out := make([]Mailbox, 0, len(resp.Data.Items))
	for _, mb := range resp.Data.Items {
		if mb == nil {
			continue
		}
		mailbox := Mailbox{}
		if mb.PublicMailboxId != nil {
			mailbox.MailboxID = *mb.PublicMailboxId
		}
		if mb.Name != nil {
			mailbox.Name = *mb.Name
		}
		// our CLI historically prints `primary_email`; SDK uses `email`.
		if mb.Email != nil {
			mailbox.PrimaryEmail = *mb.Email
		}
		out = append(out, mailbox)
	}
	return out, nil
}

func (c *Client) GetMailbox(ctx context.Context, token, mailboxID string) (Mailbox, error) {
	if !c.available() || c.coreConfig == nil {
		return Mailbox{}, ErrUnavailable
	}
	if mailboxID == "" {
		return Mailbox{}, errors.New("mailbox id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Mailbox{}, errors.New("tenant access token is required")
	}

	req := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/mail/v1/user_mailboxes/:user_mailbox_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	req.PathParams.Set("user_mailbox_id", mailboxID)

	apiResp, err := larkcore.Request(ctx, req, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return Mailbox{}, err
	}
	if apiResp == nil {
		return Mailbox{}, errors.New("get mailbox failed: empty response")
	}
	resp := &getMailboxResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return Mailbox{}, err
	}
	if !resp.Success() {
		return Mailbox{}, fmt.Errorf("get mailbox failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return Mailbox{}, nil
	}
	mailbox := resp.Data.Mailbox
	if mailbox == nil {
		mailbox = resp.Data.UserMailbox
	}
	if mailbox == nil {
		return Mailbox{}, nil
	}
	return *mailbox, nil
}

func (c *Client) ListMailMessages(ctx context.Context, token string, req ListMailMessagesRequest) (ListMailMessagesResponse, error) {
	if !c.available() {
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

	reqBuilder := larkmail.NewListUserMailboxMessageReqBuilder().
		UserMailboxId(req.MailboxID).
		PageSize(req.PageSize)
	if req.PageToken != "" {
		reqBuilder.PageToken(req.PageToken)
	}
	if req.FolderID != "" {
		reqBuilder.FolderId(req.FolderID)
	}
	if req.OnlyUnread {
		reqBuilder.OnlyUnread(true)
	}
	listReq := reqBuilder.Build()
	resp, err := c.sdk.Mail.V1.UserMailboxMessage.List(ctx, listReq, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	if resp == nil {
		return ListMailMessagesResponse{}, errors.New("list mail messages failed: empty response")
	}
	if !resp.Success() {
		return ListMailMessagesResponse{}, fmt.Errorf("list mail messages failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Items == nil {
		return ListMailMessagesResponse{}, nil
	}

	items := make([]MailMessage, 0, len(resp.Data.Items))
	for _, messageID := range resp.Data.Items {
		if messageID == "" {
			continue
		}
		items = append(items, MailMessage{MessageID: messageID})
	}

	hasMore := false
	if resp.Data.HasMore != nil {
		hasMore = *resp.Data.HasMore
	}
	pageToken := ""
	if resp.Data.PageToken != nil {
		pageToken = *resp.Data.PageToken
	}
	return ListMailMessagesResponse{
		Items:     items,
		HasMore:   hasMore,
		PageToken: pageToken,
	}, nil
}

func (c *Client) GetMailMessage(ctx context.Context, token, mailboxID, messageID string) (MailMessage, error) {
	if !c.available() {
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

	req := larkmail.NewGetUserMailboxMessageReqBuilder().
		UserMailboxId(mailboxID).
		MessageId(messageID).
		Build()
	resp, err := c.sdk.Mail.V1.UserMailboxMessage.Get(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return MailMessage{}, err
	}
	if resp == nil {
		return MailMessage{}, errors.New("get mail message failed: empty response")
	}
	if !resp.Success() {
		return MailMessage{}, fmt.Errorf("get mail message failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Message == nil {
		return MailMessage{}, nil
	}

	msg := resp.Data.Message
	out := MailMessage{}
	if msg.ThreadId != nil {
		out.ThreadID = *msg.ThreadId
	}
	if msg.Subject != nil {
		out.Subject = *msg.Subject
	}
	// MessageId is not part of `Message`; keep caller-provided messageID for stable output.
	out.MessageID = messageID

	for _, to := range msg.To {
		if to == nil {
			continue
		}
		out.To = append(out.To, MailAddress{MailAddress: derefString(to.MailAddress), Name: derefString(to.Name)})
	}
	for _, cc := range msg.Cc {
		if cc == nil {
			continue
		}
		out.CC = append(out.CC, MailAddress{MailAddress: derefString(cc.MailAddress), Name: derefString(cc.Name)})
	}
	for _, bcc := range msg.Bcc {
		if bcc == nil {
			continue
		}
		out.BCC = append(out.BCC, MailAddress{MailAddress: derefString(bcc.MailAddress), Name: derefString(bcc.Name)})
	}

	return out, nil
}

func (c *Client) SendMail(ctx context.Context, token, mailboxID string, req SendMailRequest) (string, error) {
	if !c.available() {
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

	to := make([]*larkmail.MailAddress, 0, len(req.To))
	for _, addr := range req.To {
		addr := addr
		to = append(to, &larkmail.MailAddress{MailAddress: &addr.MailAddress, Name: optionalStringPtr(addr.Name)})
	}
	cc := make([]*larkmail.MailAddress, 0, len(req.CC))
	for _, addr := range req.CC {
		addr := addr
		cc = append(cc, &larkmail.MailAddress{MailAddress: &addr.MailAddress, Name: optionalStringPtr(addr.Name)})
	}
	bcc := make([]*larkmail.MailAddress, 0, len(req.BCC))
	for _, addr := range req.BCC {
		addr := addr
		bcc = append(bcc, &larkmail.MailAddress{MailAddress: &addr.MailAddress, Name: optionalStringPtr(addr.Name)})
	}

	attachments := make([]*larkmail.Attachment, 0, len(req.Attachments))
	for _, att := range req.Attachments {
		att := att
		attachments = append(attachments, &larkmail.Attachment{Body: &att.Body, Filename: &att.Filename})
	}

	body := larkmail.NewSendUserMailboxMessageReqBodyBuilder().
		Subject(req.Subject).
		To(to).
		Build()
	if len(cc) > 0 {
		body.Cc = cc
	}
	if len(bcc) > 0 {
		body.Bcc = bcc
	}
	if req.HeadFromName != "" {
		name := req.HeadFromName
		body.HeadFrom = &larkmail.MailAddress{MailAddress: nil, Name: &name}
	}
	if req.BodyHTML != "" {
		val := req.BodyHTML
		body.BodyHtml = &val
	}
	if req.BodyPlainText != "" {
		val := req.BodyPlainText
		body.BodyPlainText = &val
	}
	if len(attachments) > 0 {
		body.Attachments = attachments
	}

	sendReq := larkmail.NewSendUserMailboxMessageReqBuilder().
		UserMailboxId(mailboxID).
		Body(body).
		Build()
	resp, err := c.sdk.Mail.V1.UserMailboxMessage.Send(ctx, sendReq, larkcore.WithUserAccessToken(token))
	if err != nil {
		return "", err
	}
	if resp == nil {
		return "", errors.New("send mail failed: empty response")
	}
	if !resp.Success() {
		return "", fmt.Errorf("send mail failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.MessageId == nil {
		return "", nil
	}
	return *resp.Data.MessageId, nil
}
