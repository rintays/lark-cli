package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type batchCreateBaseRecordsRequestBody struct {
	Records []batchCreateBaseRecordRequestBodyRecord `json:"records"`
}

type batchCreateBaseRecordRequestBodyRecord struct {
	Fields map[string]any `json:"fields"`
}

type batchCreateBaseRecordsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *batchCreateBaseRecordsResponseData `json:"data"`
}

type batchCreateBaseRecordsResponseData struct {
	Records []BaseRecord `json:"records"`
}

func (r *batchCreateBaseRecordsResponse) Success() bool { return r.Code == 0 }

func (c *Client) BatchCreateBaseRecords(ctx context.Context, token, appToken, tableID string, records []map[string]any, clientToken string, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
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
	if len(records) == 0 {
		return nil, errors.New("records are required")
	}
	for _, fields := range records {
		if fields == nil {
			return nil, errors.New("records include nil fields")
		}
		if len(fields) == 0 {
			return nil, errors.New("records include empty fields")
		}
	}

	if c.bitableRecordCreateSDKAvailable() {
		return c.batchCreateBaseRecordsSDK(ctx, tenantToken, appToken, tableID, records, clientToken, ignoreConsistencyCheck)
	}
	return c.batchCreateBaseRecordsCore(ctx, tenantToken, appToken, tableID, records, clientToken, ignoreConsistencyCheck)
}

func (c *Client) batchCreateBaseRecordsSDK(ctx context.Context, tenantToken, appToken, tableID string, records []map[string]any, clientToken string, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
	sdkRecords := make([]*larkbitable.AppTableRecord, 0, len(records))
	for _, fields := range records {
		sdkRecords = append(sdkRecords, larkbitable.NewAppTableRecordBuilder().Fields(fields).Build())
	}

	body := larkbitable.NewBatchCreateAppTableRecordReqBodyBuilder().Records(sdkRecords).Build()
	builder := larkbitable.NewBatchCreateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		Body(body)
	if clientToken != "" {
		builder.ClientToken(clientToken)
	}
	if ignoreConsistencyCheck {
		builder.IgnoreConsistencyCheck(true)
	}

	resp, err := c.sdk.Bitable.V1.AppTableRecord.BatchCreate(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch create base records failed: empty response")
	}
	if !resp.Success() {
		return nil, formatCodeError("batch create base records failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || len(resp.Data.Records) == 0 {
		return nil, nil
	}
	out := make([]BaseRecord, 0, len(resp.Data.Records))
	for _, record := range resp.Data.Records {
		if record == nil {
			continue
		}
		out = append(out, mapBaseRecordFromSDK(record))
	}
	return out, nil
}

func (c *Client) batchCreateBaseRecordsCore(ctx context.Context, tenantToken, appToken, tableID string, records []map[string]any, clientToken string, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
	reqs := make([]batchCreateBaseRecordRequestBodyRecord, 0, len(records))
	for _, fields := range records {
		reqs = append(reqs, batchCreateBaseRecordRequestBodyRecord{Fields: fields})
	}
	body := batchCreateBaseRecordsRequestBody{Records: reqs}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/batch_create",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		Body:                      body,
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	if clientToken != "" {
		apiReq.QueryParams.Set("client_token", clientToken)
	}
	if ignoreConsistencyCheck {
		apiReq.QueryParams.Set("ignore_consistency_check", "true")
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if apiResp == nil {
		return nil, errors.New("batch create base records failed: empty response")
	}
	resp := &batchCreateBaseRecordsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, formatCodeError("batch create base records failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return nil, nil
	}
	return resp.Data.Records, nil
}
