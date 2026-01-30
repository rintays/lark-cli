package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	lark "github.com/larksuite/oapi-sdk-go/v3"
	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	im "github.com/larksuite/oapi-sdk-go/v3/service/im/v1"

	"lark/internal/config"
)

var ErrUnavailable = errors.New("lark sdk unavailable")

// Option configures SDK client initialization.
type Option func(*options)

type options struct {
	httpClient        *http.Client
	tenantAccessToken string
}

// WithHTTPClient overrides the HTTP client used by the SDK.
func WithHTTPClient(httpClient *http.Client) Option {
	return func(o *options) {
		o.httpClient = httpClient
	}
}

// WithTenantAccessToken sets a default tenant access token for requests.
func WithTenantAccessToken(token string) Option {
	return func(o *options) {
		o.tenantAccessToken = token
	}
}

type Client struct {
	sdk               *lark.Client
	coreConfig        *larkcore.Config
	tenantAccessToken string
}

func New(cfg *config.Config, opts ...Option) (*Client, error) {
	if cfg == nil {
		return nil, ErrUnavailable
	}
	if cfg.AppID == "" || cfg.AppSecret == "" {
		return nil, ErrUnavailable
	}
	settings := options{tenantAccessToken: cfg.TenantAccessToken}
	for _, opt := range opts {
		opt(&settings)
	}

	clientOptions := []lark.ClientOptionFunc{
		lark.WithEnableTokenCache(false),
	}
	coreConfig := &larkcore.Config{
		BaseUrl:          lark.FeishuBaseUrl,
		AppId:            cfg.AppID,
		AppSecret:        cfg.AppSecret,
		EnableTokenCache: false,
		AppType:          larkcore.AppTypeSelfBuilt,
	}
	if cfg.BaseURL != "" {
		clientOptions = append(clientOptions, lark.WithOpenBaseUrl(cfg.BaseURL))
		coreConfig.BaseUrl = cfg.BaseURL
	}
	if settings.httpClient != nil {
		clientOptions = append(clientOptions, lark.WithHttpClient(settings.httpClient))
		coreConfig.HttpClient = settings.httpClient
	}

	larkcore.NewLogger(coreConfig)
	larkcore.NewCache(coreConfig)
	larkcore.NewSerialization(coreConfig)
	larkcore.NewHttpClient(coreConfig)

	sdk := lark.NewClient(cfg.AppID, cfg.AppSecret, clientOptions...)
	return &Client{sdk: sdk, coreConfig: coreConfig, tenantAccessToken: settings.tenantAccessToken}, nil
}

func (c *Client) available() bool {
	return c != nil && c.sdk != nil
}

func (c *Client) tenantToken(token string) string {
	if token != "" {
		return token
	}
	if c == nil {
		return ""
	}
	return c.tenantAccessToken
}

func (c *Client) ListChats(ctx context.Context, token string, req ListChatsRequest) (ListChatsResult, error) {
	if !c.available() {
		return ListChatsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListChatsResult{}, errors.New("tenant access token is required")
	}

	builder := im.NewListChatReqBuilder()
	if req.PageSize > 0 {
		builder.PageSize(req.PageSize)
	}
	if req.PageToken != "" {
		builder.PageToken(req.PageToken)
	}
	if req.UserIDType != "" {
		builder.UserIdType(req.UserIDType)
	}

	resp, err := c.sdk.Im.V1.Chat.List(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListChatsResult{}, err
	}
	if resp == nil {
		return ListChatsResult{}, errors.New("list chats failed: empty response")
	}
	if !resp.Success() {
		return ListChatsResult{}, fmt.Errorf("list chats failed: %s", resp.Msg)
	}

	result := ListChatsResult{}
	if resp.Data != nil {
		if resp.Data.Items != nil {
			result.Items = make([]Chat, 0, len(resp.Data.Items))
			for _, item := range resp.Data.Items {
				result.Items = append(result.Items, mapChat(item))
			}
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}

func mapChat(chat *im.ListChat) Chat {
	if chat == nil {
		return Chat{}
	}
	result := Chat{}
	if chat.ChatId != nil {
		result.ChatID = *chat.ChatId
	}
	if chat.Avatar != nil {
		result.Avatar = *chat.Avatar
	}
	if chat.Name != nil {
		result.Name = *chat.Name
	}
	if chat.Description != nil {
		result.Description = *chat.Description
	}
	if chat.OwnerId != nil {
		result.OwnerID = *chat.OwnerId
	}
	if chat.OwnerIdType != nil {
		result.OwnerIDType = *chat.OwnerIdType
	}
	if chat.External != nil {
		result.External = *chat.External
	}
	if chat.TenantKey != nil {
		result.TenantKey = *chat.TenantKey
	}
	return result
}
