package larksdk

import (
	"io"

	"lark/internal/larkapi"
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

type MessageRequest = larkapi.MessageRequest

type CreateExportTaskRequest = larkapi.CreateExportTaskRequest

type ExportTaskResult = larkapi.ExportTaskResult

type DocxDocument = larkapi.DocxDocument

type CreateDocxDocumentRequest = larkapi.CreateDocxDocumentRequest

type Meeting = larkapi.Meeting

type GetMeetingRequest = larkapi.GetMeetingRequest

type Minute = larkapi.Minute

type ListMinutesRequest = larkapi.ListMinutesRequest

type ListMinutesResult = larkapi.ListMinutesResult

type User = larkapi.User

type GetContactUserRequest = larkapi.GetContactUserRequest

type BatchGetUserIDRequest = larkapi.BatchGetUserIDRequest

type ListUsersByDepartmentRequest = larkapi.ListUsersByDepartmentRequest

type ListUsersByDepartmentResult = larkapi.ListUsersByDepartmentResult

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

type SheetValueRange = larkapi.SheetValueRange

type SpreadsheetMetadata = larkapi.SpreadsheetMetadata

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
