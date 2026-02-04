package larksdk

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"
)

type primaryCalendarResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *primaryCalendarResponseData `json:"data"`
}

type primaryCalendarEntry struct {
	Calendar Calendar `json:"calendar"`
}

type primaryCalendarResponseData struct {
	CalendarID string                 `json:"calendar_id"`
	Calendar   Calendar               `json:"calendar"`
	Calendars  []primaryCalendarEntry `json:"calendars"`
}

func (r *primaryCalendarResponse) Success() bool {
	return r.Code == 0
}

type listCalendarEventsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *listCalendarEventsResponseData `json:"data"`
}

type listCalendarEventsResponseData struct {
	Items     []CalendarEvent `json:"items"`
	Events    []CalendarEvent `json:"events"`
	PageToken string          `json:"page_token"`
	HasMore   bool            `json:"has_more"`
	SyncToken string          `json:"sync_token"`
}

func (r *listCalendarEventsResponse) Success() bool {
	return r.Code == 0
}

type searchCalendarEventsResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *searchCalendarEventsResponseData `json:"data"`
}

type searchCalendarEventsResponseData struct {
	Items     []CalendarEvent `json:"items"`
	PageToken string          `json:"page_token"`
}

func (r *searchCalendarEventsResponse) Success() bool {
	return r.Code == 0
}

type getCalendarEventResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *getCalendarEventResponseData `json:"data"`
}

type getCalendarEventResponseData struct {
	Event CalendarEvent `json:"event"`
}

func (r *getCalendarEventResponse) Success() bool {
	return r.Code == 0
}

type updateCalendarEventResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *updateCalendarEventResponseData `json:"data"`
}

type updateCalendarEventResponseData struct {
	Event CalendarEvent `json:"event"`
}

func (r *updateCalendarEventResponse) Success() bool {
	return r.Code == 0
}

type deleteCalendarEventResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *deleteCalendarEventResponseData `json:"data"`
}

type deleteCalendarEventResponseData struct {
	EventID string `json:"event_id"`
	Deleted *bool  `json:"deleted"`
}

func (r *deleteCalendarEventResponse) Success() bool {
	return r.Code == 0
}

type createCalendarEventResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
	Data *createCalendarEventResponseData `json:"data"`
}

type createCalendarEventResponseData struct {
	Event   CalendarEvent `json:"event"`
	EventID string        `json:"event_id"`
}

func (r *createCalendarEventResponse) Success() bool {
	return r.Code == 0
}

type createCalendarEventAttendeesResponse struct {
	*larkcore.ApiResp `json:"-"`
	larkcore.CodeError
}

func (r *createCalendarEventAttendeesResponse) Success() bool {
	return r.Code == 0
}

func (c *Client) PrimaryCalendar(ctx context.Context, token string) (Calendar, error) {
	if !c.available() || c.coreConfig == nil {
		return Calendar{}, ErrUnavailable
	}
	tenantToken := c.tenantToken(token)
	if tenantToken == "" {
		return Calendar{}, errors.New("tenant access token is required")
	}
	return c.primaryCalendar(ctx, larkcore.WithTenantAccessToken(tenantToken))
}

func (c *Client) PrimaryCalendarWithUserToken(ctx context.Context, userAccessToken string) (Calendar, error) {
	if !c.available() || c.coreConfig == nil {
		return Calendar{}, ErrUnavailable
	}
	userAccessToken = strings.TrimSpace(userAccessToken)
	if userAccessToken == "" {
		return Calendar{}, errors.New("user access token is required")
	}
	return c.primaryCalendar(ctx, larkcore.WithUserAccessToken(userAccessToken))
}

func (c *Client) primaryCalendar(ctx context.Context, option larkcore.RequestOptionFunc) (Calendar, error) {
	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/primary",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return Calendar{}, err
	}
	if apiResp == nil {
		return Calendar{}, errors.New("primary calendar failed: empty response")
	}
	resp := &primaryCalendarResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return Calendar{}, err
	}
	if !resp.Success() {
		return Calendar{}, fmt.Errorf("primary calendar failed: %s", resp.Msg)
	}
	if resp.Data == nil {
		return Calendar{}, errors.New("primary calendar response missing data")
	}
	calendar := resp.Data.Calendar
	if calendar.CalendarID == "" {
		calendar.CalendarID = resp.Data.CalendarID
	}
	if calendar.CalendarID == "" {
		for _, entry := range resp.Data.Calendars {
			if entry.Calendar.CalendarID != "" {
				calendar = entry.Calendar
				break
			}
		}
	}
	if calendar.CalendarID == "" {
		return Calendar{}, errors.New("primary calendar response missing calendar_id")
	}
	return calendar, nil
}

