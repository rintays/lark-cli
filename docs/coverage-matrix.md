# Coverage Matrix (SDK-first migration)

This doc is a **code-derived snapshot** of which Feishu/Lark OpenAPI surfaces in this repo are backed by the official **`oapi-sdk-go`** typed services vs a **`core.ApiReq`** wrapper in `internal/larksdk`.

**Columns**
- **Feature**: CLI surface / intent
- **Endpoint**: OpenAPI path used (when known)
- **Token**: `tenant` or `user`
- **Ver**: OpenAPI version (v1/v2/v3 or `docs-api`)
- **SDK?**: whether a typed SDK service is used
- **Wrapper (if no SDK)**: `internal/larksdk/<file>.go: Client.<Func>`

---

## Drive

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| List files (`drive list`) | `GET /open-apis/drive/v1/files` | tenant | v1 | yes |  |
| Download file (`drive download`) | `GET /open-apis/drive/v1/files/:file_token/download` | tenant | v1 | yes |  |
| Upload file (`drive upload`) | `POST /open-apis/drive/v1/files/upload_all` | tenant | v1 | yes |  |
| Export task create/get/download (`drive export`, `docs export`) | `/open-apis/drive/v1/export_tasks*` | tenant | v1 | yes |  |
| Search files (`drive search`) | `POST /open-apis/drive/v1/files/search` | tenant/user (CLI uses user) | v1 | no | `internal/larksdk/drive.go: Client.SearchDriveFiles` |
| Get file metadata (`drive info`; also used by `docs info` URL fill) | `GET /open-apis/drive/v1/files/:file_token` | tenant/user | v1 | no | `internal/larksdk/drive.go: Client.GetDriveFileMetadata` |
| Update public permission (`drive share`) | `PATCH /open-apis/drive/v1/permissions/:file_token/public` | tenant | v1 | no | `internal/larksdk/drive.go: Client.UpdateDrivePermissionPublic` |

## Docs (suite docs-api)

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| Search docs objects (`docs search` helper) | `POST /open-apis/suite/docs-api/search/object` | user | docs-api | no | `internal/larksdk/docs_search.go: Client.SearchDocsObjectsWithUserToken` |

## Docx (Docs documents)

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| Create document (`docs create`) | `POST /open-apis/docx/v1/documents` | tenant | v1 | yes |  |
| Get document (`docs info`) | `GET /open-apis/docx/v1/documents/:document_id` | tenant | v1 | yes |  |
| Markdown content (`docs get --format md`) | `GET /open-apis/docx/v1/documents/:document_id/blocks` | tenant | v1 | yes |  |
| Raw content (`docs get --format txt`) | `GET /open-apis/docx/v1/documents/:document_id/raw_content` | tenant | v1 | yes |  |
| List blocks (`docs get --format blocks`, `docs blocks …`) | `GET /open-apis/docx/v1/documents/:document_id/blocks` | tenant | v1 | yes |  |

## Sheets

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| Spreadsheet info (`sheets info`) | `GET /open-apis/sheets/v3/spreadsheets/:spreadsheet_token` | tenant | v3 | yes |  |
| List sheets/tabs (used by `sheets info`) | `GET /open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/query` | tenant | v3 | yes |  |
| Read range (`sheets read`) | `GET /open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values/:range` | tenant | v2 | no | `internal/larksdk/sheets.go: Client.ReadSheetRange` |
| Update range (`sheets update`) | `PUT /open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values` | tenant | v2 | no | `internal/larksdk/sheets.go: Client.UpdateSheetRange` |
| Append range (`sheets append`) | `POST /open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values_append` | tenant | v2 | no | `internal/larksdk/sheets.go: Client.AppendSheetRange` |
| Insert rows/cols (`sheets rows|cols insert`) | `POST /open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension` | tenant | v3 | no | `internal/larksdk/sheets.go: Client.InsertSheetRows` |
| Delete rows/cols (`sheets rows|cols delete`) | `DELETE /open-apis/sheets/v2/spreadsheets/:spreadsheet_token/dimension_range` | tenant | v2 | no | `internal/larksdk/sheets.go: Client.DeleteSheetRows` |

## Mail

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| List public mailboxes (`mail public-mailboxes list`) | `GET /open-apis/mail/v1/public_mailboxes` | tenant | v1 | yes |  |
| List messages (`mail list`) | `GET /open-apis/mail/v1/user_mailboxes/:mailbox_id/messages` | user | v1 | yes |  |
| Get message (`mail info`) | `GET /open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/:message_id` | user | v1 | yes |  |
| Send message (`mail send`) | `POST /open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/send` | user | v1 | yes |  |
| List folders (`mail folders`) | `GET /open-apis/mail/v1/user_mailboxes/:mailbox_id/folders` | user | v1 | no | `internal/larksdk/mail.go: Client.ListMailFolders` |
| Get mailbox (`mail mailbox info`) | `GET /open-apis/mail/v1/user_mailboxes/:user_mailbox_id` | user | v1 | no | `internal/larksdk/mail.go: Client.GetMailbox` |

