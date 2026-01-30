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

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AppID      string
	AppSecret  string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) endpoint(path string, query url.Values) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	base.Path = path
	if len(query) > 0 {
		base.RawQuery = query.Encode()
	}
	return base.String(), nil
}

type apiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type tenantTokenResponse struct {
	apiResponse
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int64  `json:"expire"`
}

func (c *Client) TenantAccessToken(ctx context.Context) (string, int64, error) {
	payload := map[string]string{
		"app_id":     c.AppID,
		"app_secret": c.AppSecret,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}
	endpoint, err := c.endpoint("/open-apis/auth/v3/tenant_access_token/internal/", nil)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", 0, fmt.Errorf("token request failed: %s", resp.Status)
	}
	var parsed tenantTokenResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", 0, err
	}
	if parsed.Code != 0 {
		return "", 0, fmt.Errorf("token request failed: %s", parsed.Msg)
	}
	if parsed.TenantAccessToken == "" {
		return "", 0, fmt.Errorf("token response missing tenant_access_token")
	}
	return parsed.TenantAccessToken, parsed.Expire, nil
}

type TenantInfo struct {
	TenantKey string `json:"tenant_key"`
	Name      string `json:"name"`
}

type whoamiResponse struct {
	apiResponse
	Data struct {
		Tenant TenantInfo `json:"tenant"`
	} `json:"data"`
}

