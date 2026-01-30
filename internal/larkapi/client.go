package larkapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/url"
)

type Client struct {
	BaseURL    string
	HTTPClient *http.Client
	AppID      string
	AppSecret  string
}

func (c *Client) httpClient() *http.Client {
	if c.HTTPClient != nil {
		return c.HTTPClient
	}
	return http.DefaultClient
}

func (c *Client) endpoint(path string, query url.Values) (string, error) {
	base, err := url.Parse(c.BaseURL)
	if err != nil {
		return "", err
	}
	base.Path = path
	if len(query) > 0 {
		base.RawQuery = query.Encode()
	}
	return base.String(), nil
}

type apiResponse struct {
	Code int    `json:"code"`
	Msg  string `json:"msg"`
}

type tenantTokenResponse struct {
	apiResponse
	TenantAccessToken string `json:"tenant_access_token"`
	Expire            int64  `json:"expire"`
}

func (c *Client) TenantAccessToken(ctx context.Context) (string, int64, error) {
	payload := map[string]string{
		"app_id":     c.AppID,
		"app_secret": c.AppSecret,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", 0, err
	}
	endpoint, err := c.endpoint("/open-apis/auth/v3/tenant_access_token/internal/", nil)
	if err != nil {
		return "", 0, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", 0, err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return "", 0, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", 0, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", 0, fmt.Errorf("token request failed: %s", resp.Status)
	}
	var parsed tenantTokenResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", 0, err
	}
	if parsed.Code != 0 {
		return "", 0, fmt.Errorf("token request failed: %s", parsed.Msg)
	}
	if parsed.TenantAccessToken == "" {
		return "", 0, fmt.Errorf("token response missing tenant_access_token")
	}
	return parsed.TenantAccessToken, parsed.Expire, nil
}

type TenantInfo struct {
	TenantKey string `json:"tenant_key"`
	Name      string `json:"name"`
}

type whoamiResponse struct {
	apiResponse
	Data struct {
		Tenant TenantInfo `json:"tenant"`
	} `json:"data"`
}

