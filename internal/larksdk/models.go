package larksdk

import (
	"encoding/json"
	"fmt"
)

type TenantInfo struct {
	TenantKey string `json:"tenant_key"`
	Name      string `json:"name"`
}

type Calendar struct {
	CalendarID string `json:"calendar_id"`
	Summary    string `json:"summary"`
}

type CalendarEventTime struct {
	Date      string `json:"date"`
	Timestamp string `json:"timestamp"`
	Timezone  string `json:"timezone"`
}

type CalendarEventMeetingSettings struct {
	OwnerID               string   `json:"owner_id,omitempty"`
	JoinMeetingPermission string   `json:"join_meeting_permission,omitempty"`
	Password              string   `json:"password,omitempty"`
	AssignHosts           []string `json:"assign_hosts,omitempty"`
	AutoRecord            *bool    `json:"auto_record,omitempty"`
	OpenLobby             *bool    `json:"open_lobby,omitempty"`
	AllowAttendeesStart   *bool    `json:"allow_attendees_start,omitempty"`
}

type CalendarEventVChat struct {
	VCType          string                        `json:"vc_type,omitempty"`
	IconType        string                        `json:"icon_type,omitempty"`
	Description     string                        `json:"description,omitempty"`
	MeetingURL      string                        `json:"meeting_url,omitempty"`
	MeetingSettings *CalendarEventMeetingSettings `json:"meeting_settings,omitempty"`
}

type CalendarEventLocation struct {
	Name      string  `json:"name,omitempty"`
	Address   string  `json:"address,omitempty"`
	Latitude  float64 `json:"latitude,omitempty"`
	Longitude float64 `json:"longitude,omitempty"`
}

type CalendarEventReminder struct {
	Minutes int `json:"minutes"`
}

type CalendarEventSchema struct {
	UIName   string `json:"ui_name,omitempty"`
	UIStatus string `json:"ui_status,omitempty"`
	AppLink  string `json:"app_link,omitempty"`
}

type CalendarEventOrganizer struct {
	UserID      string `json:"user_id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type CalendarEventAttachment struct {
	FileToken string `json:"file_token,omitempty"`
	FileSize  string `json:"file_size,omitempty"`
	Name      string `json:"name,omitempty"`
	IsDeleted *bool  `json:"is_deleted,omitempty"`
}

type CalendarEventCheckInTime struct {
	TimeType string `json:"time_type,omitempty"`
	Duration *int   `json:"duration,omitempty"`
}

type CalendarEventCheckIn struct {
	EnableCheckIn       *bool                     `json:"enable_check_in,omitempty"`
	CheckInStartTime    *CalendarEventCheckInTime `json:"check_in_start_time,omitempty"`
	CheckInEndTime      *CalendarEventCheckInTime `json:"check_in_end_time,omitempty"`
	NeedNotifyAttendees *bool                     `json:"need_notify_attendees,omitempty"`
}

type CalendarEventAttendeeChatMember struct {
	RsvpStatus  string `json:"rsvp_status,omitempty"`
	IsOptional  *bool  `json:"is_optional,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	IsOrganizer *bool  `json:"is_organizer,omitempty"`
	IsExternal  *bool  `json:"is_external,omitempty"`
	UserID      string `json:"user_id,omitempty"`
}

