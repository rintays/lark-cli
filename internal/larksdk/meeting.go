package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type getMeetingResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getMeetingResponseData `json:"data"`
}

type getMeetingResponseData struct {
	Meeting *Meeting `json:"meeting"`
}

func (r *getMeetingResponse) Success() bool {
	return r.Code == 0
}

type listMeetingsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listMeetingsResponseData `json:"data"`
}

type listMeetingsResponseData struct {
	MeetingList []*meetingListInfo `json:"meeting_list"`
	PageToken   *string            `json:"page_token"`
	HasMore     *bool              `json:"has_more"`
}

type meetingListInfo struct {
	ID        *string `json:"meeting_id"`
	Topic     *string `json:"meeting_topic"`
	Status    *int    `json:"meeting_status,omitempty"`
	StartTime *string `json:"meeting_start_time"`
	EndTime   *string `json:"meeting_end_time"`
}

func (r *listMeetingsResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) GetMeeting(ctx context.Context, token string, req GetMeetingRequest) (Meeting, error) {
	if !c.available() || c.coreConfig == nil {
		return Meeting{}, ErrUnavailable
	}
	if req.MeetingID == "" {
		return Meeting{}, errors.New("meeting id is required")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Meeting{}, errors.New("tenant access token is required")
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
		return Meeting{}, err
	}
	if apiResp == nil {
		return Meeting{}, errors.New("get meeting failed: empty response")
	}
	resp := &getMeetingResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return Meeting{}, err
	}
	if !resp.Success() {
		return Meeting{}, fmt.Errorf("get meeting failed: %s", resp.Msg)
	}
	if resp.Data == nil || resp.Data.Meeting == nil || resp.Data.Meeting.ID == "" {
		return Meeting{}, errors.New("get meeting response missing meeting")
	}
	return *resp.Data.Meeting, nil
}

func (c *Client) ListMeetings(ctx context.Context, token string, req ListMeetingsRequest) (ListMeetingsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListMeetingsResult{}, ErrUnavailable
	}
	if (req.StartTime == "") != (req.EndTime == "") {
		return ListMeetingsResult{}, errors.New("start and end must be provided together")
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return ListMeetingsResult{}, errors.New("tenant access token is required")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/vc/v1/meeting_list",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	if req.StartTime != "" {
		apiReq.QueryParams.Set("start_time", req.StartTime)
	}
	if req.EndTime != "" {
		apiReq.QueryParams.Set("end_time", req.EndTime)
	}
	if req.MeetingStatus != nil {
		apiReq.QueryParams.Set("meeting_status", fmt.Sprintf("%d", *req.MeetingStatus))
	}
	if req.MeetingNo != "" {
		apiReq.QueryParams.Set("meeting_no", req.MeetingNo)
	}
	if req.UserID != "" {
		apiReq.QueryParams.Set("user_id", req.UserID)
	}
	if req.RoomID != "" {
		apiReq.QueryParams.Set("room_id", req.RoomID)
	}
	if req.MeetingType != nil {
		apiReq.QueryParams.Set("meeting_type", fmt.Sprintf("%d", *req.MeetingType))
	}
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}
	if req.IncludeExternalMeetings != nil {
		apiReq.QueryParams.Set("include_external_meetings", fmt.Sprint(*req.IncludeExternalMeetings))
	}
	if req.IncludeWebinar != nil {
		apiReq.QueryParams.Set("include_webinar", fmt.Sprint(*req.IncludeWebinar))
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, larkcore.WithTenantAccessToken(tenantToken))
	if err != nil {
		return ListMeetingsResult{}, err
	}
	if apiResp == nil {
		return ListMeetingsResult{}, errors.New("list meetings failed: empty response")
	}
	resp := &listMeetingsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListMeetingsResult{}, err
	}
	if !resp.Success() {
		return ListMeetingsResult{}, fmt.Errorf("list meetings failed: %s", resp.Msg)
	}

	result := ListMeetingsResult{}
	if resp.Data != nil {
		if resp.Data.MeetingList != nil {
			result.Items = make([]MeetingListItem, 0, len(resp.Data.MeetingList))
			for _, meeting := range resp.Data.MeetingList {
				if meeting == nil {
					continue
				}
				item := MeetingListItem{}
				if meeting.ID != nil {
					item.ID = *meeting.ID
				}
				if meeting.Topic != nil {
					item.Topic = *meeting.Topic
				}
				if meeting.Status != nil {
					status := *meeting.Status
					item.Status = &status
				}
				if meeting.StartTime != nil {
					item.StartTime = *meeting.StartTime
				}
				if meeting.EndTime != nil {
					item.EndTime = *meeting.EndTime
				}
				result.Items = append(result.Items, item)
			}
		}
		if resp.Data.PageToken != nil {
			result.PageToken = *resp.Data.PageToken
		}
		if resp.Data.HasMore != nil {
			result.HasMore = *resp.Data.HasMore
		}
	}
	return result, nil
}