func (c *Client) WhoAmI(ctx context.Context, token string) (TenantInfo, error) {
	endpoint, err := c.endpoint("/open-apis/tenant/v2/tenant/query", nil)
	if err != nil {
		return TenantInfo{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TenantInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return TenantInfo{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return TenantInfo{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return TenantInfo{}, fmt.Errorf("whoami request failed: %s", resp.Status)
	}
	var parsed whoamiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return TenantInfo{}, err
	}
	if parsed.Code != 0 {
		return TenantInfo{}, fmt.Errorf("whoami request failed: %s", parsed.Msg)
	}
	return parsed.Data.Tenant, nil
}

type MessageRequest struct {
	ReceiveID     string
	ReceiveIDType string
	Text          string
}

type sendMessageResponse struct {
	apiResponse
	Data struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

func (c *Client) SendMessage(ctx context.Context, token string, req MessageRequest) (string, error) {
	content, err := json.Marshal(map[string]string{"text": req.Text})
	if err != nil {
		return "", err
	}
	payload := map[string]string{
		"receive_id": req.ReceiveID,
		"msg_type":   "text",
		"content":    string(content),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	receiveIDType := req.ReceiveIDType
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}
	query := url.Values{"receive_id_type": []string{receiveIDType}}
	endpoint, err := c.endpoint("/open-apis/im/v1/messages", query)
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
		return "", fmt.Errorf("send message failed: %s", resp.Status)
	}
	var parsed sendMessageResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("send message failed: %s", parsed.Msg)
	}
	return parsed.Data.MessageID, nil
}

type Chat struct {
	ChatID      string `json:"chat_id"`
	Avatar      string `json:"avatar"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	OwnerIDType string `json:"owner_id_type"`
	External    bool   `json:"external"`
	TenantKey   string `json:"tenant_key"`
}

type ListChatsRequest struct {
	PageSize   int
	PageToken  string
	UserIDType string
}

type listChatsResponse struct {
	apiResponse
	Data struct {
		Items     []Chat `json:"items"`
		PageToken string `json:"page_token"`
		HasMore   bool   `json:"has_more"`
	} `json:"data"`
}

type ListChatsResult struct {
	Items     []Chat
	PageToken string
	HasMore   bool
}

func (c *Client) ListChats(ctx context.Context, token string, req ListChatsRequest) (ListChatsResult, error) {
	query := url.Values{}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	endpoint, err := c.endpoint("/open-apis/im/v1/chats", query)
	if err != nil {
		return ListChatsResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListChatsResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListChatsResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListChatsResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListChatsResult{}, fmt.Errorf("list chats failed: %s", resp.Status)
	}
	var parsed listChatsResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListChatsResult{}, err
	}
	if parsed.Code != 0 {
		return ListChatsResult{}, fmt.Errorf("list chats failed: %s", parsed.Msg)
	}
	return ListChatsResult{
		Items:     parsed.Data.Items,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type User struct {
	UserID string `json:"user_id"`
	OpenID string `json:"open_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

type BatchGetUserIDRequest struct {
	Emails  []string
	Mobiles []string
}

type batchGetUserIDResponse struct {
	apiResponse
	Data struct {
		UserList []User `json:"user_list"`
	} `json:"data"`
}

func (c *Client) BatchGetUserIDs(ctx context.Context, token string, req BatchGetUserIDRequest) ([]User, error) {
	payload := map[string]any{}
	if len(req.Emails) > 0 {
		payload["emails"] = req.Emails
	}
	if len(req.Mobiles) > 0 {
		payload["mobiles"] = req.Mobiles
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint, err := c.endpoint("/open-apis/contact/v3/users/batch_get_id", nil)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

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
		return nil, fmt.Errorf("batch get user ids failed: %s", resp.Status)
	}
	var parsed batchGetUserIDResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("batch get user ids failed: %s", parsed.Msg)
	}
	return parsed.Data.UserList, nil
}

type ListUsersByDepartmentRequest struct {
	DepartmentID string
	PageSize     int
	PageToken    string
	UserIDType   string
}

type listUsersByDepartmentResponse struct {
	apiResponse
	Data struct {
		Items     []User `json:"items"`
		PageToken string `json:"page_token"`
		HasMore   bool   `json:"has_more"`
	} `json:"data"`
}

type ListUsersByDepartmentResult struct {
	Items     []User
	PageToken string
	HasMore   bool
}

func (c *Client) ListUsersByDepartment(ctx context.Context, token string, req ListUsersByDepartmentRequest) (ListUsersByDepartmentResult, error) {
	query := url.Values{}
	if req.DepartmentID != "" {
		query.Set("department_id", req.DepartmentID)
	}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	endpoint, err := c.endpoint("/open-apis/contact/v3/users/find_by_department", query)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListUsersByDepartmentResult{}, fmt.Errorf("list users failed: %s", resp.Status)
	}
	var parsed listUsersByDepartmentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	if parsed.Code != 0 {
		return ListUsersByDepartmentResult{}, fmt.Errorf("list users failed: %s", parsed.Msg)
	}
	return ListUsersByDepartmentResult{
		Items:     parsed.Data.Items,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type DriveFile struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	FileType  string `json:"type"`
	URL       string `json:"url"`
	ParentID  string `json:"parent_token"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_id_type"`
}

type ListDriveFilesRequest struct {
	FolderToken string
	PageSize    int
	PageToken   string
}

type listDriveFilesResponse struct {
	apiResponse
	Data struct {
		Files     []DriveFile `json:"files"`
		PageToken string      `json:"page_token"`
		HasMore   bool        `json:"has_more"`
	} `json:"data"`
}

type ListDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

func (c *Client) ListDriveFiles(ctx context.Context, token string, req ListDriveFilesRequest) (ListDriveFilesResult, error) {
	query := url.Values{}
	if req.FolderToken != "" {
		query.Set("folder_token", req.FolderToken)
	}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files", query)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", resp.Status)
	}
	var parsed listDriveFilesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListDriveFilesResult{}, err
	}
	if parsed.Code != 0 {
		return ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", parsed.Msg)
	}
	return ListDriveFilesResult{
		Files:     parsed.Data.Files,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type SearchDriveFilesRequest struct {
	Query     string
	PageSize  int
	PageToken string
}

type searchDriveFilesResponse struct {
	apiResponse
	Data struct {
		Files     []DriveFile `json:"files"`
		PageToken string      `json:"page_token"`
		HasMore   bool        `json:"has_more"`
	} `json:"data"`
}

type SearchDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

func (c *Client) SearchDriveFiles(ctx context.Context, token string, req SearchDriveFilesRequest) (SearchDriveFilesResult, error) {
	payload := map[string]any{
		"query": req.Query,
	}
	if req.PageSize > 0 {
		payload["page_size"] = req.PageSize
	}
	if req.PageToken != "" {
		payload["page_token"] = req.PageToken
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/search", nil)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SearchDriveFilesResult{}, fmt.Errorf("search drive files failed: %s", resp.Status)
	}
	var parsed searchDriveFilesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SearchDriveFilesResult{}, err
	}
	if parsed.Code != 0 {
		return SearchDriveFilesResult{}, fmt.Errorf("search drive files failed: %s", parsed.Msg)
	}
	return SearchDriveFilesResult{
		Files:     parsed.Data.Files,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type getDriveFileResponse struct {
	apiResponse
	Data struct {
		File DriveFile `json:"file"`
	} `json:"data"`
}

func (c *Client) GetDriveFile(ctx context.Context, token, fileToken string) (DriveFile, error) {
	if fileToken == "" {
		return DriveFile{}, fmt.Errorf("file token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/"+url.PathEscape(fileToken), nil)
	if err != nil {
		return DriveFile{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DriveFile{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DriveFile{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DriveFile{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DriveFile{}, fmt.Errorf("get drive file failed: %s", resp.Status)
	}
	var parsed getDriveFileResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DriveFile{}, err
	}
	if parsed.Code != 0 {
		return DriveFile{}, fmt.Errorf("get drive file failed: %s", parsed.Msg)
	}
	return parsed.Data.File, nil
}

type DocxDocument struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	RevisionID string `json:"revision_id"`
}

type CreateDocxDocumentRequest struct {
	Title       string
	FolderToken string
}

type createDocxDocumentResponse struct {
	apiResponse
	Data struct {
		Document DocxDocument `json:"document"`
	} `json:"data"`
}

func (c *Client) CreateDocxDocument(ctx context.Context, token string, req CreateDocxDocumentRequest) (DocxDocument, error) {
	payload := map[string]any{
		"title": req.Title,
	}
	if req.FolderToken != "" {
		payload["folder_token"] = req.FolderToken
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return DocxDocument{}, err
	}
	endpoint, err := c.endpoint("/open-apis/docx/v1/documents", nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return DocxDocument{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DocxDocument{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", resp.Status)
	}
	var parsed createDocxDocumentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DocxDocument{}, err
	}
	if parsed.Code != 0 {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", parsed.Msg)
	}
	return parsed.Data.Document, nil
}

type getDocxDocumentResponse struct {
	apiResponse
	Data struct {
		Document DocxDocument `json:"document"`
	} `json:"data"`
}

func (c *Client) GetDocxDocument(ctx context.Context, token, documentID string) (DocxDocument, error) {
	if documentID == "" {
		return DocxDocument{}, fmt.Errorf("document id is required")
	}
	endpoint, err := c.endpoint("/open-apis/docx/v1/documents/"+url.PathEscape(documentID), nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DocxDocument{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", resp.Status)
	}
	var parsed getDocxDocumentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DocxDocument{}, err
	}
	if parsed.Code != 0 {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", parsed.Msg)
	}
	return parsed.Data.Document, nil
}
