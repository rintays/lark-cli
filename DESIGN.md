# Design

This CLI follows a gog-inspired layout: a single root command wires shared state and subcommands, and the core behavior lives under internal packages.

## Structure

- cmd/lark: Cobra commands and wiring. The root command builds an appState and shared flags. Subcommands live beside it.
- internal/config: Load/save config (default path, JSON format, secure permissions).
- internal/larksdk: SDK client wrappers and core API requests.
- internal/output: Unified output handling (human text vs JSON).
- internal/testutil: Test HTTP helpers.

## Runtime flow

1. Root command resolves the config path, loads config, and builds the SDK client.
2. Commands call helper functions that enforce credentials and reuse cached tokens when valid.
3. Token refresh updates config on disk.

## CLI argument design (positional identifiers)

Goals:
- Make required identifiers discoverable and fast to type.
- Align CLI naming with OpenAPI field names to reduce confusion when cross-referencing docs.
- Avoid redundant required flags that duplicate positional args.

Rules:
- Required resource identifiers are positional arguments in `Use`.
- Search queries are positional: `search <query>`.
- Flags are reserved for optional filters, toggles, or when an identifier is optional.
- Positional args are trimmed and validated; errors should name the identifier (`<name> is required`).
- Help text must match positional naming (e.g., `document-id`, `spreadsheet-token`).

Naming alignment (CLI placeholder -> OpenAPI field):
- document-id -> document_id (Docs)
- file-token -> file_token (Drive/Docs/Sheets/Minutes exports)
- spreadsheet-token -> spreadsheet_token (Sheets)
- node-token -> node_token (Wiki)
- space-id -> space_id (Wiki, required flag for v2 member/node APIs)
- chat-id -> chat_id (IM)
- message-id -> message_id (IM/Mail)
- event-id -> event_id (Calendar)
- meeting-id -> meeting_id (Meetings)
- task-guid -> task_guid (Tasks)
- tasklist-guid -> tasklist_guid (Task lists)
- reserve-id -> reserve_id (Meetings reservations)
- minute-token -> minute_token (Minutes)
- app-token -> app_token (Bitable base)
- table-id -> table_id (Bitable)
- record-id -> record_id (Bitable)
- field-id -> field_id (Bitable)
- view-id -> view_id (Bitable)
- mailbox-id -> mailbox_id (Mail; optional, defaults to config or "me")

Examples:
- `lark docs get <document-id>`
- `lark drive info <file-token>`
- `lark sheets read <spreadsheet-token> <range>` (use `--sheet-id` to prefix ranges)
- `lark sheets rows insert <spreadsheet-token> <sheet-id> <start-index> <count>`
- `lark messages send <receive-id>` (use `--receive-id-type` for chat_id/open_id/user_id/email)
- `lark calendars search <query> [--start ... --end ...]`
- `lark calendars get <event-id>`

Exceptions:
- `base --app-token` remains a required flag (token is a config/scoping input, not a resource id).
- `wiki --space-id` remains a required flag (space context is shared across member/node subcommands).
- Optional identifiers remain flags (e.g., `--mailbox-id` for mail commands).

## Coverage Matrix