type CalendarEvent struct {
	EventID             string                    `json:"event_id"`
	OrganizerCalendarID string                    `json:"organizer_calendar_id,omitempty"`
	Summary             string                    `json:"summary"`
	Description         string                    `json:"description"`
	StartTime           CalendarEventTime         `json:"start_time"`
	EndTime             CalendarEventTime         `json:"end_time"`
	VChat               *CalendarEventVChat       `json:"vchat,omitempty"`
	Visibility          string                    `json:"visibility,omitempty"`
	AttendeeAbility     string                    `json:"attendee_ability,omitempty"`
	FreeBusyStatus      string                    `json:"free_busy_status,omitempty"`
	Location            *CalendarEventLocation    `json:"location,omitempty"`
	Color               *int                      `json:"color,omitempty"`
	Reminders           []CalendarEventReminder   `json:"reminders,omitempty"`
	Recurrence          string                    `json:"recurrence,omitempty"`
	Status              string                    `json:"status"`
	IsException         *bool                     `json:"is_exception,omitempty"`
	RecurringEventID    string                    `json:"recurring_event_id,omitempty"`
	CreateTime          string                    `json:"create_time,omitempty"`
	Schemas             []CalendarEventSchema     `json:"schemas,omitempty"`
	EventOrganizer      *CalendarEventOrganizer   `json:"event_organizer,omitempty"`
	AppLink             string                    `json:"app_link,omitempty"`
	Attendees           []CalendarEventAttendee   `json:"attendees,omitempty"`
	HasMoreAttendee     *bool                     `json:"has_more_attendee,omitempty"`
	Attachments         []CalendarEventAttachment `json:"attachments,omitempty"`
	EventCheckIn        *CalendarEventCheckIn     `json:"event_check_in,omitempty"`
}

type ListCalendarEventsRequest struct {
	CalendarID string
	StartTime  string
	EndTime    string
	PageSize   int
	PageToken  string
	SyncToken  string
}

type ListCalendarEventsResult struct {
	Items     []CalendarEvent
	PageToken string
	HasMore   bool
	SyncToken string
}

type SearchCalendarEventsRequest struct {
	CalendarID string
	Query      string
	StartTime  string
	EndTime    string
	UserIDs    []string
	RoomIDs    []string
	ChatIDs    []string
	PageSize   int
	PageToken  string
}

type SearchCalendarEventsResult struct {
	Items     []CalendarEvent
	PageToken string
}

type GetCalendarEventRequest struct {
	CalendarID          string
	EventID             string
	NeedMeetingSettings *bool
	NeedAttendee        *bool
	MaxAttendeeNum      *int
	UserIDType          string
}

type UpdateCalendarEventRequest struct {
	CalendarID       string
	EventID          string
	Summary          string
	Description      string
	StartTime        *int64
	EndTime          *int64
	Start            *CalendarEventTime
	End              *CalendarEventTime
	NeedNotification *bool
	Visibility       string
	AttendeeAbility  string
	FreeBusyStatus   string
	Location         *CalendarEventLocation
	Color            *int
	Reminders        []CalendarEventReminder
	Recurrence       string
	VChat            *CalendarEventVChat
	Schemas          []CalendarEventSchema
	Attachments      []CalendarEventAttachment
	EventCheckIn     *CalendarEventCheckIn
	Extra            map[string]any
	UserIDType       string
}

type DeleteCalendarEventRequest struct {
	CalendarID       string
	EventID          string
	NeedNotification bool
}

type DeleteCalendarEventResult struct {
	EventID string
	Deleted bool
}

type CreateCalendarEventRequest struct {
	CalendarID       string
	Summary          string
	Description      string
	StartTime        int64
	EndTime          int64
	Start            *CalendarEventTime
	End              *CalendarEventTime
	NeedNotification *bool
	Visibility       string
	AttendeeAbility  string
	FreeBusyStatus   string
	Location         *CalendarEventLocation
	Color            *int
	Reminders        []CalendarEventReminder
	Recurrence       string
	VChat            *CalendarEventVChat
	Schemas          []CalendarEventSchema
	Attachments      []CalendarEventAttachment
	EventCheckIn     *CalendarEventCheckIn
	Extra            map[string]any
	IdempotencyKey   string
	UserIDType       string
}

type CalendarEventAttendee struct {
	Type            string                            `json:"type,omitempty"`
	AttendeeID      string                            `json:"attendee_id,omitempty"`
	UserID          string                            `json:"user_id,omitempty"`
	ChatID          string                            `json:"chat_id,omitempty"`
	RoomID          string                            `json:"room_id,omitempty"`
	ThirdPartyEmail string                            `json:"third_party_email,omitempty"`
	OperateID       string                            `json:"operate_id,omitempty"`
	RsvpStatus      string                            `json:"rsvp_status,omitempty"`
	IsOptional      *bool                             `json:"is_optional,omitempty"`
	IsOrganizer     *bool                             `json:"is_organizer,omitempty"`
	IsExternal      *bool                             `json:"is_external,omitempty"`
	DisplayName     string                            `json:"display_name,omitempty"`
	ChatMembers     []CalendarEventAttendeeChatMember `json:"chat_members,omitempty"`
}

