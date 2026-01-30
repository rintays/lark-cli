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
