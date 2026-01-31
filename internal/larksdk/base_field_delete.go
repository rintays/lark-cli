package larksdk

import (
	"context"
	"errors"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
	larkbitable "github.com/larksuite/oapi-sdk-go/v3/service/bitable/v1"
)

type deleteBaseFieldResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *deleteBaseFieldResponseData `json:"data"`
}

type deleteBaseFieldResponseData struct {
	FieldID string `json:"field_id"`
	Deleted bool   `json:"deleted"`
}

func (r *deleteBaseFieldResponse) Success() bool { return r.Code == 0 }

func (c *Client) DeleteBaseField(ctx context.Context, token, appToken, tableID, fieldID string) (BaseFieldDeleteResult, error) {
	if !c.available() || c.coreConfig == nil {
		return BaseFieldDeleteResult{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return BaseFieldDeleteResult{}, errors.New("tenant access token is required")
	}
	if appToken == "" {
		return BaseFieldDeleteResult{}, errors.New("app token is required")
	}
	if tableID == "" {
		return BaseFieldDeleteResult{}, errors.New("table id is required")
	}
	if fieldID == "" {
		return BaseFieldDeleteResult{}, errors.New("field id is required")
	}

	if c.bitableFieldDeleteSDKAvailable() {
		return c.deleteBaseFieldSDK(ctx, tenantToken, appToken, tableID, fieldID)
	}
	return c.deleteBaseFieldCore(ctx, tenantToken, appToken, tableID, fieldID)
}

func (c *Client) bitableFieldDeleteSDKAvailable() bool {
	return c != nil && c.sdk != nil && c.sdk.Bitable != nil && c.sdk.Bitable.V1 != nil && c.sdk.Bitable.V1.AppTableField != nil
}

func (c *Client) deleteBaseFieldSDK(ctx context.Context, tenantToken, appToken, tableID, fieldID string) (BaseFieldDeleteResult, error) {
	req := larkbitable.NewDeleteAppTableFieldReqBuilder().
		AppToken(appToken).
		TableId(tableID).
		FieldId(fieldID).
		Build()
	resp, err := c.sdk.Bitable.V1.AppTableField.Delete(ctx, req, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseFieldDeleteResult{}, err
	}
	if resp == nil {
		return BaseFieldDeleteResult{}, errors.New("delete base field failed: empty response")
	}
	if !resp.Success() {
		return BaseFieldDeleteResult{}, formatCodeError("delete base field failed", resp.CodeError, resp.ApiResp)
	}
	result := BaseFieldDeleteResult{FieldID: fieldID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.FieldId != nil && *resp.Data.FieldId != "" {
			result.FieldID = *resp.Data.FieldId
		}
		if resp.Data.Deleted != nil {
			result.Deleted = *resp.Data.Deleted
		}
	}
	return result, nil
}

func (c *Client) deleteBaseFieldCore(ctx context.Context, tenantToken, appToken, tableID, fieldID string) (BaseFieldDeleteResult, error) {
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields/:field_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant},
	}
	apiReq.PathParams.Set("app_token", appToken)
	apiReq.PathParams.Set("table_id", tableID)
	apiReq.PathParams.Set("field_id", fieldID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return BaseFieldDeleteResult{}, err
	}
	if apiResp == nil {
		return BaseFieldDeleteResult{}, errors.New("delete base field failed: empty response")
	}
	resp := &deleteBaseFieldResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return BaseFieldDeleteResult{}, err
	}
	if !resp.Success() {
		return BaseFieldDeleteResult{}, formatCodeError("delete base field failed", resp.CodeError, resp.ApiResp)
	}

	result := BaseFieldDeleteResult{FieldID: fieldID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.FieldID != "" {
			result.FieldID = resp.Data.FieldID
		}
		result.Deleted = resp.Data.Deleted
	}
	return result, nil
}