type CreateCalendarEventAttendeesRequest struct {
	CalendarID string
	EventID    string
	Attendees  []CalendarEventAttendee
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

type MeetingListItem struct {
	ID        string `json:"id"`
	Topic     string `json:"topic"`
	Status    *int   `json:"status,omitempty"`
	StartTime string `json:"start_time"`
	EndTime   string `json:"end_time"`
}

type ListMeetingsRequest struct {
	StartTime               string
	EndTime                 string
	MeetingStatus           *int
	MeetingNo               string
	UserID                  string
	RoomID                  string
	MeetingType             *int
	PageSize                int
	PageToken               string
	IncludeExternalMeetings *bool
	IncludeWebinar          *bool
	UserIDType              string
}

type ListMeetingsResult struct {
	Items     []MeetingListItem
	PageToken string
	HasMore   bool
}

type ReserveMeetingSetting struct {
	Topic              *string `json:"topic,omitempty"`
	MeetingInitialType *int    `json:"meeting_initial_type,omitempty"`
	AutoRecord         *bool   `json:"auto_record,omitempty"`
	Password           *string `json:"password,omitempty"`
}

type Reserve struct {
	ID              string                 `json:"id"`
	MeetingNo       string                 `json:"meeting_no"`
	Password        string                 `json:"password"`
	URL             string                 `json:"url"`
	AppLink         string                 `json:"app_link"`
	LiveLink        string                 `json:"live_link"`
	EndTime         string                 `json:"end_time"`
	ExpireStatus    *int                   `json:"expire_status,omitempty"`
	ReserveUserID   string                 `json:"reserve_user_id"`
	MeetingSettings *ReserveMeetingSetting `json:"meeting_settings,omitempty"`
}

type ReserveCorrectionCheckInfo struct {
	InvalidHostIDList []string `json:"invalid_host_id_list,omitempty"`
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

type BatchGetUserIDRequest struct {
	Emails  []string
	Mobiles []string
}

type ListUsersByDepartmentRequest struct {
	DepartmentID string
	PageSize     int
	PageToken    string
	UserIDType   string
}

type ListUsersByDepartmentResult struct {
	Items     []User
	PageToken string
	HasMore   bool
}

type Minute struct {
	Token      string `json:"token"`
	OwnerID    string `json:"owner_id,omitempty"`
	CreateTime string `json:"create_time,omitempty"`
	Title      string `json:"title,omitempty"`
	Cover      string `json:"cover,omitempty"`
	Duration   string `json:"duration,omitempty"`
	URL        string `json:"url,omitempty"`
}

type RevisionID string

func (r *RevisionID) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*r = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*r = RevisionID(value)
		return nil
	}
	var value json.Number
	if err := json.Unmarshal(data, &value); err == nil {
		*r = RevisionID(value.String())
		return nil
	}
	return fmt.Errorf("revision_id must be a string or number, got %s", string(data))
}

type DocxDisplaySetting struct {
	ShowAuthors        *bool `json:"show_authors,omitempty"`
	ShowCreateTime     *bool `json:"show_create_time,omitempty"`
	ShowPv             *bool `json:"show_pv,omitempty"`
	ShowUv             *bool `json:"show_uv,omitempty"`
	ShowLikeCount      *bool `json:"show_like_count,omitempty"`
	ShowCommentCount   *bool `json:"show_comment_count,omitempty"`
	ShowRelatedMatters *bool `json:"show_related_matters,omitempty"`
}

type DocxCover struct {
	Token        string   `json:"token,omitempty"`
	OffsetRatioX *float64 `json:"offset_ratio_x,omitempty"`
	OffsetRatioY *float64 `json:"offset_ratio_y,omitempty"`
}

