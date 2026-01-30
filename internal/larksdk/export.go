package larksdk

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"lark/internal/larkapi"
)

type createExportTaskResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createExportTaskResponseData `json:"data"`
}

type createExportTaskResponseData struct {
	Ticket string `json:"ticket"`
}

func (r *createExportTaskResponse) Success() bool {
	return r.Code == 0
}

type getExportTaskResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getExportTaskResponseData `json:"data"`
}

type getExportTaskResponseData struct {
	Result larkapi.ExportTaskResult `json:"result"`
}

func (r *getExportTaskResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) CreateExportTask(ctx context.Context, token string, req larkapi.CreateExportTaskRequest) (string, error) {
	if !c.available() || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	if req.Token == "" {
		return "", fmt.Errorf("export token is required")
	}
	if req.Type == "" {
		return "", fmt.Errorf("export type is required")
	}
	if req.FileExtension == "" {
		return "", fmt.Errorf("file extension is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return "", errors.New("tenant access token is required")
	}

	payload := map[string]any{
		"token":          req.Token,
		"type":           req.Type,
		"file_extension": req.FileExtension,
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/export_tasks",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.Body = payload

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return "", err
	}
	if apiResp == nil {
		return "", errors.New("create export task failed: empty response")
	}
	resp := &createExportTaskResponse{ApiResp: apiResp}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return "", err
	}
	if !resp.Success() {
		return "", fmt.Errorf("create export task failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Ticket == "" {
		return "", errors.New("export task response missing ticket")
	}
	return resp.Data.Ticket, nil
}

func (c *Client) GetExportTask(ctx context.Context, token, ticket string) (larkapi.ExportTaskResult, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.ExportTaskResult{}, ErrUnavailable
	}
	if ticket == "" {
		return larkapi.ExportTaskResult{}, fmt.Errorf("export ticket is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.ExportTaskResult{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/drive/v1/export_tasks/:ticket",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("ticket", ticket)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.ExportTaskResult{}, err
	}
	if apiResp == nil {
		return larkapi.ExportTaskResult{}, errors.New("get export task failed: empty response")
	}
	resp := &getExportTaskResponse{ApiResp: apiResp}
	if err := json.Unmarshal(apiResp.RawBody, resp); err != nil {
		return larkapi.ExportTaskResult{}, err
	}
	if !resp.Success() {
		return larkapi.ExportTaskResult{}, fmt.Errorf("get export task failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return larkapi.ExportTaskResult{}, nil
	}
	return resp.Data.Result, nil
}

func (c *Client) DownloadExportedFile(ctx context.Context, token, fileToken string) (io.ReadCloser, error) {
	if !c.available() || c.coreConfig == nil {
		return nil, ErrUnavailable
	}
	if fileToken == "" {
		return nil, fmt.Errorf("export file token is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/export_tasks/file/" + url.PathEscape(fileToken) + "/download")
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+tenantToken)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("export download failed: %s", resp.Status)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("export download failed: %s", resp.Status)
		}
		return nil, fmt.Errorf("export download failed: %s: %s", resp.Status, string(bytes.TrimSpace(data)))
	}
	return resp.Body, nil
}

func (c *Client) endpoint(path string) (string, error) {
	if c == nil || c.coreConfig == nil {
		return "", ErrUnavailable
	}
	base, err := url.Parse(c.coreConfig.BaseUrl)
	if err != nil {
		return "", err
	}
	base.Path = path
	base.RawQuery = ""
	return base.String(), nil
}

func (c *Client) httpClient() larkcore.HttpClient {
	if c != nil && c.coreConfig != nil && c.coreConfig.HttpClient != nil {
		return c.coreConfig.HttpClient
	}
	return http.DefaultClient
}
