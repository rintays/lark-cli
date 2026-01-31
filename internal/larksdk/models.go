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

type ListCalendarEventsResult struct {
	Items     []CalendarEvent
	PageToken string
	HasMore   bool
	SyncToken string
}

type CreateCalendarEventRequest struct {
	CalendarID  string
	Summary     string
	Description string
	StartTime   int64
	EndTime     int64
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

type ListMinutesRequest struct {
	PageSize   int
	PageToken  string
	UserIDType string
}

type ListMinutesResult struct {
	Items     []Minute
	PageToken string
	HasMore   bool
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

type DocxDocument struct {
	DocumentID string     `json:"document_id"`
	Title      string     `json:"title"`
	URL        string     `json:"url"`
	RevisionID RevisionID `json:"revision_id"`
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

type MailFolder struct {
	FolderID       string `json:"folder_id"`
	Name           string `json:"name"`
	ParentFolderID string `json:"parent_folder_id,omitempty"`
	FolderType     string `json:"folder_type,omitempty"`
}
