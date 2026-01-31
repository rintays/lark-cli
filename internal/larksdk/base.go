package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type listBaseTablesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listBaseTablesResponseData `json:"data"`
}

type listBaseTablesResponseData struct {
	Items     []BaseTable `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

func (r *listBaseTablesResponse) Success() bool { return r.Code == 0 }

func (c *Client) ListBaseTables(ctx context.Context, token, appToken string) (ListBaseTablesResult, error) {
	return c.ListBaseTablesPage(ctx, token, appToken, "", 0)
}

func (c *Client) ListBaseTablesPage(ctx context.Context, token, appToken, pageToken string, pageSize int) (ListBaseTablesResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListBaseTablesResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListBaseTablesResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return ListBaseTablesResult{}, errors.New("app token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	if pageToken != "" {
		apiReq.QueryParams.Set("page_token", pageToken)
	}
	if pageSize > 0 {
		apiReq.QueryParams.Set("page_size", strconv.Itoa(pageSize))
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListBaseTablesResult{}, err
	}
	if apiResp == nil {
		return ListBaseTablesResult{}, errors.New("list base tables failed: empty response")
	}
	resp := &listBaseTablesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListBaseTablesResult{}, err
	}
	if !resp.Success() {
		return ListBaseTablesResult{}, fmt.Errorf("list base tables failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListBaseTablesResult{}, nil
	}
	return ListBaseTablesResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}

func (c *Client) ListBaseTablesAll(ctx context.Context, token, appToken string) ([]BaseTable, error) {
	items := make([]BaseTable, 0)
	pageToken := ""
	for {
		res, err := c.ListBaseTablesPage(ctx, token, appToken, pageToken, 100)
		if err != nil {
			return nil, err
		}
		items = append(items, res.Items...)
		if !res.HasMore || res.PageToken == "" {
			break
		}
		pageToken = res.PageToken
	}
	return items, nil
}

type listBaseFieldsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listBaseFieldsResponseData `json:"data"`
}

type listBaseFieldsResponseData struct {
	Items     []BaseField `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

func (r *listBaseFieldsResponse) Success() bool { return r.Code == 0 }

func (c *Client) ListBaseFields(ctx context.Context, token, appToken, tableID string) (ListBaseFieldsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListBaseFieldsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListBaseFieldsResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return ListBaseFieldsResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return ListBaseFieldsResult{}, errors.New("table id is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListBaseFieldsResult{}, err
	}
	if apiResp == nil {
		return ListBaseFieldsResult{}, errors.New("list base fields failed: empty response")
	}
	resp := &listBaseFieldsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListBaseFieldsResult{}, err
	}
	if !resp.Success() {
		return ListBaseFieldsResult{}, fmt.Errorf("list base fields failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListBaseFieldsResult{}, nil
	}
	return ListBaseFieldsResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}

type createBaseFieldRequestBody struct {
	FieldName   string         `json:"field_name"`
	Type        int            `json:"type"`
	Property    map[string]any `json:"property,omitempty"`
	Description map[string]any `json:"description,omitempty"`
}

type createBaseFieldResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createBaseFieldResponseData `json:"data"`
}

type createBaseFieldResponseData struct {
	Field BaseField `json:"field"`
}

func (r *createBaseFieldResponse) Success() bool { return r.Code == 0 }

