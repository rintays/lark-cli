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

## Coverage Matrix

| Area | API | SDK Support | Token | Version | Notes |
| --- | --- | --- | --- | --- | --- |
| Auth tenant token | `/open-apis/auth/v3/tenant_access_token/internal` | SDK core | tenant | v3 | Used by `lark auth tenant` + token caching. |
| Auth user token refresh | `/open-apis/authen/v1/refresh_access_token` | SDK authen | app+user | v1 | Uses app access token + refresh token. |
| Whoami | `/open-apis/tenant/v2/tenant/query` | Core ApiReq wrapper | tenant | v2 | `lark whoami`. |
| Chats list | `/open-apis/im/v1/chats` | SDK im | tenant | v1 | `lark chats list`. |
| Message send | `/open-apis/im/v1/messages` | SDK im | tenant | v1 | `lark messages send`. |
| Users info | `/open-apis/contact/v3/users/:user_id` | SDK contact | tenant | v3 | `lark users info`, `lark contacts user info`. |
| Users search (batch IDs) | `/open-apis/contact/v3/users/batch_get_id` | SDK contact | tenant | v3 | `lark users search --email/--mobile`. |
| Users search (dept) | `/open-apis/contact/v3/users/find_by_department` | SDK contact | tenant | v3 | `lark users search --name`. |
| Drive list | `/open-apis/drive/v1/files` | SDK drive | tenant | v1 | `lark drive list`. |
| Drive search | `/open-apis/drive/v1/files/search` | Core ApiReq wrapper | tenant/user | v1 | `lark drive search`. |
| Drive metadata | `/open-apis/drive/v1/files/:file_token` | Core ApiReq wrapper | tenant/user | v1 | `lark drive info` / `lark drive urls`. |
| Drive upload | `/open-apis/drive/v1/files/upload_all` | Custom HTTP wrapper | tenant | v1 | Multipart upload. |
| Drive permissions | `/open-apis/drive/v1/permissions/:file_token/public` | Core ApiReq wrapper | tenant/user | v1 | `lark drive share`. |
| Drive export task | `/open-apis/drive/v1/export_tasks` | Core ApiReq wrapper | tenant/user | v1 | Used by docs/drive export. |
| Drive export status | `/open-apis/drive/v1/export_tasks/:ticket` | Core ApiReq wrapper | tenant/user | v1 | Used by docs/drive export. |
| Drive export download | `/open-apis/drive/v1/export_tasks/file/:file_token/download` | Custom HTTP wrapper | tenant | v1 | File download. |
| Docs create/info | `/open-apis/docx/v1/documents` | Core ApiReq wrapper | tenant/user | v1 | `lark docs create/info`. |
| Sheets create | `/open-apis/sheets/v3/spreadsheets` | Core ApiReq wrapper | tenant/user | v3 | `lark sheets create`. |
| Sheets read | `/open-apis/sheets/v2/spreadsheets/:token/values/:range` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets read`. |
| Sheets update | `/open-apis/sheets/v2/spreadsheets/:token/values` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets update`. |
| Sheets append | `/open-apis/sheets/v2/spreadsheets/:token/values_append` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets append`. |
| Sheets clear | `/open-apis/sheets/v2/spreadsheets/:token/values_clear` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets clear`. |
| Sheets info | `/open-apis/sheets/v2/spreadsheets/:token/metainfo` | Core ApiReq wrapper | tenant/user | v2 | `lark sheets info`. |
| Calendar primary | `/open-apis/calendar/v4/calendars/primary` | Core ApiReq wrapper | tenant/user | v4 | `lark calendar list/create`. |
| Calendar events | `/open-apis/calendar/v4/calendars/:id/events` | Core ApiReq wrapper | tenant/user | v4 | `lark calendar list/create`. |
| Calendar attendees | `/open-apis/calendar/v4/calendars/:id/events/:event_id/attendees` | Core ApiReq wrapper | tenant/user | v4 | `lark calendar create`. |
| Meetings info | `/open-apis/vc/v1/meetings/:meeting_id` | Core ApiReq wrapper | tenant/user | v1 | `lark meetings info`. |
| Minutes info | `/open-apis/minutes/v1/minutes/:minute_token` | SDK minutes | tenant | v1 | `lark minutes info`. |
| Minutes list | `/open-apis/minutes/v1/minutes` | Core ApiReq wrapper | tenant/user | v1 | `lark minutes list`. |
| Mail folders | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/folders` | Core ApiReq wrapper | tenant/user | v1 | `lark mail folders`. |
| Mail list | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages` | Core ApiReq wrapper | tenant/user | v1 | `lark mail list`. |
| Mail info | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/:message_id` | Core ApiReq wrapper | tenant/user | v1 | `lark mail info`. |
| Mail send | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/messages/send` | Core ApiReq wrapper | user | v1 | `lark mail send`. |

## Config + caching

- Default config path: ~/.config/lark/config.json
- Token caching is stored in config and reused until close to expiry.
- Commands are written to accept an alternate config path via the global --config flag.
