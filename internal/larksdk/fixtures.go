package larksdk

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

// NOTE: The helpers in this file are primarily used by integration tests.

type createDriveFolderResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createDriveFolderResponseData `json:"data"`
}

type createDriveFolderResponseData struct {
	FolderToken *string `json:"folder_token"`
}

func (r *createDriveFolderResponse) Success() bool { return r.Code == 0 }

func (c *Client) CreateDriveFolder(ctx context.Context, token string, name string, parentFolderToken string) (string, error) {
	if !c.available() || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	if name == "" {
		return "", errors.New("folder name is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}
	payload := map[string]any{
		"name": name,
	}
	if parentFolderToken != "" {
		payload["folder_token"] = parentFolderToken
	}
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/files/create_folder",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("create drive folder failed: empty response")
	}
	resp := &createDriveFolderResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("create drive folder failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.FolderToken == nil {
		return "", errors.New("create drive folder failed: missing folder_token")
	}
	return *resp.Data.FolderToken, nil
}

type deleteResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *deleteResponse) Success() bool { return r.Code == 0 }

type createSpreadsheetResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createSpreadsheetResponseData `json:"data"`
}

type createSpreadsheetResponseData struct {
	Spreadsheet *SpreadsheetInfo `json:"spreadsheet"`
}

type SpreadsheetInfo struct {
	SpreadsheetToken string `json:"spreadsheet_token,omitempty"`
	Title            string `json:"title,omitempty"`
	FolderToken      string `json:"folder_token,omitempty"`
	URL              string `json:"url,omitempty"`
	WithoutMount     *bool  `json:"without_mount,omitempty"`
	OwnerID          string `json:"owner_id,omitempty"`
}

func (r *createSpreadsheetResponse) Success() bool { return r.Code == 0 }

// CreateSpreadsheet creates a new spreadsheet and returns its spreadsheet_token.
func (c *Client) CreateSpreadsheet(ctx context.Context, token string, title string, folderToken string) (string, error) {
	if !c.available() || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	if title == "" {
		return "", errors.New("spreadsheet title is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}

	payload := map[string]any{
		"title": title,
	}
	if folderToken != "" {
		payload["folder_token"] = folderToken
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/sheets/v3/spreadsheets",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("create spreadsheet failed: empty response")
	}
	resp := &createSpreadsheetResponse{ApiResp: apiResp}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("create spreadsheet failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Spreadsheet == nil || resp.Data.Spreadsheet.SpreadsheetToken == "" {
		return "", errors.New("create spreadsheet failed: missing spreadsheet_token")
	}
	return resp.Data.Spreadsheet.SpreadsheetToken, nil
}

// CreateChat creates a group chat. userIDs should be user_id values.
func (c *Client) CreateChat(ctx context.Context, token string, name string, userIDs []string) (string, error) {
	chat, err := c.CreateChatDetail(ctx, token, CreateChatRequest{
		Name:       name,
		UserIDList: userIDs,
	})
	if err != nil {
		return "", err
	}
	if chat.ChatID == "" {
		return "", errors.New("create chat failed: missing chat_id")
	}
	return chat.ChatID, nil
}

// Best-effort delete chat. This might not be supported for all app types/permissions.
func (c *Client) DeleteChat(ctx context.Context, token string, chatID string) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	if chatID == "" {
		return errors.New("chat id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/im/v1/chats/:chat_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("chat_id", chatID)
	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("delete chat failed: empty response")
	}
	resp := &deleteResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("delete chat failed: %s", resp.Msg)
	}
	return nil
}
