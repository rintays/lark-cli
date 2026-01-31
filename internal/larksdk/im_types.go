package larksdk

import (
	larkdocx "github.com/larksuite/oapi-sdk-go/v3/service/docx/v1"
)

type Message struct {
	MessageID      string           `json:"message_id"`
	RootID         string           `json:"root_id,omitempty"`
	ParentID       string           `json:"parent_id,omitempty"`
	ThreadID       string           `json:"thread_id,omitempty"`
	MsgType        string           `json:"msg_type,omitempty"`
	CreateTime     string           `json:"create_time,omitempty"`
	UpdateTime     string           `json:"update_time,omitempty"`
	Deleted        bool             `json:"deleted,omitempty"`
	Updated        bool             `json:"updated,omitempty"`
	ChatID         string           `json:"chat_id,omitempty"`
	Sender         MessageSender    `json:"sender,omitempty"`
	Body           MessageBody      `json:"body,omitempty"`
	Mentions       []MessageMention `json:"mentions,omitempty"`
	UpperMessageID string           `json:"upper_message_id,omitempty"`
}

type MessageSender struct {
	ID         string `json:"id,omitempty"`
	IDType     string `json:"id_type,omitempty"`
	SenderType string `json:"sender_type,omitempty"`
	TenantKey  string `json:"tenant_key,omitempty"`
}

type MessageBody struct {
	Content string `json:"content,omitempty"`
}

type MessageMention struct {
	Key       string `json:"key,omitempty"`
	ID        string `json:"id,omitempty"`
	IDType    string `json:"id_type,omitempty"`
	Name      string `json:"name,omitempty"`
	TenantKey string `json:"tenant_key,omitempty"`
}

type ListMessagesRequest struct {
	ContainerIDType string
	ContainerID     string
	StartTime       string
	EndTime         string
	SortType        string
	PageSize        int
	PageToken       string
}

type ListMessagesResult struct {
	Items     []Message
	PageToken string
	HasMore   bool
}

type MessageSearchRequest struct {
	Query        string
	FromIDs      []string
	ChatIDs      []string
	MessageType  string
	AtChatterIDs []string
	FromType     string
	ChatType     string
	StartTime    string
	EndTime      string
	PageSize     int
	PageToken    string
	UserIDType   string
}

type MessageSearchResult struct {
	Items     []string
	PageToken string
	HasMore   bool
}

type MessageReaction struct {
	ReactionID string           `json:"reaction_id,omitempty"`
	Operator   ReactionOperator `json:"operator,omitempty"`
	ActionTime string           `json:"action_time,omitempty"`
	Reaction   Emoji            `json:"reaction_type,omitempty"`
}

type ReactionOperator struct {
	OperatorID   string `json:"operator_id,omitempty"`
	OperatorType string `json:"operator_type,omitempty"`
}

type Emoji struct {
	EmojiType string `json:"emoji_type,omitempty"`
}

type Pin struct {
	MessageID      string `json:"message_id,omitempty"`
	ChatID         string `json:"chat_id,omitempty"`
	OperatorID     string `json:"operator_id,omitempty"`
	OperatorIDType string `json:"operator_id_type,omitempty"`
	CreateTime     string `json:"create_time,omitempty"`
}

type I18nNames struct {
	ZhCn string `json:"zh_cn,omitempty"`
	EnUs string `json:"en_us,omitempty"`
	JaJp string `json:"ja_jp,omitempty"`
}

type RestrictedModeSetting struct {
	Status                         bool   `json:"status,omitempty"`
	ScreenshotHasPermissionSetting string `json:"screenshot_has_permission_setting,omitempty"`
	DownloadHasPermissionSetting   string `json:"download_has_permission_setting,omitempty"`
	MessageHasPermissionSetting    string `json:"message_has_permission_setting,omitempty"`
}