## Wiki

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| List spaces (`wiki space list`) | `GET /open-apis/wiki/v2/spaces` | tenant/user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.ListWikiSpacesV2` |
| Get space (`wiki space info`) | `GET /open-apis/wiki/v2/spaces/:space_id` | tenant/user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.GetWikiSpaceV2` |
| Create space (`wiki space create`) | `POST /open-apis/wiki/v2/spaces` | user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.CreateWikiSpaceV2` |
| Update space settings (`wiki space update-setting`) | `PUT /open-apis/wiki/v2/spaces/:space_id/setting` | tenant/user | v2 | yes | `internal/larksdk/wiki_space_setting_v2.go: Client.UpdateWikiSpaceSettingV2` |
| List nodes (`wiki node list`) | `GET /open-apis/wiki/v2/spaces/:space_id/nodes` | tenant/user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.ListWikiNodesV2` |
| Get node (`wiki node info`) | `GET /open-apis/wiki/v2/spaces/get_node` | tenant/user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.GetWikiNodeV2` |
| Create node (`wiki node create`) | `POST /open-apis/wiki/v2/spaces/:space_id/nodes` | tenant/user | v2 | yes | `internal/larksdk/wiki_node_v2_extra.go: Client.CreateWikiNodeV2` |
| Move node (`wiki node move`) | `POST /open-apis/wiki/v2/spaces/:space_id/nodes/:node_token/move` | tenant/user | v2 | yes | `internal/larksdk/wiki_node_v2_extra.go: Client.MoveWikiNodeV2` |
| Update node title (`wiki node update-title`) | `POST /open-apis/wiki/v2/spaces/:space_id/nodes/:node_token/update_title` | tenant/user | v2 | yes | `internal/larksdk/wiki_node_v2_extra.go: Client.UpdateWikiNodeTitleV2` |
| Attach docs to wiki (`wiki node attach`) | `POST /open-apis/wiki/v2/spaces/:space_id/nodes/move_docs_to_wiki` | tenant/user | v2 | yes | `internal/larksdk/wiki_node_v2_extra.go: Client.MoveDocsToWikiV2` |
| Space members list/add/delete (`wiki member …`) | `/open-apis/wiki/v2/spaces/*/members*` | tenant/user | v2 | yes | `internal/larksdk/wiki_v2.go: Client.ListWikiSpaceMembersV2` |
| Task get (`wiki task info`) | `GET /open-apis/wiki/v2/tasks/:task_id` | tenant/user | v2 | yes | `internal/larksdk/wiki_task_v2.go: Client.GetWikiTaskV2` |
| Node search (`wiki node search`) | `POST /open-apis/wiki/v1/nodes/search` | user | v1 | no | `internal/larksdk/wiki_search_v1.go: Client.SearchWikiNodesV1` |

## Base (Bitable)

| Feature | Endpoint | Token | Ver | SDK? | Wrapper (if no SDK) |
|---|---|---:|:---:|:---:|---|
| App create (`base app create`) | `POST /open-apis/bitable/v1/apps` | tenant | v1 | no | `internal/larksdk/bitable_app.go: Client.CreateBitableApp` |
| App get (`base app info`) | `GET /open-apis/bitable/v1/apps/:app_token` | tenant | v1 | yes |  |
| Table list (`base table list`) | `GET /open-apis/bitable/v1/apps/:app_token/tables` | tenant | v1 | no | `internal/larksdk/base.go: Client.ListBaseTablesPage` |
| Table create/delete (`base table create/delete`) | `/open-apis/bitable/v1/apps/:app_token/tables/*` | tenant | v1 | yes |  |
| Field list (`base field list`) | `GET /open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields` | tenant | v1 | no | `internal/larksdk/base.go: Client.ListBaseFields` |
| Record create/update/delete (`base record create/update/delete`) | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records*` | tenant | v1 | yes |  |
| Record info (`base record info`) | `GET /open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id` | tenant | v1 | no | `internal/larksdk/base.go: Client.GetBaseRecord` |
| Record search (`base record search`) | `POST /open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/search` | tenant | v1 | no | `internal/larksdk/base.go: Client.SearchBaseRecords` |