func (c *Client) ListCalendarEvents(ctx context.Context, token string, tokenType AccessTokenType, req ListCalendarEventsRequest) (ListCalendarEventsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return ListCalendarEventsResult{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return ListCalendarEventsResult{}, errors.New("calendar id is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	if req.StartTime != "" {
		apiReq.QueryParams.Set("start_time", req.StartTime)
	}
	if req.EndTime != "" {
		apiReq.QueryParams.Set("end_time", req.EndTime)
	}
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}
	if req.SyncToken != "" {
		apiReq.QueryParams.Set("sync_token", req.SyncToken)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}
	if apiResp == nil {
		return ListCalendarEventsResult{}, errors.New("list calendar events failed: empty response")
	}
	resp := &listCalendarEventsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return ListCalendarEventsResult{}, err
	}
	if !resp.Success() {
		return ListCalendarEventsResult{}, fmt.Errorf("list calendar events failed: %s (code=%d)", resp.Msg, resp.Code)
	}

	result := ListCalendarEventsResult{}
	if resp.Data != nil {
		items := resp.Data.Items
		if len(items) == 0 {
			items = resp.Data.Events
		}
		result.Items = items
		result.PageToken = resp.Data.PageToken
		result.HasMore = resp.Data.HasMore
		result.SyncToken = resp.Data.SyncToken
	}
	return result, nil
}