type ChatInfo struct {
	ChatID                 string                 `json:"chat_id"`
	Avatar                 string                 `json:"avatar,omitempty"`
	Name                   string                 `json:"name,omitempty"`
	Description            string                 `json:"description,omitempty"`
	I18nNames              *I18nNames             `json:"i18n_names,omitempty"`
	OwnerID                string                 `json:"owner_id,omitempty"`
	OwnerIDType            string                 `json:"owner_id_type,omitempty"`
	UserManagerIDList      []string               `json:"user_manager_id_list,omitempty"`
	BotManagerIDList       []string               `json:"bot_manager_id_list,omitempty"`
	GroupMessageType       string                 `json:"group_message_type,omitempty"`
	ChatMode               string                 `json:"chat_mode,omitempty"`
	ChatType               string                 `json:"chat_type,omitempty"`
	ChatTag                string                 `json:"chat_tag,omitempty"`
	JoinMessageVisibility  string                 `json:"join_message_visibility,omitempty"`
	LeaveMessageVisibility string                 `json:"leave_message_visibility,omitempty"`
	MembershipApproval     string                 `json:"membership_approval,omitempty"`
	ModerationPermission   string                 `json:"moderation_permission,omitempty"`
	AddMemberPermission    string                 `json:"add_member_permission,omitempty"`
	ShareCardPermission    string                 `json:"share_card_permission,omitempty"`
	AtAllPermission        string                 `json:"at_all_permission,omitempty"`
	EditPermission         string                 `json:"edit_permission,omitempty"`
	UrgentSetting          string                 `json:"urgent_setting,omitempty"`
	VideoConferenceSetting string                 `json:"video_conference_setting,omitempty"`
	PinManageSetting       string                 `json:"pin_manage_setting,omitempty"`
	HideMemberCountSetting string                 `json:"hide_member_count_setting,omitempty"`
	RestrictedModeSetting  *RestrictedModeSetting `json:"restricted_mode_setting,omitempty"`
	External               bool                   `json:"external"`
	TenantKey              string                 `json:"tenant_key,omitempty"`
	UserCount              string                 `json:"user_count,omitempty"`
	BotCount               string                 `json:"bot_count,omitempty"`
	ChatStatus             string                 `json:"chat_status,omitempty"`
}

type CreateChatRequest struct {
	UserIDType             string
	SetBotManager          *bool
	UUID                   string
	Avatar                 string
	Name                   string
	Description            string
	I18nNames              *I18nNames
	OwnerID                string
	UserIDList             []string
	BotIDList              []string
	GroupMessageType       string
	ChatMode               string
	ChatType               string
	JoinMessageVisibility  string
	LeaveMessageVisibility string
	MembershipApproval     string
	RestrictedModeSetting  *RestrictedModeSetting
	UrgentSetting          string
	VideoConferenceSetting string
	EditPermission         string
	HideMemberCountSetting string
	PinManageSetting       string
	External               *bool
}

type GetChatRequest struct {
	ChatID     string
	UserIDType string
}

type UpdateChatRequest struct {
	ChatID                 string
	UserIDType             string
	Avatar                 string
	Name                   string
	Description            string
	I18nNames              *I18nNames
	AddMemberPermission    string
	ShareCardPermission    string
	AtAllPermission        string
	EditPermission         string
	OwnerID                string
	JoinMessageVisibility  string
	LeaveMessageVisibility string
	MembershipApproval     string
	RestrictedModeSetting  *RestrictedModeSetting
	ChatType               string
	GroupMessageType       string
	UrgentSetting          string
	VideoConferenceSetting string
	PinManageSetting       string
	HideMemberCountSetting string
}

type ChatAnnouncement struct {
	Content          string            `json:"content,omitempty"`
	Revision         string            `json:"revision,omitempty"`
	CreateTime       string            `json:"create_time,omitempty"`
	UpdateTime       string            `json:"update_time,omitempty"`
	OwnerIDType      string            `json:"owner_id_type,omitempty"`
	OwnerID          string            `json:"owner_id,omitempty"`
	ModifierIDType   string            `json:"modifier_id_type,omitempty"`
	ModifierID       string            `json:"modifier_id,omitempty"`
	AnnouncementType string            `json:"announcement_type,omitempty"`
	RevisionID       RevisionID        `json:"revision_id,omitempty"`
	Blocks           []*larkdocx.Block `json:"blocks,omitempty"`
}
