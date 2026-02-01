package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type updateBaseFieldRequestBody struct {
	FieldName   string         `json:"field_name,omitempty"`
	Property    map[string]any `json:"property,omitempty"`
	Description map[string]any `json:"description,omitempty"`
}

type updateBaseFieldResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateBaseFieldResponseData `json:"data"`
}

type updateBaseFieldResponseData struct {
	Field BaseField `json:"field"`
}

func (r *updateBaseFieldResponse) Success() bool { return r.Code == 0 }

func (c *Client) UpdateBaseField(ctx context.Context, token, appToken, tableID, fieldID, fieldName string, property, description map[string]any) (BaseField, error) {
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
	if fieldID == "" {
		return BaseField{}, errors.New("field id is required")
	}
	if fieldName == "" && property == nil && description == nil {
		return BaseField{}, errors.New("at least one update field is required")
	}

	// SDK-first for the rename-only use-case. For advanced payloads, use core.
	if property == nil && description == nil && fieldName != "" && c.bitableFieldUpdateSDKAvailable() {
		field, err := c.updateBaseFieldSDK(ctx, tenantToken, appToken, tableID, fieldID, fieldName)
		if err == nil {
			return field, nil
		}
	}
	return c.updateBaseFieldCore(ctx, tenantToken, appToken, tableID, fieldID, fieldName, property, description)
}

func (c *Client) bitableFieldUpdateSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableField != nil
}

func (c *Client) updateBaseFieldSDK(ctx context.Context, tenantToken, appToken, tableID, fieldID, fieldName string) (BaseField, error) {
	field := larkbitable.NewAppTableFieldBuilder().FieldName(fieldName).Build()
	req := larkbitable.NewUpdateAppTableFieldReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		FieldId(fieldID).
		AppTableField(field).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableField.Update(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseField{}, err
	}
	if resp == nil {
		return BaseField{}, errors.New("update base field failed: empty response")
	}
	if !resp.Success() {
		return BaseField{}, formatCodeError("update base field failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil || resp.Data.Field == nil {
		return BaseField{}, nil
	}
	return mapBaseFieldFromSDK(resp.Data.Field), nil
}

func (c *Client) updateBaseFieldCore(ctx context.Context, tenantToken, appToken, tableID, fieldID, fieldName string, property, description map[string]any) (BaseField, error) {
	body := updateBaseFieldRequestBody{FieldName: fieldName, Property: property, Description: description}
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields/:field_id",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
		Body:                      body,
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("field_id", fieldID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseField{}, err
	}
	if apiResp == nil {
		return BaseField{}, errors.New("update base field failed: empty response")
	}
	resp := &updateBaseFieldResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseField{}, err
	}
	if !resp.Success() {
		return BaseField{}, formatCodeError("update base field failed", resp.CodeError, resp.ApiResp)
	}
	if resp.Data == nil {
		return BaseField{}, nil
	}
	return resp.Data.Field, nil
}