type DocxDocument struct {
	DocumentID     string              `json:"document_id"`
	Title          string              `json:"title,omitempty"`
	URL            string              `json:"url,omitempty"`
	RevisionID     RevisionID          `json:"revision_id,omitempty"`
	DisplaySetting *DocxDisplaySetting `json:"display_setting,omitempty"`
	Cover          *DocxCover          `json:"cover,omitempty"`
}

type CreateDocxDocumentRequest struct {
	Title       string
	FolderToken string
}

type CreateExportTaskRequest struct {
	Token         string
	Type          string
	FileExtension string
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

type SheetValueRange struct {
	Range          string  `json:"range"`
	MajorDimension string  `json:"major_dimension"`
	Values         [][]any `json:"values"`
}

type SheetDimensionInsertResult struct {
	StartIndex int `json:"start_index"`
	Count      int `json:"count"`
	EndIndex   int `json:"end_index"`
}

type SheetDimensionDeleteResult struct {
	StartIndex int `json:"start_index"`
	Count      int `json:"count"`
	EndIndex   int `json:"end_index"`
}

type SpreadsheetGridProperties struct {
	FrozenRowCount    int `json:"frozenRowCount,omitempty"`
	FrozenColumnCount int `json:"frozenColumnCount,omitempty"`
	RowCount          int `json:"rowCount,omitempty"`
	ColumnCount       int `json:"columnCount,omitempty"`
}

type SpreadsheetMergeRange struct {
	StartRowIndex    int `json:"startRowIndex,omitempty"`
	EndRowIndex      int `json:"endRowIndex,omitempty"`
	StartColumnIndex int `json:"startColumnIndex,omitempty"`
	EndColumnIndex   int `json:"endColumnIndex,omitempty"`
}

type SpreadsheetSheet struct {
	SheetID        string                     `json:"sheetId"`
	Title          string                     `json:"title"`
	Index          int                        `json:"index"`
	Hidden         bool                       `json:"hidden"`
	ResourceType   string                     `json:"resourceType,omitempty"`
	GridProperties *SpreadsheetGridProperties `json:"gridProperties,omitempty"`
	Merges         []SpreadsheetMergeRange    `json:"merges,omitempty"`
}

type SpreadsheetMetadata struct {
	Spreadsheet SpreadsheetInfo    `json:"spreadsheet"`
	Sheets      []SpreadsheetSheet `json:"sheets,omitempty"`
}

type MailFolderType string

func (t *MailFolderType) UnmarshalJSON(data []byte) error {
	if len(data) == 0 || string(data) == "null" {
		*t = ""
		return nil
	}
	if data[0] == '"' {
		var value string
		if err := json.Unmarshal(data, &value); err != nil {
			return err
		}
		*t = MailFolderType(value)
		return nil
	}
	var value json.Number
	if err := json.Unmarshal(data, &value); err == nil {
		*t = MailFolderType(value.String())
		return nil
	}
	return fmt.Errorf("folder_type must be a string or number, got %s", string(data))
}

func (t MailFolderType) String() string {
	return string(t)
}

type MailFolder struct {
	FolderID       string         `json:"folder_id"`
	Name           string         `json:"name"`
	ParentFolderID string         `json:"parent_folder_id,omitempty"`
	FolderType     MailFolderType `json:"folder_type,omitempty"`
}

func (f *MailFolder) UnmarshalJSON(data []byte) error {
	var aux struct {
		ID             string         `json:"id"`
		FolderID       string         `json:"folder_id"`
		Name           string         `json:"name"`
		ParentFolderID string         `json:"parent_folder_id"`
		FolderType     MailFolderType `json:"folder_type"`
	}
	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}
	f.FolderID = aux.FolderID
	if f.FolderID == "" {
		f.FolderID = aux.ID
	}
	f.Name = aux.Name
	f.ParentFolderID = aux.ParentFolderID
	f.FolderType = aux.FolderType
	return nil
}
