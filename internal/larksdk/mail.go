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
