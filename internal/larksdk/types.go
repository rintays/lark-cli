package larksdk

import (
	"encoding/json"
	"io"
)

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

type ListChatsResult struct {
	Items     []Chat
	PageToken string
	HasMore   bool
}

type MessageRequest struct {
	ReceiveID     string
	ReceiveIDType string
	Text          string
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

type ListDriveFilesRequest struct {
	FolderToken string
	PageSize    int
	PageToken   string
}

type ListDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

type SearchDriveFilesRequest struct {
	Query     string
	PageSize  int
	PageToken string
}

type SearchDriveFilesResult struct {
	Files     []DriveFile
	PageToken string
	HasMore   bool
}

type SheetValueRangeInput struct {
	Range          string  `json:"range"`
	MajorDimension string  `json:"major_dimension,omitempty"`
	Values         [][]any `json:"values"`
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

type ClearSheetRangeResult struct {
	ClearedRange string `json:"clearedRange"`
}

type BaseTable struct {
	TableID string `json:"table_id"`
	Name    string `json:"name"`
}

type BaseField struct {
	FieldID   string `json:"field_id"`
	FieldName string `json:"field_name"`
	Type      int    `json:"type"`
}

type BaseView struct {
	ViewID   string `json:"view_id"`
	Name     string `json:"name"`
	ViewType string `json:"view_type"`
}

type BaseRecord struct {
	RecordID         string         `json:"record_id"`
	Fields           map[string]any `json:"fields,omitempty"`
	CreatedTime      string         `json:"created_time"`
	LastModifiedTime string         `json:"last_modified_time"`
}

type SearchBaseRecordsRequest struct {
	ViewID   string          `json:"view_id,omitempty"`
	Filter   json.RawMessage `json:"filter,omitempty"`
	Sort     json.RawMessage `json:"sort,omitempty"`
	PageSize int             `json:"page_size,omitempty"`
}

type SearchBaseRecordsResult struct {
	Items     []BaseRecord `json:"items"`
	PageToken string       `json:"page_token"`
	HasMore   bool         `json:"has_more"`
}

type ListBaseTablesResult struct {
	Items     []BaseTable `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

type ListBaseFieldsResult struct {
	Items     []BaseField `json:"items"`
	PageToken string      `json:"page_token"`
	HasMore   bool        `json:"has_more"`
}

type ListBaseViewsResult struct {
	Items     []BaseView `json:"items"`
	PageToken string     `json:"page_token"`
	HasMore   bool       `json:"has_more"`
}
