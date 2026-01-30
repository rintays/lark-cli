package larksdk

import "lark/internal/larkapi"

type Chat = larkapi.Chat

type ListChatsRequest = larkapi.ListChatsRequest

type ListChatsResult = larkapi.ListChatsResult

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

type DriveFile = larkapi.DriveFile

type GetDriveFileRequest = larkapi.GetDriveFileRequest

type UploadDriveFileRequest = larkapi.UploadDriveFileRequest

type DriveUploadResult = larkapi.DriveUploadResult

type ListDriveFilesRequest = larkapi.ListDriveFilesRequest

type ListDriveFilesResult = larkapi.ListDriveFilesResult

type SearchDriveFilesRequest = larkapi.SearchDriveFilesRequest

type SearchDriveFilesResult = larkapi.SearchDriveFilesResult

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
