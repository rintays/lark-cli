package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type batchDeleteBaseRecordsRequestBody struct {
	Records []string `json:"records"`
}

type batchDeleteBaseRecordsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *batchDeleteBaseRecordsResponseData `json:"data"`
}

type batchDeleteBaseRecordsResponseData struct {
	Records []BaseRecordDeleteResult `json:"records"`
}

func (r *batchDeleteBaseRecordsResponse) Success() bool { return r.Code == 0 }

func (c *Client) BatchDeleteBaseRecords(ctx context.Context, token, appToken, tableID string, recordIDs []string) ([]BaseRecordDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return nil, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return nil, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return nil, errors.New("app token is required")
	}
	if tableID == "" {
		return nil, errors.New("table id is required")
	}
	if len(recordIDs) == 0 {
		return nil, errors.New("record ids are required")
	}

	if c.bitableRecordDeleteSDKAvailable() {
		return c.batchDeleteBaseRecordsSDK(ctx, tenantToken, appToken, tableID, recordIDs)
	}
	return c.batchDeleteBaseRecordsCore(ctx, tenantToken, appToken, tableID, recordIDs)
}

func (c *Client) batchDeleteBaseRecordsSDK(ctx context.Context, tenantToken, appToken, tableID string, recordIDs []string) ([]BaseRecordDeleteResult, error) {
	body, err := larkbitable.NewBatchDeleteAppTableRecordPathReqBodyBuilder().Records(recordIDs).Build()
	if err != nil {
		return nil, err
	}
	req := larkbitable.NewBatchDeleteAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		Body(body).
		Build()

	resp, err := c.sdk.Bitable.V1.AppTableRecord.BatchDelete(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch delete base records failed: empty response")
	}
	if !resp.Success() {
		return nil, formatCodeError("batch delete base records failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || len(resp.Data.Records) == 0 {
		out := make([]BaseRecordDeleteResult, 0, len(recordIDs))
		for _, id := range recordIDs {
			out = append(out, BaseRecordDeleteResult{RecordID: id, Deleted: true})
		}
		return out, nil
	}
	out := make([]BaseRecordDeleteResult, 0, len(resp.Data.Records))
	for _, item := range resp.Data.Records {
		result := BaseRecordDeleteResult{}
		if item != nil {
			if item.RecordId != nil {
				result.RecordID = *item.RecordId
			}
			if item.Deleted != nil {
				result.Deleted = *item.Deleted
			}
		}
		out = append(out, result)
	}
	return out, nil
}

func (c *Client) batchDeleteBaseRecordsCore(ctx context.Context, tenantToken, appToken, tableID string, recordIDs []string) ([]BaseRecordDeleteResult, error) {
	body := batchDeleteBaseRecordsRequestBody{Records: recordIDs}
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/batch_delete",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		Body:                      body,
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if apiResp == nil {
		return nil, errors.New("batch delete base records failed: empty response")
	}
	resp := &batchDeleteBaseRecordsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, formatCodeError("batch delete base records failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		out := make([]BaseRecordDeleteResult, 0, len(recordIDs))
		for _, id := range recordIDs {
			out = append(out, BaseRecordDeleteResult{RecordID: id, Deleted: true})
		}
		return out, nil
	}
	return resp.Data.Records, nil
}
