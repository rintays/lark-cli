package larksdk

import (
	"context"
	"errors"
	"net/http"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type batchUpdateBaseRecordsRequestBody struct {
	Records []batchUpdateBaseRecordRequestBodyRecord `json:"records"`
}

type batchUpdateBaseRecordRequestBodyRecord struct {
	RecordID string         `json:"record_id"`
	Fields   map[string]any `json:"fields"`
}

type batchUpdateBaseRecordsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *batchUpdateBaseRecordsResponseData `json:"data"`
}

type batchUpdateBaseRecordsResponseData struct {
	Records []BaseRecord `json:"records"`
}

func (r *batchUpdateBaseRecordsResponse) Success() bool { return r.Code == 0 }

func (c *Client) BatchUpdateBaseRecords(ctx context.Context, token, appToken, tableID string, records []BaseRecordUpdate, clientToken string, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
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
	normalized := make([]BaseRecordUpdate, 0, len(records))
	for _, record := range records {
		recordID := strings.TrimSpace(record.RecordID)
		if recordID == "" {
			return nil, errors.New("records include empty record_id")
		}
		if record.Fields == nil {
			return nil, errors.New("records include nil fields")
		}
		if len(record.Fields) == 0 {
			return nil, errors.New("records include empty fields")
		}
		normalized = append(normalized, BaseRecordUpdate{RecordID: recordID, Fields: record.Fields})
	}

	// SDK doesn't currently expose client_token for this endpoint; fall back to core when needed.
	if clientToken == "" && c.bitableRecordUpdateSDKAvailable() {
		return c.batchUpdateBaseRecordsSDK(ctx, tenantToken, appToken, tableID, normalized, ignoreConsistencyCheck)
	}
	return c.batchUpdateBaseRecordsCore(ctx, tenantToken, appToken, tableID, normalized, clientToken, ignoreConsistencyCheck)
}

func (c *Client) batchUpdateBaseRecordsSDK(ctx context.Context, tenantToken, appToken, tableID string, records []BaseRecordUpdate, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
	sdkRecords := make([]*larkbitable.AppTableRecord, 0, len(records))
	for _, record := range records {
		sdkRecords = append(sdkRecords, larkbitable.NewAppTableRecordBuilder().RecordId(record.RecordID).Fields(record.Fields).Build())
	}

	body := larkbitable.NewBatchUpdateAppTableRecordReqBodyBuilder().Records(sdkRecords).Build()
	builder := larkbitable.NewBatchUpdateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		Body(body)
	if ignoreConsistencyCheck {
		builder.IgnoreConsistencyCheck(true)
	}

	resp, err := c.sdk.Bitable.V1.AppTableRecord.BatchUpdate(ctx, builder.Build(), larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return nil, err
	}
	if resp == nil {
		return nil, errors.New("batch update base records failed: empty response")
	}
	if !resp.Success() {
		return nil, formatCodeError("batch update base records failed", resp.CodeError, resp.ApiResp)
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

func (c *Client) batchUpdateBaseRecordsCore(ctx context.Context, tenantToken, appToken, tableID string, records []BaseRecordUpdate, clientToken string, ignoreConsistencyCheck bool) ([]BaseRecord, error) {
	reqs := make([]batchUpdateBaseRecordRequestBodyRecord, 0, len(records))
	for _, record := range records {
		reqs = append(reqs, batchUpdateBaseRecordRequestBodyRecord{RecordID: record.RecordID, Fields: record.Fields})
	}
	body := batchUpdateBaseRecordsRequestBody{Records: reqs}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/batch_update",
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
		return nil, errors.New("batch update base records failed: empty response")
	}
	resp := &batchUpdateBaseRecordsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return nil, err
	}
	if !resp.Success() {
		return nil, formatCodeError("batch update base records failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return nil, nil
	}
	return resp.Data.Records, nil
}
