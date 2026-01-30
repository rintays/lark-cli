package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"lark/internal/larkapi"
)

type getMeetingResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getMeetingResponseData `json:"data"`
}

type getMeetingResponseData struct {
	Meeting *larkapi.Meeting `json:"meeting"`
}

func (r *getMeetingResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) GetMeeting(ctx context.Context, token string, req larkapi.GetMeetingRequest) (larkapi.Meeting, error) {
	if !c.available() || c.coreConfig == nil {
		return larkapi.Meeting{}, ErrUnavailable
	}
	if req.MeetingID == "" {
		return larkapi.Meeting{}, errors.New("meeting id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return larkapi.Meeting{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/vc/v1/meetings/:meeting_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("meeting_id", req.MeetingID)
	if req.WithParticipants {
		apiReq.QueryParams.Set("with_participants", "true")
	}
	if req.WithMeetingAbility {
		apiReq.QueryParams.Set("with_meeting_ability", "true")
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}
	if req.QueryMode != 0 {
		apiReq.QueryParams.Set("query_mode", fmt.Sprintf("%d", req.QueryMode))
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return larkapi.Meeting{}, err
	}
	if apiResp == nil {
		return larkapi.Meeting{}, errors.New("get meeting failed: empty response")
	}
	resp := &getMeetingResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return larkapi.Meeting{}, err
	}
	if !resp.Success() {
		return larkapi.Meeting{}, fmt.Errorf("get meeting failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Meeting == nil || resp.Data.Meeting.ID == "" {
		return larkapi.Meeting{}, errors.New("get meeting response missing meeting")
	}
	return *resp.Data.Meeting, nil
}