func (c *Client) SearchCalendarEvents(ctx context.Context, token string, tokenType AccessTokenType, req SearchCalendarEventsRequest) (SearchCalendarEventsResult, error) {
	if !c.available() || c.coreConfig == nil {
		return SearchCalendarEventsResult{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return SearchCalendarEventsResult{}, errors.New("calendar id is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return SearchCalendarEventsResult{}, err
	}

	payload := map[string]any{
		"query": req.Query,
	}
	filter := map[string]any{}
	if req.StartTime != "" {
		filter["start_time"] = map[string]string{
			"timestamp": req.StartTime,
		}
	}
	if req.EndTime != "" {
		filter["end_time"] = map[string]string{
			"timestamp": req.EndTime,
		}
	}
	if len(req.UserIDs) > 0 {
		filter["user_ids"] = req.UserIDs
	}
	if len(req.RoomIDs) > 0 {
		filter["room_ids"] = req.RoomIDs
	}
	if len(req.ChatIDs) > 0 {
		filter["chat_ids"] = req.ChatIDs
	}
	if len(filter) > 0 {
		payload["filter"] = filter
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events/search",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	if req.PageSize > 0 {
		apiReq.QueryParams.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		apiReq.QueryParams.Set("page_token", req.PageToken)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return SearchCalendarEventsResult{}, err
	}
	if apiResp == nil {
		return SearchCalendarEventsResult{}, errors.New("search calendar events failed: empty response")
	}
	resp := &searchCalendarEventsResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return SearchCalendarEventsResult{}, err
	}
	if !resp.Success() {
		return SearchCalendarEventsResult{}, fmt.Errorf("search calendar events failed: %s (code=%d)", resp.Msg, resp.Code)
	}

	result := SearchCalendarEventsResult{}
	if resp.Data != nil {
		result.Items = resp.Data.Items
		result.PageToken = resp.Data.PageToken
	}
	return result, nil
}

func (c *Client) GetCalendarEvent(ctx context.Context, token string, tokenType AccessTokenType, req GetCalendarEventRequest) (CalendarEvent, error) {
	if !c.available() || c.coreConfig == nil {
		return CalendarEvent{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return CalendarEvent{}, errors.New("calendar id is required")
	}
	if req.EventID == "" {
		return CalendarEvent{}, errors.New("event id is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return CalendarEvent{}, err
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events/:event_id",
		HttpMethod:                http.MethodGet,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	apiReq.PathParams.Set("event_id", req.EventID)
	if req.NeedMeetingSettings != nil {
		apiReq.QueryParams.Set("need_meeting_settings", strconv.FormatBool(*req.NeedMeetingSettings))
	}
	if req.NeedAttendee != nil {
		apiReq.QueryParams.Set("need_attendee", strconv.FormatBool(*req.NeedAttendee))
	}
	if req.MaxAttendeeNum != nil && *req.MaxAttendeeNum > 0 {
		apiReq.QueryParams.Set("max_attendee_num", fmt.Sprintf("%d", *req.MaxAttendeeNum))
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return CalendarEvent{}, err
	}
	if apiResp == nil {
		return CalendarEvent{}, errors.New("get calendar event failed: empty response")
	}
	resp := &getCalendarEventResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return CalendarEvent{}, err
	}
	if !resp.Success() {
		return CalendarEvent{}, fmt.Errorf("get calendar event failed: %s (code=%d)", resp.Msg, resp.Code)
	}
	if resp.Data == nil {
		return CalendarEvent{}, errors.New("get calendar event response missing data")
	}
	result := resp.Data.Event
	if result.EventID == "" {
		result.EventID = req.EventID
	}
	return result, nil
}

func (c *Client) UpdateCalendarEvent(ctx context.Context, token string, tokenType AccessTokenType, req UpdateCalendarEventRequest) (CalendarEvent, error) {
	if !c.available() || c.coreConfig == nil {
		return CalendarEvent{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return CalendarEvent{}, errors.New("calendar id is required")
	}
	if req.EventID == "" {
		return CalendarEvent{}, errors.New("event id is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return CalendarEvent{}, err
	}

	payload := map[string]any{}
	if req.Summary != "" {
		payload["summary"] = req.Summary
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.Start != nil || req.End != nil {
		if req.Start == nil || req.End == nil {
			return CalendarEvent{}, errors.New("start_time and end_time must be provided together")
		}
		payload["start_time"] = req.Start
		payload["end_time"] = req.End
	} else if req.StartTime != nil || req.EndTime != nil {
		if req.StartTime == nil || req.EndTime == nil {
			return CalendarEvent{}, errors.New("start and end timestamps must be provided together")
		}
		payload["start_time"] = map[string]string{
			"timestamp": fmt.Sprintf("%d", *req.StartTime),
		}
		payload["end_time"] = map[string]string{
			"timestamp": fmt.Sprintf("%d", *req.EndTime),
		}
	}
	if req.NeedNotification != nil {
		payload["need_notification"] = *req.NeedNotification
	}
	if req.Visibility != "" {
		payload["visibility"] = req.Visibility
	}
	if req.AttendeeAbility != "" {
		payload["attendee_ability"] = req.AttendeeAbility
	}
	if req.FreeBusyStatus != "" {
		payload["free_busy_status"] = req.FreeBusyStatus
	}
	if req.Location != nil {
		payload["location"] = req.Location
	}
	if req.Color != nil {
		payload["color"] = *req.Color
	}
	if req.Reminders != nil {
		payload["reminders"] = req.Reminders
	}
	if req.Recurrence != "" {
		payload["recurrence"] = req.Recurrence
	}
	if req.VChat != nil {
		payload["vchat"] = req.VChat
	}
	if req.Schemas != nil {
		payload["schemas"] = req.Schemas
	}
	if req.Attachments != nil {
		payload["attachments"] = req.Attachments
	}
	if req.EventCheckIn != nil {
		payload["event_check_in"] = req.EventCheckIn
	}
	if req.Extra != nil {
		startExtra := req.Extra["start_time"] != nil
		endExtra := req.Extra["end_time"] != nil
		if req.Start == nil && req.End == nil && req.StartTime == nil && req.EndTime == nil && startExtra != endExtra {
			return CalendarEvent{}, errors.New("start_time and end_time must be provided together")
		}
	}
	if req.Extra != nil {
		for key, value := range req.Extra {
			if _, exists := payload[key]; exists {
				continue
			}
			payload[key] = value
		}
	}
	if len(payload) == 0 {
		return CalendarEvent{}, errors.New("update calendar event requires at least one field")
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events/:event_id",
		HttpMethod:                http.MethodPatch,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	apiReq.PathParams.Set("event_id", req.EventID)
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return CalendarEvent{}, err
	}
	if apiResp == nil {
		return CalendarEvent{}, errors.New("update calendar event failed: empty response")
	}
	resp := &updateCalendarEventResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return CalendarEvent{}, err
	}
	if !resp.Success() {
		return CalendarEvent{}, fmt.Errorf("update calendar event failed: %s (code=%d)", resp.Msg, resp.Code)
	}
	if resp.Data == nil {
		return CalendarEvent{}, errors.New("update calendar event response missing data")
	}
	result := resp.Data.Event
	if result.EventID == "" {
		result.EventID = req.EventID
	}
	return result, nil
}

func (c *Client) DeleteCalendarEvent(ctx context.Context, token string, tokenType AccessTokenType, req DeleteCalendarEventRequest) (DeleteCalendarEventResult, error) {
	if !c.available() || c.coreConfig == nil {
		return DeleteCalendarEventResult{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return DeleteCalendarEventResult{}, errors.New("calendar id is required")
	}
	if req.EventID == "" {
		return DeleteCalendarEventResult{}, errors.New("event id is required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return DeleteCalendarEventResult{}, err
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events/:event_id",
		HttpMethod:                http.MethodDelete,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	apiReq.PathParams.Set("event_id", req.EventID)
	apiReq.QueryParams.Set("need_notification", fmt.Sprintf("%t", req.NeedNotification))

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return DeleteCalendarEventResult{}, err
	}
	if apiResp == nil {
		return DeleteCalendarEventResult{}, errors.New("delete calendar event failed: empty response")
	}
	resp := &deleteCalendarEventResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return DeleteCalendarEventResult{}, err
	}
	if !resp.Success() {
		return DeleteCalendarEventResult{}, fmt.Errorf("delete calendar event failed: %s (code=%d)", resp.Msg, resp.Code)
	}
	result := DeleteCalendarEventResult{EventID: req.EventID, Deleted: true}
	if resp.Data != nil {
		if resp.Data.EventID != "" {
			result.EventID = resp.Data.EventID
		}
		if resp.Data.Deleted != nil {
			result.Deleted = *resp.Data.Deleted
		}
	}
	return result, nil
}

func (c *Client) CreateCalendarEvent(ctx context.Context, token string, tokenType AccessTokenType, req CreateCalendarEventRequest) (CalendarEvent, error) {
	if !c.available() || c.coreConfig == nil {
		return CalendarEvent{}, ErrUnavailable
	}
	if req.CalendarID == "" {
		return CalendarEvent{}, errors.New("calendar id is required")
	}
	if req.Summary == "" && (req.Extra == nil || req.Extra["summary"] == nil) {
		return CalendarEvent{}, errors.New("summary is required")
	}
	if req.StartTime == 0 && req.EndTime == 0 && req.Start == nil && req.End == nil && (req.Extra == nil || (req.Extra["start_time"] == nil && req.Extra["end_time"] == nil)) {
		return CalendarEvent{}, errors.New("start and end times are required")
	}
	if req.StartTime == 0 && req.EndTime == 0 && req.Start == nil && req.End == nil && req.Extra != nil {
		startExtra := req.Extra["start_time"] != nil
		endExtra := req.Extra["end_time"] != nil
		if startExtra != endExtra {
			return CalendarEvent{}, errors.New("start_time and end_time must be provided together")
		}
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return CalendarEvent{}, err
	}

	payload := map[string]any{}
	if req.Summary != "" {
		payload["summary"] = req.Summary
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	if req.Start != nil || req.End != nil {
		if req.Start == nil || req.End == nil {
			return CalendarEvent{}, errors.New("start_time and end_time must be provided together")
		}
		payload["start_time"] = req.Start
		payload["end_time"] = req.End
	} else if req.StartTime != 0 || req.EndTime != 0 {
		if req.StartTime == 0 || req.EndTime == 0 {
			return CalendarEvent{}, errors.New("start and end times are required")
		}
		payload["start_time"] = map[string]string{
			"timestamp": fmt.Sprintf("%d", req.StartTime),
		}
		payload["end_time"] = map[string]string{
			"timestamp": fmt.Sprintf("%d", req.EndTime),
		}
	}
	if req.NeedNotification != nil {
		payload["need_notification"] = *req.NeedNotification
	}
	if req.Visibility != "" {
		payload["visibility"] = req.Visibility
	}
	if req.AttendeeAbility != "" {
		payload["attendee_ability"] = req.AttendeeAbility
	}
	if req.FreeBusyStatus != "" {
		payload["free_busy_status"] = req.FreeBusyStatus
	}
	if req.Location != nil {
		payload["location"] = req.Location
	}
	if req.Color != nil {
		payload["color"] = *req.Color
	}
	if req.Reminders != nil {
		payload["reminders"] = req.Reminders
	}
	if req.Recurrence != "" {
		payload["recurrence"] = req.Recurrence
	}
	if req.VChat != nil {
		payload["vchat"] = req.VChat
	}
	if req.Schemas != nil {
		payload["schemas"] = req.Schemas
	}
	if req.Attachments != nil {
		payload["attachments"] = req.Attachments
	}
	if req.EventCheckIn != nil {
		payload["event_check_in"] = req.EventCheckIn
	}
	if req.Extra != nil {
		for key, value := range req.Extra {
			if _, exists := payload[key]; exists {
				continue
			}
			payload[key] = value
		}
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      payload,
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	if req.IdempotencyKey != "" {
		apiReq.QueryParams.Set("idempotency_key", req.IdempotencyKey)
	}
	if req.UserIDType != "" {
		apiReq.QueryParams.Set("user_id_type", req.UserIDType)
	}

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return CalendarEvent{}, err
	}
	if apiResp == nil {
		return CalendarEvent{}, errors.New("create calendar event failed: empty response")
	}
	resp := &createCalendarEventResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return CalendarEvent{}, err
	}
	if !resp.Success() {
		return CalendarEvent{}, fmt.Errorf("create calendar event failed: %s (code=%d)", resp.Msg, resp.Code)
	}
	if resp.Data == nil {
		return CalendarEvent{}, errors.New("create calendar event response missing data")
	}
	result := resp.Data.Event
	if result.EventID == "" {
		result.EventID = resp.Data.EventID
	}
	if result.EventID == "" {
		return CalendarEvent{}, errors.New("create calendar event response missing event_id")
	}
	return result, nil
}

func (c *Client) CreateCalendarEventAttendees(ctx context.Context, token string, tokenType AccessTokenType, req CreateCalendarEventAttendeesRequest) error {
	if !c.available() || c.coreConfig == nil {
		return ErrUnavailable
	}
	if req.CalendarID == "" {
		return errors.New("calendar id is required")
	}
	if req.EventID == "" {
		return errors.New("event id is required")
	}
	if len(req.Attendees) == 0 {
		return errors.New("attendees are required")
	}
	option, _, err := c.accessTokenOption(token, tokenType)
	if err != nil {
		return err
	}

	apiReq := &larkcore.ApiReq{
		ApiPath:                   "/open-apis/calendar/v4/calendars/:calendar_id/events/:event_id/attendees",
		HttpMethod:                http.MethodPost,
		PathParams:                larkcore.PathParams{},
		QueryParams:               larkcore.QueryParams{},
		Body:                      map[string]any{"attendees": req.Attendees},
		SupportedAccessTokenTypes: []larkcore.AccessTokenType{larkcore.AccessTokenTypeTenant, larkcore.AccessTokenTypeUser},
	}
	apiReq.PathParams.Set("calendar_id", req.CalendarID)
	apiReq.PathParams.Set("event_id", req.EventID)

	apiResp, err := larkcore.Request(ctx, apiReq, c.coreConfig, option)
	if err != nil {
		return err
	}
	if apiResp == nil {
		return errors.New("create calendar event attendees failed: empty response")
	}
	resp := &createCalendarEventAttendeesResponse{ApiResp: apiResp}
	if err := apiResp.JSONUnmarshalBody(resp, c.coreConfig); err != nil {
		return err
	}
	if !resp.Success() {
		return fmt.Errorf("create calendar event attendees failed: %s (code=%d)", resp.Msg, resp.Code)
	}
	return nil
}