func (c *Client) CreateBaseField(ctx context.Context, token, appToken, tableID, fieldName string, fieldType int, property, description map[string]any) (BaseField, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseField{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseField{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseField{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseField{}, errors.New("table id is required")
	}
	if fieldName == "" {
		return BaseField{}, errors.New("field name is required")
	}
	if fieldType == 0 {
		return BaseField{}, errors.New("field type is required")
	}

	// SDK-first for the minimal field create use-case. If the typed SDK service is
	// not available (or if advanced property/description payloads are used), fall
	// back to a core.ApiReq wrapper.
	if property == nil && description == nil && c.bitableFieldCreateSDKAvailable() {
		field, err := c.createBaseFieldSDK(ctx, tenantToken, appToken, tableID, fieldName, fieldType)
		if err == nil {
			return field, nil
		}
	}
	return c.createBaseFieldCore(ctx, tenantToken, appToken, tableID, fieldName, fieldType, property, description)
}

func (c *Client) bitableFieldCreateSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableField != nil
}

func (c *Client) createBaseFieldSDK(ctx context.Context, tenantToken, appToken, tableID, fieldName string, fieldType int) (BaseField, error) {
	field := larkbitable.NewAppTableFieldBuilder().FieldName(fieldName).Type(fieldType).Build()
	req := larkbitable.NewCreateAppTableFieldReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		AppTableField(field).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableField.Create(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseField{}, err
	}
	if resp == nil {
		return BaseField{}, errors.New("create base field failed: empty response")
	}
	if !resp.Success() {
		return BaseField{}, formatCodeError("create base field failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.Field == nil {
		return BaseField{}, nil
	}
	return mapBaseFieldFromSDK(resp.Data.Field), nil
}

func mapBaseFieldFromSDK(field *larkbitable.AppTableField) BaseField {
	if field == nil {
		return BaseField{}
	}
	result := BaseField{}
	if field.FieldId != nil {
		result.FieldID = *field.FieldId
	}
	if field.FieldName != nil {
		result.FieldName = *field.FieldName
	}
	if field.Type != nil {
		result.Type = *field.Type
	}
	return result
}

func (c *Client) createBaseFieldCore(ctx context.Context, tenantToken, appToken, tableID, fieldName string, fieldType int, property, description map[string]any) (BaseField, error) {
	body := createBaseFieldRequestBody{
		FieldName:   fieldName,
		Type:        fieldType,
		Property:    property,
		Description: description,
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields",
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
		return BaseField{}, err
	}
	if apiResp == nil {
		return BaseField{}, errors.New("create base field failed: empty response")
	}
	resp := &createBaseFieldResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseField{}, err
	}
	if !resp.Success() {
		return BaseField{}, formatCodeError("create base field failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return BaseField{}, nil
	}
	return resp.Data.Field, nil
}

type listBaseViewsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listBaseViewsResponseData `json:"data"`
}

type listBaseViewsResponseData struct {
	Items     []BaseView `json:"items"`
	PageToken string     `json:"page_token"`
	HasMore   bool       `json:"has_more"`
}

func (r *listBaseViewsResponse) Success() bool { return r.Code == 0 }

func (c *Client) ListBaseViews(ctx context.Context, token, appToken, tableID string) (ListBaseViewsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListBaseViewsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListBaseViewsResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return ListBaseViewsResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return ListBaseViewsResult{}, errors.New("table id is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/views",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListBaseViewsResult{}, err
	}
	if apiResp == nil {
		return ListBaseViewsResult{}, errors.New("list base views failed: empty response")
	}
	resp := &listBaseViewsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListBaseViewsResult{}, err
	}
	if !resp.Success() {
		return ListBaseViewsResult{}, fmt.Errorf("list base views failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return ListBaseViewsResult{}, nil
	}
	return ListBaseViewsResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}

func (c *Client) CreateBaseView(ctx context.Context, token, appToken, tableID, viewName, viewType string) (BaseView, error) {
	if !c.available() || c.sdk == nil || c.sdk.Bitable == nil || c.sdk.Bitable.V1 == nil || c.sdk.Bitable.V1.AppTableView == nil {
		return BaseView{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseView{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseView{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseView{}, errors.New("table id is required")
	}
	if viewName == "" {
		return BaseView{}, errors.New("view name is required")
	}

	viewBuilder := larkbitable.NewReqViewBuilder().ViewName(viewName)
	if viewType != "" {
		viewBuilder.ViewType(viewType)
	}
	req := larkbitable.NewCreateAppTableViewReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		ReqView(viewBuilder.Build()).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableView.Create(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseView{}, err
	}
	if resp == nil {
		return BaseView{}, errors.New("create base view failed: empty response")
	}
	if !resp.Success() {
		return BaseView{}, formatCodeError("create base view failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.View == nil {
		return BaseView{}, nil
	}
	return mapBaseViewFromSDK(resp.Data.View), nil
}

type createBaseRecordRequestBody struct {
	Fields map[string]any `json:"fields"`
}

type createBaseRecordResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createBaseRecordResponseData `json:"data"`
}

type createBaseRecordResponseData struct {
	Record BaseRecord `json:"record"`
}

func (r *createBaseRecordResponse) Success() bool { return r.Code == 0 }

func (c *Client) CreateBaseRecord(ctx context.Context, token, appToken, tableID string, fields map[string]any) (BaseRecord, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseRecord{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseRecord{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseRecord{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseRecord{}, errors.New("table id is required")
	}
	if fields == nil {
		return BaseRecord{}, errors.New("fields are required")
	}
	if c.bitableRecordCreateSDKAvailable() {
		return c.createBaseRecordSDK(ctx, tenantToken, appToken, tableID, fields)
	}
	return c.createBaseRecordCore(ctx, tenantToken, appToken, tableID, fields)
}

func (c *Client) bitableRecordCreateSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableRecord != nil
}

func (c *Client) createBaseRecordSDK(ctx context.Context, tenantToken, appToken, tableID string, fields map[string]any) (BaseRecord, error) {
	record := larkbitable.NewAppTableRecordBuilder().Fields(fields).Build()
	req := larkbitable.NewCreateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		AppTableRecord(record).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableRecord.Create(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecord{}, err
	}
	if resp == nil {
		return BaseRecord{}, errors.New("create base record failed: empty response")
	}
	if !resp.Success() {
		return BaseRecord{}, formatCodeError("create base record failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.Record == nil {
		return BaseRecord{}, nil
	}
	return mapBaseRecordFromSDK(resp.Data.Record), nil
}

func (c *Client) createBaseRecordCore(ctx context.Context, tenantToken, appToken, tableID string, fields map[string]any) (BaseRecord, error) {
	body := createBaseRecordRequestBody{
		Fields: fields,
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records",
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
		return BaseRecord{}, err
	}
	if apiResp == nil {
		return BaseRecord{}, errors.New("create base record failed: empty response")
	}
	resp := &createBaseRecordResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseRecord{}, err
	}
	if !resp.Success() {
		return BaseRecord{}, formatCodeError("create base record failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return BaseRecord{}, nil
	}
	return resp.Data.Record, nil
}

type getBaseRecordResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getBaseRecordResponseData `json:"data"`
}

type getBaseRecordResponseData struct {
	Record BaseRecord `json:"record"`
}

func (r *getBaseRecordResponse) Success() bool { return r.Code == 0 }

func (c *Client) GetBaseRecord(ctx context.Context, token, appToken, tableID, recordID string) (BaseRecord, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseRecord{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseRecord{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseRecord{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseRecord{}, errors.New("table id is required")
	}
	if recordID == "" {
		return BaseRecord{}, errors.New("record id is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("record_id", recordID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecord{}, err
	}
	if apiResp == nil {
		return BaseRecord{}, errors.New("get base record failed: empty response")
	}
	resp := &getBaseRecordResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseRecord{}, err
	}
	if !resp.Success() {
		return BaseRecord{}, fmt.Errorf("get base record failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return BaseRecord{}, nil
	}
	return resp.Data.Record, nil
}

type updateBaseRecordRequestBody struct {
	Fields map[string]any `json:"fields"`
}

type updateBaseRecordResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateBaseRecordResponseData `json:"data"`
}

type updateBaseRecordResponseData struct {
	Record BaseRecord `json:"record"`
}

func (r *updateBaseRecordResponse) Success() bool { return r.Code == 0 }

type deleteBaseRecordResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *deleteBaseRecordResponseData `json:"data"`
}

type deleteBaseRecordResponseData struct {
	Deleted  bool   `json:"deleted"`
	RecordID string `json:"record_id"`
}

func (r *deleteBaseRecordResponse) Success() bool { return r.Code == 0 }

func (c *Client) UpdateBaseRecord(ctx context.Context, token, appToken, tableID, recordID string, fields map[string]any) (BaseRecord, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseRecord{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseRecord{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseRecord{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseRecord{}, errors.New("table id is required")
	}
	if recordID == "" {
		return BaseRecord{}, errors.New("record id is required")
	}
	if fields == nil {
		return BaseRecord{}, errors.New("fields are required")
	}
	if c.bitableRecordUpdateSDKAvailable() {
		return c.updateBaseRecordSDK(ctx, tenantToken, appToken, tableID, recordID, fields)
	}
	return c.updateBaseRecordCore(ctx, tenantToken, appToken, tableID, recordID, fields)
}

func (c *Client) bitableRecordUpdateSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableRecord != nil
}

func (c *Client) updateBaseRecordSDK(ctx context.Context, tenantToken, appToken, tableID, recordID string, fields map[string]any) (BaseRecord, error) {
	record := larkbitable.NewAppTableRecordBuilder().Fields(fields).Build()
	req := larkbitable.NewUpdateAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		RecordId(recordID).
		AppTableRecord(record).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableRecord.Update(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecord{}, err
	}
	if resp == nil {
		return BaseRecord{}, errors.New("update base record failed: empty response")
	}
	if !resp.Success() {
		return BaseRecord{}, formatCodeError("update base record failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.Record == nil {
		return BaseRecord{}, nil
	}
	result := mapBaseRecordFromSDK(resp.Data.Record)
	if result.RecordID == "" {
		result.RecordID = recordID
	}
	return result, nil
}

func (c *Client) updateBaseRecordCore(ctx context.Context, tenantToken, appToken, tableID, recordID string, fields map[string]any) (BaseRecord, error) {
	body := updateBaseRecordRequestBody{
		Fields: fields,
	}
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		Body:                      body,
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("record_id", recordID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecord{}, err
	}
	if apiResp == nil {
		return BaseRecord{}, errors.New("update base record failed: empty response")
	}
	resp := &updateBaseRecordResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseRecord{}, err
	}
	if !resp.Success() {
		return BaseRecord{}, formatCodeError("update base record failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return BaseRecord{}, nil
	}
	return resp.Data.Record, nil
}

func (c *Client) DeleteBaseRecord(ctx context.Context, token, appToken, tableID, recordID string) (BaseRecordDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseRecordDeleteResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseRecordDeleteResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseRecordDeleteResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseRecordDeleteResult{}, errors.New("table id is required")
	}
	if recordID == "" {
		return BaseRecordDeleteResult{}, errors.New("record id is required")
	}
	if c.bitableRecordDeleteSDKAvailable() {
		return c.deleteBaseRecordSDK(ctx, tenantToken, appToken, tableID, recordID)
	}
	return c.deleteBaseRecordCore(ctx, tenantToken, appToken, tableID, recordID)
}

func (c *Client) bitableRecordDeleteSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableRecord != nil
}

func (c *Client) deleteBaseRecordSDK(ctx context.Context, tenantToken, appToken, tableID, recordID string) (BaseRecordDeleteResult, error) {
	req := larkbitable.NewDeleteAppTableRecordReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		RecordId(recordID).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableRecord.Delete(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecordDeleteResult{}, err
	}
	if resp == nil {
		return BaseRecordDeleteResult{}, errors.New("delete base record failed: empty response")
	}
	if !resp.Success() {
		return BaseRecordDeleteResult{}, fmt.Errorf("delete base record failed: %s", resp.Msg)
	}
	result := BaseRecordDeleteResult{RecordID: recordID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.RecordId != nil && *resp.Data.RecordId != "" {
			result.RecordID = *resp.Data.RecordId
		}
		if resp.Data.Deleted != nil {
			result.Deleted = *resp.Data.Deleted
		}
	}
	return result, nil
}

func (c *Client) deleteBaseRecordCore(ctx context.Context, tenantToken, appToken, tableID, recordID string) (BaseRecordDeleteResult, error) {
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("record_id", recordID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseRecordDeleteResult{}, err
	}
	if apiResp == nil {
		return BaseRecordDeleteResult{}, errors.New("delete base record failed: empty response")
	}
	resp := &deleteBaseRecordResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseRecordDeleteResult{}, err
	}
	if !resp.Success() {
		return BaseRecordDeleteResult{}, fmt.Errorf("delete base record failed: %s", resp.Msg)
	}
	result := BaseRecordDeleteResult{RecordID: recordID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.RecordID != "" {
			result.RecordID = resp.Data.RecordID
		}
		result.Deleted = resp.Data.Deleted
	}
	return result, nil
}

func mapBaseRecordFromSDK(record *larkbitable.AppTableRecord) BaseRecord {
	if record == nil {
		return BaseRecord{}
	}
	result := BaseRecord{
		Fields: record.Fields,
	}
	if record.RecordId != nil {
		result.RecordID = *record.RecordId
	}
	if record.CreatedTime != nil {
		result.CreatedTime = strconv.FormatInt(*record.CreatedTime, 10)
	}
	if record.LastModifiedTime != nil {
		result.LastModifiedTime = strconv.FormatInt(*record.LastModifiedTime, 10)
	}
	return result
}

func mapBaseViewFromSDK(view *larkbitable.AppTableView) BaseView {
	if view == nil {
		return BaseView{}
	}
	result := BaseView{}
	if view.ViewId != nil {
		result.ViewID = *view.ViewId
	}
	if view.ViewName != nil {
		result.Name = *view.ViewName
	}
	if view.ViewType != nil {
		result.ViewType = *view.ViewType
	}
	return result
}

type searchBaseRecordsRequestBody struct {
	AppToken string `json:"app_token"`
	TableID  string `json:"table_id"`
	SearchBaseRecordsRequest
}

type searchBaseRecordsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *searchBaseRecordsResponseData `json:"data"`
}

type searchBaseRecordsResponseData struct {
	Items     []BaseRecord `json:"items"`
	PageToken string       `json:"page_token"`
	HasMore   bool         `json:"has_more"`
}

func (r *searchBaseRecordsResponse) Success() bool { return r.Code == 0 }

func (c *Client) SearchBaseRecords(ctx context.Context, token, appToken, tableID string, req SearchBaseRecordsRequest) (SearchBaseRecordsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SearchBaseRecordsResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return SearchBaseRecordsResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return SearchBaseRecordsResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return SearchBaseRecordsResult{}, errors.New("table id is required")
	}

	body := searchBaseRecordsRequestBody{
		AppToken:                 appToken,
		TableID:                  tableID,
		SearchBaseRecordsRequest: req,
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/search",
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
		return SearchBaseRecordsResult{}, err
	}
	if apiResp == nil {
		return SearchBaseRecordsResult{}, errors.New("search base records failed: empty response")
	}
	resp := &searchBaseRecordsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SearchBaseRecordsResult{}, err
	}
	if !resp.Success() {
		return SearchBaseRecordsResult{}, fmt.Errorf("search base records failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return SearchBaseRecordsResult{}, nil
	}
	return SearchBaseRecordsResult{Items: resp.Data.Items, PageToken: resp.Data.PageToken, HasMore: resp.Data.HasMore}, nil
}
