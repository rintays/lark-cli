package larkapi

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type MailFolder struct {
	FolderID       string `json:"folder_id"`
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	FolderType     string `json:"folder_type,omitempty"`
}

type listMailFoldersResponse struct {
	apiResponse
	Data struct {
		Items []MailFolder `json:"items"`
	} `json:"data"`
}

func (c *Client) ListMailFolders(ctx context.Context, token, mailboxID string) ([]MailFolder, error) {
	if mailboxID == "" {
		return nil, fmt.Errorf("mailbox id is required")
	}
	path := fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/folders", url.PathEscape(mailboxID))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("list mail folders failed: %s", resp.Status)
	}
	var parsed listMailFoldersResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("list mail folders failed: %s", parsed.Msg)
	}
	return parsed.Data.Items, nil
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
	apiResponse
	Data ListMailMessagesResponse `json:"data"`
}

func (c *Client) ListMailMessages(ctx context.Context, token string, req ListMailMessagesRequest) (ListMailMessagesResponse, error) {
	if req.MailboxID == "" {
		return ListMailMessagesResponse{}, fmt.Errorf("mailbox id is required")
	}
	if req.PageSize <= 0 {
		return ListMailMessagesResponse{}, fmt.Errorf("page size must be greater than 0")
	}
	path := fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/messages", url.PathEscape(req.MailboxID))
	query := url.Values{}
	query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.FolderID != "" {
		query.Set("folder_id", req.FolderID)
	}
	if req.OnlyUnread {
		query.Set("only_unread", "true")
	}
	endpoint, err := c.endpoint(path, query)
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListMailMessagesResponse{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListMailMessagesResponse{}, fmt.Errorf("list mail messages failed: %s", resp.Status)
	}
	var parsed listMailMessagesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListMailMessagesResponse{}, err
	}
	if parsed.Code != 0 {
		return ListMailMessagesResponse{}, fmt.Errorf("list mail messages failed: %s", parsed.Msg)
	}
	return parsed.Data, nil
}

type getMailMessageResponse struct {
	apiResponse
	Data struct {
		Message MailMessage `json:"message"`
	} `json:"data"`
}

func (c *Client) GetMailMessage(ctx context.Context, token, mailboxID, messageID string) (MailMessage, error) {
	if mailboxID == "" {
		return MailMessage{}, fmt.Errorf("mailbox id is required")
	}
	if messageID == "" {
		return MailMessage{}, fmt.Errorf("message id is required")
	}
	path := fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/messages/%s", url.PathEscape(mailboxID), url.PathEscape(messageID))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return MailMessage{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return MailMessage{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return MailMessage{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return MailMessage{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return MailMessage{}, fmt.Errorf("get mail message failed: %s", resp.Status)
	}
	var parsed getMailMessageResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return MailMessage{}, err
	}
	if parsed.Code != 0 {
		return MailMessage{}, fmt.Errorf("get mail message failed: %s", parsed.Msg)
	}
	return parsed.Data.Message, nil
}

type MailAddressInput struct {
	MailAddress string `json:"mail_address"`
	Name        string `json:"name,omitempty"`
}

type MailAttachment struct {
	Body     string `json:"body"`
	Filename string `json:"filename"`
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
	apiResponse
	Data struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

func (c *Client) SendMail(ctx context.Context, token, mailboxID string, req SendMailRequest) (string, error) {
	if mailboxID == "" {
		return "", fmt.Errorf("mailbox id is required")
	}
	if len(req.To) == 0 {
		return "", fmt.Errorf("to is required")
	}
	if req.Subject == "" {
		return "", fmt.Errorf("subject is required")
	}
	if req.BodyHTML == "" && req.BodyPlainText == "" {
		return "", fmt.Errorf("body_html or body_plain_text is required")
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
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/open-apis/mail/v1/user_mailboxes/%s/messages/send", url.PathEscape(mailboxID))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("send mail failed: %s", resp.Status)
	}
	var parsed sendMailResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("send mail failed: %s", parsed.Msg)
	}
	return parsed.Data.MessageID, nil
}