func (c *Client) WhoAmI(ctx context.Context, token string) (TenantInfo, error) {
	endpoint, err := c.endpoint("/open-apis/tenant/v2/tenant/query", nil)
	if err != nil {
		return TenantInfo{}, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return TenantInfo{}, err
	}
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(req)
	if err != nil {
		return TenantInfo{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return TenantInfo{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return TenantInfo{}, fmt.Errorf("whoami request failed: %s", resp.Status)
	}
	var parsed whoamiResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return TenantInfo{}, err
	}
	if parsed.Code != 0 {
		return TenantInfo{}, fmt.Errorf("whoami request failed: %s", parsed.Msg)
	}
	return parsed.Data.Tenant, nil
}

type Calendar struct {
	CalendarID string `json:"calendar_id"`
	Summary    string `json:"summary"`
}

type calendarPrimaryResponse struct {
	apiResponse
	Data struct {
		CalendarID string   `json:"calendar_id"`
		Calendar   Calendar `json:"calendar"`
	} `json:"data"`
}

func (c *Client) PrimaryCalendar(ctx context.Context, token string) (Calendar, error) {
	endpoint, err := c.endpoint("/open-apis/calendar/v4/calendars/primary", nil)
	if err != nil {
		return Calendar{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, nil)
	if err != nil {
		return Calendar{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return Calendar{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Calendar{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Calendar{}, fmt.Errorf("primary calendar request failed: %s", resp.Status)
	}
	var parsed calendarPrimaryResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Calendar{}, err
	}
	if parsed.Code != 0 {
		return Calendar{}, fmt.Errorf("primary calendar request failed: %s", parsed.Msg)
	}
	if parsed.Data.Calendar.CalendarID == "" && parsed.Data.CalendarID == "" {
		return Calendar{}, fmt.Errorf("primary calendar response missing calendar_id")
	}
	if parsed.Data.Calendar.CalendarID == "" {
		parsed.Data.Calendar.CalendarID = parsed.Data.CalendarID
	}
	return parsed.Data.Calendar, nil
}

type CalendarEventTime struct {
	Date      string `json:"date"`
	Timestamp string `json:"timestamp"`
	Timezone  string `json:"timezone"`
}

type CalendarEvent struct {
	EventID     string            `json:"event_id"`
	Summary     string            `json:"summary"`
	Description string            `json:"description"`
	StartTime   CalendarEventTime `json:"start_time"`
	EndTime     CalendarEventTime `json:"end_time"`
}

type ListCalendarEventsRequest struct {
	CalendarID string
	StartTime  string
	EndTime    string
	PageSize   int
	PageToken  string
	SyncToken  string
}

type listCalendarEventsResponse struct {
	apiResponse
	Data struct {
		Items     []CalendarEvent `json:"items"`
		Events    []CalendarEvent `json:"events"`
		PageToken string          `json:"page_token"`
		HasMore   bool            `json:"has_more"`
		SyncToken string          `json:"sync_token"`
	} `json:"data"`
}

type ListCalendarEventsResult struct {
	Items     []CalendarEvent
	PageToken string
	HasMore   bool
	SyncToken string
}

func (c *Client) ListCalendarEvents(ctx context.Context, token string, req ListCalendarEventsRequest) (ListCalendarEventsResult, error) {
	if req.CalendarID == "" {
		return ListCalendarEventsResult{}, fmt.Errorf("calendar id is required")
	}
	query := url.Values{}
	if req.StartTime != "" {
		query.Set("start_time", req.StartTime)
	}
	if req.EndTime != "" {
		query.Set("end_time", req.EndTime)
	}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.SyncToken != "" {
		query.Set("sync_token", req.SyncToken)
	}
	endpoint, err := c.endpoint("/open-apis/calendar/v4/calendars/"+url.PathEscape(req.CalendarID)+"/events", query)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListCalendarEventsResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListCalendarEventsResult{}, fmt.Errorf("list calendar events failed: %s", resp.Status)
	}
	var parsed listCalendarEventsResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListCalendarEventsResult{}, err
	}
	if parsed.Code != 0 {
		return ListCalendarEventsResult{}, fmt.Errorf("list calendar events failed: %s", parsed.Msg)
	}
	items := parsed.Data.Items
	if len(items) == 0 {
		items = parsed.Data.Events
	}
	return ListCalendarEventsResult{
		Items:     items,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
		SyncToken: parsed.Data.SyncToken,
	}, nil
}

type CreateCalendarEventRequest struct {
	CalendarID  string
	Summary     string
	Description string
	StartTime   int64
	EndTime     int64
}

type createCalendarEventResponse struct {
	apiResponse
	Data struct {
		Event   CalendarEvent `json:"event"`
		EventID string        `json:"event_id"`
	} `json:"data"`
}

func (c *Client) CreateCalendarEvent(ctx context.Context, token string, req CreateCalendarEventRequest) (CalendarEvent, error) {
	if req.CalendarID == "" {
		return CalendarEvent{}, fmt.Errorf("calendar id is required")
	}
	if req.Summary == "" {
		return CalendarEvent{}, fmt.Errorf("summary is required")
	}
	if req.StartTime == 0 || req.EndTime == 0 {
		return CalendarEvent{}, fmt.Errorf("start and end times are required")
	}
	payload := map[string]any{
		"summary": req.Summary,
		"start_time": map[string]string{
			"timestamp": fmt.Sprintf("%d", req.StartTime),
		},
		"end_time": map[string]string{
			"timestamp": fmt.Sprintf("%d", req.EndTime),
		},
	}
	if req.Description != "" {
		payload["description"] = req.Description
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return CalendarEvent{}, err
	}
	endpoint, err := c.endpoint("/open-apis/calendar/v4/calendars/"+url.PathEscape(req.CalendarID)+"/events", nil)
	if err != nil {
		return CalendarEvent{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return CalendarEvent{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return CalendarEvent{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return CalendarEvent{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return CalendarEvent{}, fmt.Errorf("create calendar event failed: %s", resp.Status)
	}
	var parsed createCalendarEventResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return CalendarEvent{}, err
	}
	if parsed.Code != 0 {
		return CalendarEvent{}, fmt.Errorf("create calendar event failed: %s", parsed.Msg)
	}
	event := parsed.Data.Event
	if event.EventID == "" {
		event.EventID = parsed.Data.EventID
	}
	if event.EventID == "" {
		return CalendarEvent{}, fmt.Errorf("create calendar event response missing event_id")
	}
	return event, nil
}

type CalendarEventAttendee struct {
	Type            string `json:"type,omitempty"`
	UserID          string `json:"user_id,omitempty"`
	ChatID          string `json:"chat_id,omitempty"`
	RoomID          string `json:"room_id,omitempty"`
	ThirdPartyEmail string `json:"third_party_email,omitempty"`
}

type CreateCalendarEventAttendeesRequest struct {
	CalendarID string
	EventID    string
	Attendees  []CalendarEventAttendee
}

type createCalendarEventAttendeesResponse struct {
	apiResponse
}

func (c *Client) CreateCalendarEventAttendees(ctx context.Context, token string, req CreateCalendarEventAttendeesRequest) error {
	if req.CalendarID == "" {
		return fmt.Errorf("calendar id is required")
	}
	if req.EventID == "" {
		return fmt.Errorf("event id is required")
	}
	if len(req.Attendees) == 0 {
		return fmt.Errorf("attendees are required")
	}
	payload := map[string]any{
		"attendees": req.Attendees,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	endpoint, err := c.endpoint("/open-apis/calendar/v4/calendars/"+url.PathEscape(req.CalendarID)+"/events/"+url.PathEscape(req.EventID)+"/attendees", nil)
	if err != nil {
		return err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("create calendar event attendees failed: %s", resp.Status)
	}
	var parsed createCalendarEventAttendeesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return err
	}
	if parsed.Code != 0 {
		return fmt.Errorf("create calendar event attendees failed: %s", parsed.Msg)
	}
	return nil
}

type MeetingUser struct {
	ID       string `json:"id"`
	UserType int    `json:"user_type"`
}

type MeetingParticipant struct {
	ID                string `json:"id"`
	FirstJoinTime     string `json:"first_join_time"`
	FinalLeaveTime    string `json:"final_leave_time"`
	InMeetingDuration string `json:"in_meeting_duration"`
	UserType          int    `json:"user_type"`
	IsHost            bool   `json:"is_host"`
	IsCohost          bool   `json:"is_cohost"`
	IsExternal        bool   `json:"is_external"`
	Status            int    `json:"status"`
}

type MeetingAbility struct {
	UseVideo        bool `json:"use_video"`
	UseAudio        bool `json:"use_audio"`
	UseShareScreen  bool `json:"use_share_screen"`
	UseFollowScreen bool `json:"use_follow_screen"`
	UseRecording    bool `json:"use_recording"`
	UsePSTN         bool `json:"use_pstn"`
}

type Meeting struct {
	ID                          string               `json:"id"`
	Topic                       string               `json:"topic"`
	URL                         string               `json:"url"`
	MeetingNo                   string               `json:"meeting_no"`
	Password                    string               `json:"password"`
	CreateTime                  string               `json:"create_time"`
	StartTime                   string               `json:"start_time"`
	EndTime                     string               `json:"end_time"`
	HostUser                    MeetingUser          `json:"host_user"`
	Status                      int                  `json:"status"`
	ParticipantCount            string               `json:"participant_count"`
	ParticipantCountAccumulated string               `json:"participant_count_accumulated"`
	Participants                []MeetingParticipant `json:"participants,omitempty"`
	Ability                     MeetingAbility       `json:"ability,omitempty"`
}

type GetMeetingRequest struct {
	MeetingID          string
	WithParticipants   bool
	WithMeetingAbility bool
	UserIDType         string
	QueryMode          int
}

type getMeetingResponse struct {
	apiResponse
	Data struct {
		Meeting Meeting `json:"meeting"`
	} `json:"data"`
}

func (c *Client) GetMeeting(ctx context.Context, token string, req GetMeetingRequest) (Meeting, error) {
	if req.MeetingID == "" {
		return Meeting{}, fmt.Errorf("meeting id is required")
	}
	query := url.Values{}
	if req.WithParticipants {
		query.Set("with_participants", "true")
	}
	if req.WithMeetingAbility {
		query.Set("with_meeting_ability", "true")
	}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	if req.QueryMode != 0 {
		query.Set("query_mode", fmt.Sprintf("%d", req.QueryMode))
	}
	endpoint, err := c.endpoint("/open-apis/vc/v1/meetings/"+url.PathEscape(req.MeetingID), query)
	if err != nil {
		return Meeting{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return Meeting{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return Meeting{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return Meeting{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return Meeting{}, fmt.Errorf("get meeting failed: %s", resp.Status)
	}
	var parsed getMeetingResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return Meeting{}, err
	}
	if parsed.Code != 0 {
		return Meeting{}, fmt.Errorf("get meeting failed: %s", parsed.Msg)
	}
	if parsed.Data.Meeting.ID == "" {
		return Meeting{}, fmt.Errorf("get meeting response missing meeting")
	}
	return parsed.Data.Meeting, nil
}

type MessageRequest struct {
	ReceiveID     string
	ReceiveIDType string
	Text          string
}

type sendMessageResponse struct {
	apiResponse
	Data struct {
		MessageID string `json:"message_id"`
	} `json:"data"`
}

func (c *Client) SendMessage(ctx context.Context, token string, req MessageRequest) (string, error) {
	content, err := json.Marshal(map[string]string{"text": req.Text})
	if err != nil {
		return "", err
	}
	payload := map[string]string{
		"receive_id": req.ReceiveID,
		"msg_type":   "text",
		"content":    string(content),
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	receiveIDType := req.ReceiveIDType
	if receiveIDType == "" {
		receiveIDType = "chat_id"
	}
	query := url.Values{"receive_id_type": []string{receiveIDType}}
	endpoint, err := c.endpoint("/open-apis/im/v1/messages", query)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("send message failed: %s", resp.Status)
	}
	var parsed sendMessageResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("send message failed: %s", parsed.Msg)
	}
	return parsed.Data.MessageID, nil
}

type Chat struct {
	ChatID      string `json:"chat_id"`
	Avatar      string `json:"avatar"`
	Name        string `json:"name"`
	Description string `json:"description"`
	OwnerID     string `json:"owner_id"`
	OwnerIDType string `json:"owner_id_type"`
	External    bool   `json:"external"`
	TenantKey   string `json:"tenant_key"`
}

type ListChatsRequest struct {
	PageSize   int
	PageToken  string
	UserIDType string
}

type listChatsResponse struct {
	apiResponse
	Data struct {
		Items     []Chat `json:"items"`
		PageToken string `json:"page_token"`
		HasMore   bool   `json:"has_more"`
	} `json:"data"`
}

type ListChatsResult struct {
	Items     []Chat
	PageToken string
	HasMore   bool
}

func (c *Client) ListChats(ctx context.Context, token string, req ListChatsRequest) (ListChatsResult, error) {
	query := url.Values{}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	endpoint, err := c.endpoint("/open-apis/im/v1/chats", query)
	if err != nil {
		return ListChatsResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListChatsResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListChatsResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListChatsResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListChatsResult{}, fmt.Errorf("list chats failed: %s", resp.Status)
	}
	var parsed listChatsResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListChatsResult{}, err
	}
	if parsed.Code != 0 {
		return ListChatsResult{}, fmt.Errorf("list chats failed: %s", parsed.Msg)
	}
	return ListChatsResult{
		Items:     parsed.Data.Items,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type User struct {
	UserID string `json:"user_id"`
	OpenID string `json:"open_id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Mobile string `json:"mobile"`
}

type GetContactUserRequest struct {
	UserID     string
	UserIDType string
}

type getContactUserResponse struct {
	apiResponse
	Data struct {
		User User `json:"user"`
	} `json:"data"`
}

func (c *Client) GetContactUser(ctx context.Context, token string, req GetContactUserRequest) (User, error) {
	if req.UserID == "" {
		return User{}, fmt.Errorf("user id is required")
	}
	query := url.Values{}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	endpoint, err := c.endpoint("/open-apis/contact/v3/users/"+url.PathEscape(req.UserID), query)
	if err != nil {
		return User{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return User{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return User{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return User{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return User{}, fmt.Errorf("get contact user failed: %s", resp.Status)
	}
	var parsed getContactUserResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return User{}, err
	}
	if parsed.Code != 0 {
		return User{}, fmt.Errorf("get contact user failed: %s", parsed.Msg)
	}
	return parsed.Data.User, nil
}

type BatchGetUserIDRequest struct {
	Emails  []string
	Mobiles []string
}

type batchGetUserIDResponse struct {
	apiResponse
	Data struct {
		UserList []User `json:"user_list"`
	} `json:"data"`
}

func (c *Client) BatchGetUserIDs(ctx context.Context, token string, req BatchGetUserIDRequest) ([]User, error) {
	payload := map[string]any{}
	if len(req.Emails) > 0 {
		payload["emails"] = req.Emails
	}
	if len(req.Mobiles) > 0 {
		payload["mobiles"] = req.Mobiles
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	endpoint, err := c.endpoint("/open-apis/contact/v3/users/batch_get_id", nil)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("batch get user ids failed: %s", resp.Status)
	}
	var parsed batchGetUserIDResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, err
	}
	if parsed.Code != 0 {
		return nil, fmt.Errorf("batch get user ids failed: %s", parsed.Msg)
	}
	return parsed.Data.UserList, nil
}

type ListUsersByDepartmentRequest struct {
	DepartmentID string
	PageSize     int
	PageToken    string
	UserIDType   string
}

type listUsersByDepartmentResponse struct {
	apiResponse
	Data struct {
		Items     []User `json:"items"`
		PageToken string `json:"page_token"`
		HasMore   bool   `json:"has_more"`
	} `json:"data"`
}

type ListUsersByDepartmentResult struct {
	Items     []User
	PageToken string
	HasMore   bool
}

func (c *Client) ListUsersByDepartment(ctx context.Context, token string, req ListUsersByDepartmentRequest) (ListUsersByDepartmentResult, error) {
	query := url.Values{}
	if req.DepartmentID != "" {
		query.Set("department_id", req.DepartmentID)
	}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	if req.UserIDType != "" {
		query.Set("user_id_type", req.UserIDType)
	}
	endpoint, err := c.endpoint("/open-apis/contact/v3/users/find_by_department", query)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListUsersByDepartmentResult{}, fmt.Errorf("list users failed: %s", resp.Status)
	}
	var parsed listUsersByDepartmentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListUsersByDepartmentResult{}, err
	}
	if parsed.Code != 0 {
		return ListUsersByDepartmentResult{}, fmt.Errorf("list users failed: %s", parsed.Msg)
	}
	return ListUsersByDepartmentResult{
		Items:     parsed.Data.Items,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type DriveFile struct {
	Token     string `json:"token"`
	Name      string `json:"name"`
	FileType  string `json:"type"`
	URL       string `json:"url"`
	ParentID  string `json:"parent_token"`
	OwnerID   string `json:"owner_id"`
	OwnerType string `json:"owner_id_type"`
}

type GetDriveFileRequest struct {
	FileToken string
}

type UploadDriveFileRequest struct {
	FileName    string
	FolderToken string
	Size        int64
	File        io.Reader
}

type DriveUploadResult struct {
	FileToken string
	File      DriveFile
}

type uploadDriveFileResponse struct {
	apiResponse
	Data struct {
		FileToken string    `json:"file_token"`
		File      DriveFile `json:"file"`
	} `json:"data"`
}

func (c *Client) UploadDriveFile(ctx context.Context, token string, req UploadDriveFileRequest) (DriveUploadResult, error) {
	if req.File == nil {
		return DriveUploadResult{}, fmt.Errorf("file is required")
	}
	if req.FileName == "" {
		return DriveUploadResult{}, fmt.Errorf("file name is required")
	}
	if req.Size < 0 {
		return DriveUploadResult{}, fmt.Errorf("file size must be non-negative")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/upload_all", nil)
	if err != nil {
		return DriveUploadResult{}, err
	}

	pipeReader, pipeWriter := io.Pipe()
	writer := multipart.NewWriter(pipeWriter)
	contentType := writer.FormDataContentType()

	go func() {
		defer pipeWriter.Close()
		defer writer.Close()
		if err := writer.WriteField("file_name", req.FileName); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		parentNode := req.FolderToken
		if parentNode == "" {
			parentNode = "root"
		}
		if err := writer.WriteField("parent_type", "explorer"); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		if err := writer.WriteField("parent_node", parentNode); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		if err := writer.WriteField("size", fmt.Sprintf("%d", req.Size)); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		part, err := writer.CreateFormFile("file", req.FileName)
		if err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
		if _, err := io.Copy(part, req.File); err != nil {
			pipeWriter.CloseWithError(err)
			return
		}
	}()

	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, pipeReader)
	if err != nil {
		return DriveUploadResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", contentType)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DriveUploadResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DriveUploadResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DriveUploadResult{}, fmt.Errorf("drive upload failed: %s", resp.Status)
	}
	var parsed uploadDriveFileResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DriveUploadResult{}, err
	}
	if parsed.Code != 0 {
		return DriveUploadResult{}, fmt.Errorf("drive upload failed: %s", parsed.Msg)
	}
	result := DriveUploadResult{
		FileToken: parsed.Data.FileToken,
		File:      parsed.Data.File,
	}
	if result.FileToken == "" {
		result.FileToken = result.File.Token
	}
	if result.FileToken == "" {
		return DriveUploadResult{}, fmt.Errorf("drive upload response missing file token")
	}
	return result, nil
}

type ListDriveFilesRequest struct {
	FolderToken string
	PageSize    int
	PageToken   string
}

type listDriveFilesResponse struct {
	apiResponse
	Data struct {
		Files     []DriveFile `json:"files"`
		PageToken string      `json:"page_token"`
		HasMore   bool        `json:"has_more"`
	} `json:"data"`
}

type ListDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

func (c *Client) ListDriveFiles(ctx context.Context, token string, req ListDriveFilesRequest) (ListDriveFilesResult, error) {
	query := url.Values{}
	if req.FolderToken != "" {
		query.Set("folder_token", req.FolderToken)
	}
	if req.PageSize > 0 {
		query.Set("page_size", fmt.Sprintf("%d", req.PageSize))
	}
	if req.PageToken != "" {
		query.Set("page_token", req.PageToken)
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files", query)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ListDriveFilesResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", resp.Status)
	}
	var parsed listDriveFilesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ListDriveFilesResult{}, err
	}
	if parsed.Code != 0 {
		return ListDriveFilesResult{}, fmt.Errorf("list drive files failed: %s", parsed.Msg)
	}
	return ListDriveFilesResult{
		Files:     parsed.Data.Files,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type SearchDriveFilesRequest struct {
	Query     string
	PageSize  int
	PageToken string
}

type searchDriveFilesResponse struct {
	apiResponse
	Data struct {
		Files     []DriveFile `json:"files"`
		PageToken string      `json:"page_token"`
		HasMore   bool        `json:"has_more"`
	} `json:"data"`
}

type SearchDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

func (c *Client) SearchDriveFiles(ctx context.Context, token string, req SearchDriveFilesRequest) (SearchDriveFilesResult, error) {
	payload := map[string]any{
		"query": req.Query,
	}
	if req.PageSize > 0 {
		payload["page_size"] = req.PageSize
	}
	if req.PageToken != "" {
		payload["page_token"] = req.PageToken
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/search", nil)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SearchDriveFilesResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SearchDriveFilesResult{}, fmt.Errorf("search drive files failed: %s", resp.Status)
	}
	var parsed searchDriveFilesResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SearchDriveFilesResult{}, err
	}
	if parsed.Code != 0 {
		return SearchDriveFilesResult{}, fmt.Errorf("search drive files failed: %s", parsed.Msg)
	}
	return SearchDriveFilesResult{
		Files:     parsed.Data.Files,
		PageToken: parsed.Data.PageToken,
		HasMore:   parsed.Data.HasMore,
	}, nil
}

type getDriveFileResponse struct {
	apiResponse
	Data struct {
		File DriveFile `json:"file"`
	} `json:"data"`
}

func (c *Client) GetDriveFileMetadata(ctx context.Context, token, fileToken string) (DriveFile, error) {
	if fileToken == "" {
		return DriveFile{}, fmt.Errorf("file token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/"+url.PathEscape(fileToken), nil)
	if err != nil {
		return DriveFile{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DriveFile{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DriveFile{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DriveFile{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DriveFile{}, fmt.Errorf("get drive file failed: %s", resp.Status)
	}
	var parsed getDriveFileResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DriveFile{}, err
	}
	if parsed.Code != 0 {
		return DriveFile{}, fmt.Errorf("get drive file failed: %s", parsed.Msg)
	}
	return parsed.Data.File, nil
}

func (c *Client) GetDriveFile(ctx context.Context, token, fileToken string) (DriveFile, error) {
	return c.GetDriveFileMetadata(ctx, token, fileToken)
}

func (c *Client) DownloadDriveFile(ctx context.Context, token, fileToken string) (io.ReadCloser, error) {
	if fileToken == "" {
		return nil, fmt.Errorf("file token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/files/"+url.PathEscape(fileToken)+"/download", nil)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("drive download failed: %s", resp.Status)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("drive download failed: %s", resp.Status)
		}
		return nil, fmt.Errorf("drive download failed: %s: %s", resp.Status, string(bytes.TrimSpace(data)))
	}
	return resp.Body, nil
}

type DrivePermissionPublic struct {
	ExternalAccess  bool   `json:"external_access"`
	SecurityEntity  string `json:"security_entity"`
	CommentEntity   string `json:"comment_entity"`
	ShareEntity     string `json:"share_entity"`
	LinkShareEntity string `json:"link_share_entity"`
	InviteExternal  bool   `json:"invite_external"`
	LockSwitch      bool   `json:"lock_switch"`
}

type UpdateDrivePermissionPublicRequest struct {
	ExternalAccess  *bool  `json:"external_access,omitempty"`
	SecurityEntity  string `json:"security_entity,omitempty"`
	CommentEntity   string `json:"comment_entity,omitempty"`
	ShareEntity     string `json:"share_entity,omitempty"`
	LinkShareEntity string `json:"link_share_entity,omitempty"`
	InviteExternal  *bool  `json:"invite_external,omitempty"`
}

type updateDrivePermissionPublicResponse struct {
	apiResponse
	Data struct {
		Permission DrivePermissionPublic `json:"permission_public"`
	} `json:"data"`
}

func (c *Client) UpdateDrivePermissionPublic(ctx context.Context, token, fileToken, fileType string, req UpdateDrivePermissionPublicRequest) (DrivePermissionPublic, error) {
	if fileToken == "" {
		return DrivePermissionPublic{}, fmt.Errorf("file token is required")
	}
	if fileType == "" {
		return DrivePermissionPublic{}, fmt.Errorf("file type is required")
	}
	if !hasDrivePermissionPublicUpdate(req) {
		return DrivePermissionPublic{}, fmt.Errorf("permission update requires at least one field")
	}
	body, err := json.Marshal(req)
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	query := url.Values{}
	query.Set("type", fileType)
	endpoint, err := c.endpoint("/open-apis/drive/v1/permissions/"+url.PathEscape(fileToken)+"/public", query)
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPatch, endpoint, bytes.NewReader(body))
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DrivePermissionPublic{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DrivePermissionPublic{}, fmt.Errorf("update drive permission failed: %s", resp.Status)
	}
	var parsed updateDrivePermissionPublicResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DrivePermissionPublic{}, err
	}
	if parsed.Code != 0 {
		return DrivePermissionPublic{}, fmt.Errorf("update drive permission failed: %s", parsed.Msg)
	}
	return parsed.Data.Permission, nil
}

func hasDrivePermissionPublicUpdate(req UpdateDrivePermissionPublicRequest) bool {
	if req.ExternalAccess != nil || req.InviteExternal != nil {
		return true
	}
	if req.SecurityEntity != "" || req.CommentEntity != "" || req.ShareEntity != "" || req.LinkShareEntity != "" {
		return true
	}
	return false
}

type DocxDocument struct {
	DocumentID string `json:"document_id"`
	Title      string `json:"title"`
	URL        string `json:"url"`
	RevisionID string `json:"revision_id"`
}

type CreateDocxDocumentRequest struct {
	Title       string
	FolderToken string
}

type createDocxDocumentResponse struct {
	apiResponse
	Data struct {
		Document DocxDocument `json:"document"`
	} `json:"data"`
}

func (c *Client) CreateDocxDocument(ctx context.Context, token string, req CreateDocxDocumentRequest) (DocxDocument, error) {
	payload := map[string]any{
		"title": req.Title,
	}
	if req.FolderToken != "" {
		payload["folder_token"] = req.FolderToken
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return DocxDocument{}, err
	}
	endpoint, err := c.endpoint("/open-apis/docx/v1/documents", nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return DocxDocument{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DocxDocument{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", resp.Status)
	}
	var parsed createDocxDocumentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DocxDocument{}, err
	}
	if parsed.Code != 0 {
		return DocxDocument{}, fmt.Errorf("create docx document failed: %s", parsed.Msg)
	}
	return parsed.Data.Document, nil
}

type getDocxDocumentResponse struct {
	apiResponse
	Data struct {
		Document DocxDocument `json:"document"`
	} `json:"data"`
}

func (c *Client) GetDocxDocument(ctx context.Context, token, documentID string) (DocxDocument, error) {
	if documentID == "" {
		return DocxDocument{}, fmt.Errorf("document id is required")
	}
	endpoint, err := c.endpoint("/open-apis/docx/v1/documents/"+url.PathEscape(documentID), nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return DocxDocument{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return DocxDocument{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return DocxDocument{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", resp.Status)
	}
	var parsed getDocxDocumentResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return DocxDocument{}, err
	}
	if parsed.Code != 0 {
		return DocxDocument{}, fmt.Errorf("get docx document failed: %s", parsed.Msg)
	}
	return parsed.Data.Document, nil
}

type CreateExportTaskRequest struct {
	Token         string
	Type          string
	FileExtension string
}

type createExportTaskResponse struct {
	apiResponse
	Data struct {
		Ticket string `json:"ticket"`
	} `json:"data"`
}

func (c *Client) CreateExportTask(ctx context.Context, token string, req CreateExportTaskRequest) (string, error) {
	if req.Token == "" {
		return "", fmt.Errorf("export token is required")
	}
	if req.Type == "" {
		return "", fmt.Errorf("export type is required")
	}
	if req.FileExtension == "" {
		return "", fmt.Errorf("file extension is required")
	}
	payload := map[string]any{
		"token":          req.Token,
		"type":           req.Type,
		"file_extension": req.FileExtension,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/export_tasks", nil)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("create export task failed: %s", resp.Status)
	}
	var parsed createExportTaskResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("create export task failed: %s", parsed.Msg)
	}
	if parsed.Data.Ticket == "" {
		return "", errors.New("export task response missing ticket")
	}
	return parsed.Data.Ticket, nil
}

type ExportTaskResult struct {
	FileExtension string `json:"file_extension"`
	Type          string `json:"type"`
	FileName      string `json:"file_name"`
	FileToken     string `json:"file_token"`
	FileSize      int64  `json:"file_size"`
	JobErrorMsg   string `json:"job_error_msg"`
	JobStatus     int    `json:"job_status"`
}

type getExportTaskResponse struct {
	apiResponse
	Data struct {
		Result ExportTaskResult `json:"result"`
	} `json:"data"`
}

func (c *Client) GetExportTask(ctx context.Context, token, ticket string) (ExportTaskResult, error) {
	if ticket == "" {
		return ExportTaskResult{}, fmt.Errorf("export ticket is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/export_tasks/"+url.PathEscape(ticket), nil)
	if err != nil {
		return ExportTaskResult{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return ExportTaskResult{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return ExportTaskResult{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ExportTaskResult{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return ExportTaskResult{}, fmt.Errorf("get export task failed: %s", resp.Status)
	}
	var parsed getExportTaskResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return ExportTaskResult{}, err
	}
	if parsed.Code != 0 {
		return ExportTaskResult{}, fmt.Errorf("get export task failed: %s", parsed.Msg)
	}
	return parsed.Data.Result, nil
}

func (c *Client) DownloadExportedFile(ctx context.Context, token, fileToken string) (io.ReadCloser, error) {
	if fileToken == "" {
		return nil, fmt.Errorf("export file token is required")
	}
	endpoint, err := c.endpoint("/open-apis/drive/v1/export_tasks/file/"+url.PathEscape(fileToken)+"/download", nil)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		data, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("export download failed: %s", resp.Status)
		}
		if len(data) == 0 {
			return nil, fmt.Errorf("export download failed: %s", resp.Status)
		}
		return nil, fmt.Errorf("export download failed: %s: %s", resp.Status, string(bytes.TrimSpace(data)))
	}
	return resp.Body, nil
}

type SheetValueRange struct {
	Range          string  `json:"range"`
	MajorDimension string  `json:"major_dimension"`
	Values         [][]any `json:"values"`
}

type SheetValueRangeInput struct {
	Range          string  `json:"range"`
	MajorDimension string  `json:"major_dimension,omitempty"`
	Values         [][]any `json:"values"`
}

type readSheetRangeResponse struct {
	apiResponse
	Data struct {
		ValueRange SheetValueRange `json:"valueRange"`
	} `json:"data"`
}

type SheetValueUpdate struct {
	SpreadsheetToken string `json:"spreadsheetToken"`
	UpdatedRange     string `json:"updatedRange"`
	UpdatedRows      int    `json:"updatedRows"`
	UpdatedColumns   int    `json:"updatedColumns"`
	UpdatedCells     int    `json:"updatedCells"`
	Revision         int64  `json:"revision"`
}

type SheetValueAppend struct {
	SpreadsheetToken string           `json:"spreadsheetToken"`
	TableRange       string           `json:"tableRange"`
	Revision         int64            `json:"revision"`
	Updates          SheetValueUpdate `json:"updates"`
}

type updateSheetRangeResponse struct {
	apiResponse
	Data SheetValueUpdate `json:"data"`
}

type appendSheetRangeResponse struct {
	apiResponse
	Data SheetValueAppend `json:"data"`
}

type clearSheetRangeResponse struct {
	apiResponse
	Data map[string]any `json:"data"`
}

type SpreadsheetProperties struct {
	Title      string `json:"title"`
	OwnerUser  int64  `json:"ownerUser"`
	SheetCount int    `json:"sheetCount"`
	Revision   int64  `json:"revision"`
}

type SpreadsheetSheet struct {
	SheetID string `json:"sheetId"`
	Title   string `json:"title"`
	Index   int    `json:"index"`
	Hidden  bool   `json:"hidden"`
}

type SpreadsheetMetadata struct {
	Properties SpreadsheetProperties `json:"properties"`
	Sheets     []SpreadsheetSheet    `json:"sheets"`
}

type spreadsheetMetadataResponse struct {
	apiResponse
	Data SpreadsheetMetadata `json:"data"`
}

func (c *Client) ReadSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (SheetValueRange, error) {
	if spreadsheetToken == "" {
		return SheetValueRange{}, fmt.Errorf("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueRange{}, fmt.Errorf("range is required")
	}
	path := fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/values/%s", url.PathEscape(spreadsheetToken), url.PathEscape(sheetRange))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return SheetValueRange{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SheetValueRange{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SheetValueRange{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SheetValueRange{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SheetValueRange{}, fmt.Errorf("read sheet range failed: %s", resp.Status)
	}
	var parsed readSheetRangeResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SheetValueRange{}, err
	}
	if parsed.Code != 0 {
		return SheetValueRange{}, fmt.Errorf("read sheet range failed: %s", parsed.Msg)
	}
	return parsed.Data.ValueRange, nil
}

func (c *Client) UpdateSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any) (SheetValueUpdate, error) {
	if spreadsheetToken == "" {
		return SheetValueUpdate{}, fmt.Errorf("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueUpdate{}, fmt.Errorf("range is required")
	}
	if len(values) == 0 {
		return SheetValueUpdate{}, fmt.Errorf("values are required")
	}
	payload := map[string]any{
		"valueRange": SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SheetValueUpdate{}, err
	}
	path := fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/values", url.PathEscape(spreadsheetToken))
	query := url.Values{}
	query.Set("valueInputOption", "RAW")
	endpoint, err := c.endpoint(path, query)
	if err != nil {
		return SheetValueUpdate{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPut, endpoint, bytes.NewReader(body))
	if err != nil {
		return SheetValueUpdate{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SheetValueUpdate{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SheetValueUpdate{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SheetValueUpdate{}, fmt.Errorf("update sheet range failed: %s", resp.Status)
	}
	var parsed updateSheetRangeResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SheetValueUpdate{}, err
	}
	if parsed.Code != 0 {
		return SheetValueUpdate{}, fmt.Errorf("update sheet range failed: %s", parsed.Msg)
	}
	return parsed.Data, nil
}

func (c *Client) AppendSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string, values [][]any, insertDataOption string) (SheetValueAppend, error) {
	if spreadsheetToken == "" {
		return SheetValueAppend{}, fmt.Errorf("spreadsheet token is required")
	}
	if sheetRange == "" {
		return SheetValueAppend{}, fmt.Errorf("range is required")
	}
	if len(values) == 0 {
		return SheetValueAppend{}, fmt.Errorf("values are required")
	}
	payload := map[string]any{
		"valueRange": SheetValueRangeInput{
			Range:  sheetRange,
			Values: values,
		},
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return SheetValueAppend{}, err
	}
	path := fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/values_append", url.PathEscape(spreadsheetToken))
	query := url.Values{}
	if insertDataOption != "" {
		query.Set("insertDataOption", insertDataOption)
	}
	endpoint, err := c.endpoint(path, query)
	if err != nil {
		return SheetValueAppend{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return SheetValueAppend{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SheetValueAppend{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SheetValueAppend{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SheetValueAppend{}, fmt.Errorf("append sheet range failed: %s", resp.Status)
	}
	var parsed appendSheetRangeResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SheetValueAppend{}, err
	}
	if parsed.Code != 0 {
		return SheetValueAppend{}, fmt.Errorf("append sheet range failed: %s", parsed.Msg)
	}
	return parsed.Data, nil
}

func (c *Client) ClearSheetRange(ctx context.Context, token, spreadsheetToken, sheetRange string) (string, error) {
	if spreadsheetToken == "" {
		return "", fmt.Errorf("spreadsheet token is required")
	}
	if sheetRange == "" {
		return "", fmt.Errorf("range is required")
	}
	payload := map[string]string{
		"range": sheetRange,
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}
	path := fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/values_clear", url.PathEscape(spreadsheetToken))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return "", err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	request.Header.Set("Authorization", "Bearer "+token)
	request.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return "", fmt.Errorf("clear sheet range failed: %s", resp.Status)
	}
	var parsed clearSheetRangeResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return "", err
	}
	if parsed.Code != 0 {
		return "", fmt.Errorf("clear sheet range failed: %s", parsed.Msg)
	}
	clearedRange := sheetRange
	if parsed.Data != nil {
		if value, ok := parsed.Data["clearedRange"].(string); ok && value != "" {
			clearedRange = value
		} else if value, ok := parsed.Data["cleared_range"].(string); ok && value != "" {
			clearedRange = value
		}
	}
	return clearedRange, nil
}

func (c *Client) GetSpreadsheetMetadata(ctx context.Context, token, spreadsheetToken string) (SpreadsheetMetadata, error) {
	if spreadsheetToken == "" {
		return SpreadsheetMetadata{}, fmt.Errorf("spreadsheet token is required")
	}
	path := fmt.Sprintf("/open-apis/sheets/v2/spreadsheets/%s/metainfo", url.PathEscape(spreadsheetToken))
	endpoint, err := c.endpoint(path, nil)
	if err != nil {
		return SpreadsheetMetadata{}, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return SpreadsheetMetadata{}, err
	}
	request.Header.Set("Authorization", "Bearer "+token)

	resp, err := c.httpClient().Do(request)
	if err != nil {
		return SpreadsheetMetadata{}, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return SpreadsheetMetadata{}, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return SpreadsheetMetadata{}, fmt.Errorf("get spreadsheet metadata failed: %s", resp.Status)
	}
	var parsed spreadsheetMetadataResponse
	if err := json.Unmarshal(data, &parsed); err != nil {
		return SpreadsheetMetadata{}, err
	}
	if parsed.Code != 0 {
		return SpreadsheetMetadata{}, fmt.Errorf("get spreadsheet metadata failed: %s", parsed.Msg)
	}
	return parsed.Data, nil
}