| Area | API | SDK Support | Token | Version | Notes |
| --- | --- | --- | --- | --- | --- |
| Auth tenant token | `/open-apis/auth/v3/tenant_access_token/internal` | SDK core | tenant | v3 | Used by `lark auth tenant` + token caching. |
| Auth user token refresh | `/open-apis/authen/v1/refresh_access_token` | SDK authen | app+user | v1 | Uses app access token + refresh token. |
| Whoami (tenant) | `/open-apis/tenant/v2/tenant/query` | Core ApiReq wrapper | tenant | v2 | `lark whoami`. |
| Whoami (user) | `/open-apis/authen/v1/user_info` | SDK authen | user | v1 | `lark --token-type user whoami`. |
| Chats list | `/open-apis/im/v1/chats` | SDK im | tenant | v1 | `lark chats list`. |
| Message send | `/open-apis/im/v1/messages` | SDK im | tenant | v1 | `lark messages send`. |
| Users info | `/open-apis/contact/v3/users/:user_id` | SDK contact | tenant | v3 | `lark users info`, `lark contacts user info`. |
| Users search | `/open-apis/search/v1/user` | Core ApiReq wrapper | user | v1 | `lark users search <search_query>`. |
| Drive list | `/open-apis/drive/v1/files` | SDK drive | tenant | v1 | `lark drive list`. |
| Drive search | `/open-apis/drive/v1/files/search` | Core ApiReq wrapper | tenant/user | v1 | `lark drive search`. |
| Drive metadata | `/open-apis/drive/v1/files/:file_token` | Core ApiReq wrapper | tenant/user | v1 | `lark drive info` / `lark drive urls`. |
| Drive upload | `/open-apis/drive/v1/files/upload_all` | Custom HTTP wrapper | tenant | v1 | Multipart upload. |
| Drive permissions | `/open-apis/drive/v1/permissions/:file_token/public` | Core ApiReq wrapper | tenant/user | v1 | `lark drive share`. |
| Drive permission members | `/open-apis/drive/v1/permissions/:token/members` | Core ApiReq wrapper | tenant/user | v1 | `lark drive permissions list/add`. |
| Drive permission member update | `/open-apis/drive/v1/permissions/:token/members/:member_id` | Core ApiReq wrapper | tenant/user | v1 | `lark drive permissions update`. |
| Drive permission member delete | `/open-apis/drive/v1/permissions/:token/members/:member_id` | Core ApiReq wrapper | tenant/user | v1 | `lark drive permissions delete`. |
| Drive export task | `/open-apis/drive/v1/export_tasks` | Core ApiReq wrapper | tenant/user | v1 | Used by docs/drive export. |
| Drive export status | `/open-apis/drive/v1/export_tasks/:ticket` | Core ApiReq wrapper | tenant/user | v1 | Used by docs/drive export. |
| Drive export download | `/open-apis/drive/v1/export_tasks/file/:file_token/download` | Custom HTTP wrapper | tenant | v1 | File download. |
| Docs create/info | `/open-apis/docx/v1/documents` | Core ApiReq wrapper | tenant/user | v1 | `lark docs create/info`. |
| Docs blocks get/list | `/open-apis/docx/v1/documents/:document_id/blocks` | SDK docx | tenant/user | v1 | `lark docs blocks get/list`. |
| Docs blocks update | `/open-apis/docx/v1/documents/:document_id/blocks/:block_id` | SDK docx | tenant/user | v1 | `lark docs blocks update`. |
| Docs blocks batch update | `/open-apis/docx/v1/documents/:document_id/blocks/batch_update` | SDK docx | tenant/user | v1 | `lark docs blocks batch-update`. |
| Docs block children | `/open-apis/docx/v1/documents/:document_id/blocks/:block_id/children` | SDK docx | tenant/user | v1 | `lark docs blocks children list/create/delete`. |
| Docs block descendant | `/open-apis/docx/v1/documents/:document_id/blocks/:block_id/descendant` | SDK docx | tenant/user | v1 | `lark docs blocks descendant create`. |
| Docs convert | `/open-apis/docx/v1/documents/blocks/convert` | SDK docx | tenant/user | v1 | `lark docs convert/overwrite`. |
| Sheets create | `/open-apis/sheets/v3/spreadsheets` | Core ApiReq wrapper | tenant/user | v3 | `lark sheets create`. |
| Sheets read | `/open-apis/sheets/v2/spreadsheets/:token/values/:range` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets read`. |
| Sheets update | `/open-apis/sheets/v2/spreadsheets/:token/values` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets update`. |
| Sheets append | `/open-apis/sheets/v2/spreadsheets/:token/values_append` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets append`. |
| Sheets clear | `/open-apis/sheets/v2/spreadsheets/:token/values_clear` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets clear`. |
| Sheets info | `/open-apis/sheets/v2/spreadsheets/:token/metainfo` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets info`. |
| Sheets delete | `/open-apis/drive/v1/files/:file_token` | SDK drive delete | tenant | v1 | `lark sheets delete` (type=sheet). |
| Calendar primary | `/open-apis/calendar/v4/calendars/primary` | Core ApiReq wrapper | tenant/user | v4 | `lark calendars list/create` (alias: `calendar`). |
| Calendar events | `/open-apis/calendar/v4/calendars/:id/events` | Core ApiReq wrapper | tenant/user | v4 | `lark calendars list/create` (alias: `calendar`). |
| Calendar attendees | `/open-apis/calendar/v4/calendars/:id/events/:event_id/attendees` | Core ApiReq wrapper | tenant/user | v4 | `lark calendars create` (alias: `calendar`). |
| Tasks list | `/open-apis/task/v2/tasks` | Core ApiReq wrapper | user | v2 | `lark tasks list` (my_tasks). |
| Tasks get | `/open-apis/task/v2/tasks/:task_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasks info`. |
| Tasks create | `/open-apis/task/v2/tasks` | Core ApiReq wrapper | tenant/user | v2 | `lark tasks create`. |
| Tasks update | `/open-apis/task/v2/tasks/:task_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasks update`. |
| Tasks delete | `/open-apis/task/v2/tasks/:task_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasks delete`. |
| Task lists create | `/open-apis/task/v2/tasklists` | Core ApiReq wrapper | tenant/user | v2 | `lark tasklists create`. |
| Task lists info | `/open-apis/task/v2/tasklists/:tasklist_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasklists info`. |
| Task lists update | `/open-apis/task/v2/tasklists/:tasklist_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasklists update`. |
| Task lists delete | `/open-apis/task/v2/tasklists/:tasklist_guid` | Core ApiReq wrapper | tenant/user | v2 | `lark tasklists delete`. |
| Meetings info | `/open-apis/vc/v1/meetings/:meeting_id` | Core ApiReq wrapper | tenant/user | v1 | `lark meetings info`. |
| Minutes info | `/open-apis/minutes/v1/minutes/:minute_token` | SDK minutes | tenant | v1 | `lark minutes info`. |
| Minutes list | `/open-apis/drive/v1/files` | SDK drive list (filter type=minutes) | tenant/user | v1 | `lark minutes list`. |
| Minutes delete | `/open-apis/drive/v1/files/:file_token` | SDK drive delete | tenant/user | v1 | `lark minutes delete`. |
| Minutes update | `/open-apis/drive/v1/permissions/:file_token/public` | SDK drive permissions | tenant/user | v1 | `lark minutes update`. |
| Mail folders | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/folders` | Core ApiReq wrapper | tenant/user | v1 | `lark mail folders`. |
| Mail list | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages` | Core ApiReq wrapper | tenant/user | v1 | `lark mail list`. |
| Mail info (metadata) | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/:message_id` | Core ApiReq wrapper | tenant/user | v1 | `lark mail info`. |
| Mail get (content) | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/:message_id` | Core ApiReq wrapper | tenant/user | v1 | `lark mail get`. |
| Mail send | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/send` | Core ApiReq wrapper | user | v1 | `lark mail send`. |

## Config + caching

- Default config path: ~/.config/lark/config.json
- Token caching is stored in config and reused until close to expiry.
- Commands are written to accept an alternate config path via the global --config flag.
