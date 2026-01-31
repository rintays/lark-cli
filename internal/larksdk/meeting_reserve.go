package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type ApplyReserveRequest struct {
	EndTime         string
	OwnerID         string
	UserIDType      string
	MeetingSettings *ReserveMeetingSetting
}

type UpdateReserveRequest struct {
	ReserveID       string
	EndTime         string
	UserIDType      string
	MeetingSettings *ReserveMeetingSetting
}

type DeleteReserveRequest struct {
	ReserveID string
}

type applyReserveResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *applyReserveResponseData `json:"data"`
}

type applyReserveResponseData struct {
	Reserve                    *Reserve                    `json:"reserve"`
	ReserveCorrectionCheckInfo *ReserveCorrectionCheckInfo `json:"reserve_correction_check_info,omitempty"`
}

func (r *applyReserveResponse) Success() bool {
	return r.Code == 0
}

type updateReserveResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateReserveResponseData `json:"data"`
}

type updateReserveResponseData struct {
	Reserve                    *Reserve                    `json:"reserve"`
	ReserveCorrectionCheckInfo *ReserveCorrectionCheckInfo `json:"reserve_correction_check_info,omitempty"`
}

func (r *updateReserveResponse) Success() bool {
	return r.Code == 0
}

type deleteReserveResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *deleteReserveResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) ApplyReserve(ctx context.Context, token string, req ApplyReserveRequest) (Reserve, *ReserveCorrectionCheckInfo, error) {
	if !c.available() || c.coreConfig == nil {
		return Reserve{}, nil, ErrUnavailable
	}
	if req.EndTime == "" {
		return Reserve{}, nil, errors.New("end time is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Reserve{}, nil, errors.New("tenant access token is required")
	}

	payload := map[string]any{
		"end_time": req.EndTime,
	}
	if req.OwnerID != "" {
		payload["owner_id"] = req.OwnerID
	}
	if req.MeetingSettings != nil {
		payload["meeting_settings"] = req.MeetingSettings
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/vc/v1/reserves/apply",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return Reserve{}, nil, err
	}
	if apiResp == nil {
		return Reserve{}, nil, errors.New("apply reserve failed: empty response")
	}
	resp := &applyReserveResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return Reserve{}, nil, err
	}
	if !resp.Success() {
		return Reserve{}, nil, fmt.Errorf("apply reserve failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Reserve == nil {
		return Reserve{}, nil, errors.New("apply reserve response missing reserve")
	}
	reserve := *resp.Data.Reserve
	if reserve.ID == "" {
		return Reserve{}, nil, errors.New("apply reserve response missing reserve id")
	}
	return reserve, resp.Data.ReserveCorrectionCheckInfo, nil
}

func (c *Client) UpdateReserve(ctx context.Context, token string, req UpdateReserveRequest) (Reserve, *ReserveCorrectionCheckInfo, error) {
	if !c.available() || c.coreConfig == nil {
		return Reserve{}, nil, ErrUnavailable
	}
	if req.ReserveID == "" {
		return Reserve{}, nil, errors.New("reserve id is required")
	}
	if req.EndTime == "" && req.MeetingSettings == nil {
		return Reserve{}, nil, errors.New("at least one field is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Reserve{}, nil, errors.New("tenant access token is required")
	}

	payload := map[string]any{}
	if req.EndTime != "" {
		payload["end_time"] = req.EndTime
	}
	if req.MeetingSettings != nil {
		payload["meeting_settings"] = req.MeetingSettings
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/vc/v1/reserves/:reserve_id",
		HttpMethod:                http.MethodPut,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("reserve_id", req.ReserveID)
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return Reserve{}, nil, err
	}
	if apiResp == nil {
		return Reserve{}, nil, errors.New("update reserve failed: empty response")
	}
	resp := &updateReserveResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return Reserve{}, nil, err
	}
	if !resp.Success() {
		return Reserve{}, nil, fmt.Errorf("update reserve failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Reserve == nil {
		return Reserve{}, nil, errors.New("update reserve response missing reserve")
	}
	reserve := *resp.Data.Reserve
	if reserve.ID == "" {
		return Reserve{}, nil, errors.New("update reserve response missing reserve id")
	}
	return reserve, resp.Data.ReserveCorrectionCheckInfo, nil
}

func (c *Client) DeleteReserve(ctx context.Context, token string, req DeleteReserveRequest) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	if req.ReserveID == "" {
		return errors.New("reserve id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/vc/v1/reserves/:reserve_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("reserve_id", req.ReserveID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("delete reserve failed: empty response")
	}
	resp := &deleteReserveResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("delete reserve failed: %s", resp.Msg)
	}
	return nil
}
