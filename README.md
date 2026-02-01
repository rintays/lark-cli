# lark — Feishu/Lark in your terminal

Fast, script-friendly CLI for **Feishu (飞书)** / **Lark**.

- **JSON-first output** (`--json`) for scripting
- Consistent command layout (top-level product areas → subcommands)
- **SDK-first** implementation using the official `oapi-sdk-go` (with `core.ApiReq` for gaps)

> Status: actively developed. See “Not implemented yet / TODO” at the bottom.

---

## Features (today)

- **Auth**
  - Tenant token fetch + caching
  - Config file support + env fallback
- **Users / Chats / Messages (IM)**
  - search users
  - list/create/get/update chats
  - get/update chat announcements
  - send/reply messages (text/post/image/file/media)
  - list/search messages
  - add/delete reactions, pin/unpin messages
- **Drive**
  - list/search/info/urls/download/upload
  - share permission updates
- **Docs (docx)**
  - create/info/export/get
- **Sheets**
  - read/update/append/clear/info/delete/list/search
- **Calendar**
  - list/create events
- **Contacts**
  - basic user lookup
- **Meetings / Minutes**
  - meeting list/get + reservation create/update/delete

---

## Installation

### Build from source

```bash
git clone https://github.com/rintays/lark.git
cd lark
go build -o lark ./cmd/lark

./lark --help
./lark chats list --help
./lark users list --help
./lark users info --help
./lark messages send --help  # alias: msg
./lark calendars --help  # alias: calendar
```

---

## Quick start

### 1) Configure app credentials

Store credentials in config (default: `~/.config/lark/config.json`, or `~/.config/lark/profiles/<profile>/config.json` with `--profile`/`LARK_PROFILE`):

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET>
```

Or (equivalent):

```bash
lark config set --app-id <APP_ID> --app-secret <APP_SECRET>
```

Set the default platform base URL (optional):

```bash
lark auth platform set feishu|lark
lark auth platform info
```

Or set env vars:

```bash
# App credentials: used only when config is empty (config wins).
export LARK_APP_ID=<APP_ID>
export LARK_APP_SECRET=<APP_SECRET>

# Optional profile selection.
export LARK_PROFILE=<profile>

# Token storage backend: env wins over config.keyring_backend.
# auto prefers keychain on macOS/Windows; otherwise falls back to file.
export LARK_KEYRING_BACKEND=file  # or: keychain|auto
```

View the currently loaded config:

```bash
lark config info
```

Set the base URL directly (optional):

```bash
lark config set --base-url https://open.feishu.cn
```

Set the platform base URL (optional):

```bash
lark config set --platform feishu|lark
```

Clear the persisted base URL:

```bash
lark config unset --base-url
```

Clear the default mailbox id:

```bash
lark config unset --default-mailbox-id
```

Clear user access tokens:

```bash
lark config unset --user-tokens
```
Clears all stored user OAuth tokens (file or keychain).

### 2) Get tenant token

```bash
lark auth
```

### 3) Try basic commands

```bash
lark whoami
lark chats list --limit 10
lark users search --email user@example.com
lark messages send --receive-id <CHAT_ID> --text "hello"
```

---

## Output modes

- Default: human-friendly text
- `--json`: machine-readable JSON (recommended for scripts)

Examples:

```bash
lark chats list --json
lark users search --email user@example.com --json
```

---

## Global flags

- `--config <path>`: override config path
- `--profile <name>`: use a named config profile (env: `LARK_PROFILE`)
- `--json`: JSON output
- `--verbose`: verbose output
- `--platform feishu|lark`: runtime base URL selection (not saved)
- `--base-url <url>`: runtime base URL override (not saved; wins over `--platform`)

Precedence:
`--base-url` > `--platform` > `config.base_url` > default (`https://open.feishu.cn`).

---

## Common recipes (examples)

### Send a message

```bash
lark messages send --receive-id <CHAT_ID> --text "hello"
```

Send to a user by email:

```bash
lark messages send --receive-id-type email --receive-id user@example.com --text "hello"
```

Send a post (rich text):

```bash
lark messages send --receive-id <CHAT_ID> --msg-type post --content '{"zh_cn":{"content":[[{"tag":"text","text":"hello"}]]}}'
```

Send an image:

```bash
lark messages send --receive-id <CHAT_ID> --image-key <IMAGE_KEY>
```

Reply in thread:

```bash
lark messages reply <MESSAGE_ID> --text "reply" --reply-in-thread
```

Search messages (user token required):

```bash
lark messages search "hello" --chat-id <CHAT_ID>
```

Search results include message metadata and content in the default output.

List recent messages:

```bash
lark messages list <CHAT_ID> --limit 20
```

Add a reaction:

```bash
lark messages reactions add <MESSAGE_ID> SMILE
```

Pin a message:

```bash
lark messages pin <MESSAGE_ID>
```

### Chats

Create a chat:

```bash
lark chats create --name "Demo Chat"
```

Get chat info:

```bash
lark chats get <CHAT_ID>
```
By default this includes a member preview; adjust or disable it with:

```bash
lark chats get <CHAT_ID> --members-limit 50
lark chats get <CHAT_ID> --members-limit 0
```

Update chat info:

```bash
lark chats update <CHAT_ID> --name "New Name"
```

Get chat announcement:

```bash
lark chats announcement get <CHAT_ID>
```

Update chat announcement:

```bash
lark chats announcement update --chat-id <CHAT_ID> --revision 12 --request '<REQUEST_JSON>'
```

### Drive

List files:

```bash
lark drive list --folder-id <FOLDER_TOKEN> --limit 20
```

Search files:

```bash
lark drive search "budget" --limit 10 --type sheet --type docx
```

Drive search uses a **user access token**. Make sure your app has `drive:drive`, `drive:drive:readonly`, or `search:docs:read` user scopes, then run `lark auth user login` to refresh user authorization.

Download:

```bash
lark drive download <FILE_TOKEN> --out ./downloaded.bin
```

Upload:

```bash
lark drive upload ./report.pdf --folder-token <FOLDER_TOKEN> --name "report.pdf"
```

Update share:

```bash
lark drive share <FILE_TOKEN> --type docx --link-share tenant_readable --external-access
```

### Docs (docx)

List:

```bash
lark docs list --folder-id <FOLDER_TOKEN> --limit 50
```

Create:

```bash
lark docs create "Weekly Update" --folder-id <FOLDER_TOKEN>
```

Export:

```bash
lark docs export <DOCUMENT_ID> --format pdf --out ./document.pdf
```

Get:

```bash
lark docs get <DOCUMENT_ID>

# or blocks
lark docs get <DOCUMENT_ID> --format blocks
```

Blocks:

```bash
lark docs blocks list <DOCUMENT_ID> --limit 50
lark docs blocks get <DOCUMENT_ID> <BLOCK_ID>
lark docs blocks update <DOCUMENT_ID> <BLOCK_ID> --body-json '<UPDATE_REQUEST_JSON>'
```

Convert/Overwrite:

```bash
lark docs convert --content "# Title"
lark docs overwrite <DOCUMENT_ID> --content-file ./doc.md
```

### Sheets

List:

```bash
lark sheets list --folder-id <FOLDER_TOKEN> --limit 50
```

Create:

```bash
lark sheets create --title "Budget Q1" --folder-id <FOLDER_TOKEN>
```

Read:

```bash
lark sheets read <SPREADSHEET_TOKEN> "Sheet1!A1:B2"
# or
lark sheets read <SPREADSHEET_TOKEN> A1:B2 --sheet-id <SHEET_ID>
```

Search:

```bash
lark sheets search <TEXT> --limit 50 # requires user_access_token or `lark auth user login`
```

Update:

```bash
lark sheets update <SPREADSHEET_TOKEN> "Sheet1!A1:B2" --values '[["Name","Amount"],["Ada",42]]'
# or
lark sheets update <SPREADSHEET_TOKEN> A1:B2 --sheet-id <SHEET_ID> --values-file ./values.csv
```

Append:

```bash
lark sheets append <SPREADSHEET_TOKEN> "Sheet1!A1:B2" --values '[["Name","Amount"],["Ada",42]]' --insert-data-option INSERT_ROWS
# or
lark sheets append <SPREADSHEET_TOKEN> A1:B2 --sheet-id <SHEET_ID> --values @./values.json
```

Clear:

```bash
lark sheets clear <SPREADSHEET_TOKEN> "Sheet1!A1:B2"
# or
lark sheets clear <SPREADSHEET_TOKEN> A1:B2 --sheet-id <SHEET_ID>
```

Info:

```bash
lark sheets info <SPREADSHEET_TOKEN>
```

Delete:

```bash
lark sheets delete <SPREADSHEET_TOKEN>
```

Insert rows:

```bash
lark sheets rows insert <SPREADSHEET_TOKEN> <SHEET_ID> 1 2
```

Delete rows:

```bash
lark sheets rows delete <SPREADSHEET_TOKEN> <SHEET_ID> 1 2
```

Insert cols:

```bash
lark sheets cols insert <SPREADSHEET_TOKEN> <SHEET_ID> 1 2
```

Delete cols:

```bash
lark sheets cols delete <SPREADSHEET_TOKEN> <SHEET_ID> 1 2
```

### Calendar

List events:

```bash
lark calendars list --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --limit 20
```

Create event:

```bash
lark calendars create --summary "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --attendee dev@example.com
```

Search events:

```bash
lark calendars search --query "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --limit 20
```

Get event:

```bash
lark calendars get --event-id <EVENT_ID>
```

Update event:

```bash
lark calendars update --event-id <EVENT_ID> --summary "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z"
```

Delete event:

```bash
lark calendars delete --event-id <EVENT_ID> --notify=false
```

---

## User OAuth scopes (important)

Manage default user OAuth scopes:

```bash
lark auth user scopes list
lark auth user scopes set offline_access drive:drive:readonly
lark auth user scopes add drive:drive
lark auth user scopes remove drive:drive:readonly
```

Log in with explicit scopes:

```bash
lark auth user login --scopes "offline_access drive:drive:readonly" --force-consent
```

By default, `auth user login` uses incremental authorization (requests only new scopes). Disable with `--incremental=false` to request the full scope set.

Service-style scopes (gog-like):

```bash
lark auth user services
lark auth user login --services drive --drive-scope readonly --force-consent
lark auth user login --services drive --drive-scope full --force-consent
```

Read-only shortcut:

```bash
lark auth user login --readonly --force-consent
```

Manage user OAuth accounts:

```bash
lark auth user accounts list
lark auth user accounts set <ACCOUNT>
lark auth user accounts remove <ACCOUNT>
```

Set default via config:

```bash
lark config set --default-user-account <ACCOUNT>
```

Select an account per command:

```bash
lark --account <ACCOUNT> auth user status
```

Environment override: `LARK_ACCOUNT`.

Token storage backend: `keyring_backend=file|keychain|auto` (config).

- `file`: store user OAuth tokens in `config.json`.
- `keychain`: store user OAuth tokens in the OS keychain (via go-keyring).
- `auto`: prefer `keychain` on macOS/Windows; otherwise fall back to `file`.

Explain auth requirements (services → token types/scopes) for a command:

```bash
lark auth explain drive search
lark auth explain --readonly drive search
lark auth explain mail send
```

---

## Mail: user OAuth token (important)

Some Mail actions (notably **`mail send`**) require a **user access token** (OAuth), not a tenant token.

Current behavior:
- Run `lark auth user login` to launch OAuth and store tokens locally (add `--force-consent` if you need to re-grant scopes / refresh token)
- Provide via `--user-access-token <token>`
- or env `LARK_USER_ACCESS_TOKEN`
- Mail commands `mail folders/list/info/get/send` default `--mailbox-id` to `config.default_mailbox_id` or `me`
- Set a default with `lark config set --default-mailbox-id <id|me>` or `lark mail mailbox set --mailbox-id <id>`
- `mail info` shows metadata; `mail get` returns full message content (raw/body/attachments)

Example:

```bash
./lark auth user login --help
./lark mail public-mailboxes list --help
./lark bases app create --help
./lark bases app copy --help
./lark bases app info --help
./lark bases app update --help
./lark bases table list --help  # alias: base
./lark bases table create --help
./lark bases field list --help
./lark bases field create --help
./lark bases field update --help
./lark bases field types --help
./lark bases view list --help
./lark bases view create --help
./lark bases view delete --help
./lark bases view info --help
./lark bases record create --help
./lark bases record batch-create --help
./lark bases record batch-delete --help
./lark bases record batch-update --help
./lark bases record info --help
./lark bases record search --help
./lark bases record update --help
./lark bases record delete --help
./lark wiki member list --help
./lark wiki member delete --help
./lark wiki node search --help
./lark wiki task info --help
./lark mail mailbox info --help
./lark mail mailbox set --mailbox-id <MAILBOX_ID>
./lark mail info <MESSAGE_ID>
./lark mail get <MESSAGE_ID>
./lark mail send --subject "Hello" --to "user@example.com" --text "Hi there"
./lark mail send --raw-file ./message.eml
```

---

## Bitable (Base) concepts

Bitable is Lark/Feishu's database product. In the API, a **base** is also called an **app**.

- **App/Base:** the top-level container; identified by an app token.
- **Table:** a grid inside the base; defines fields (columns) and stores records (rows).
- **Field:** a column definition (type + name) used by every record in the table.
- **Record:** a single row of data for the table's fields.
- **View:** a saved presentation of a table (filters/sorts/grouping/hidden columns); it doesn't change the underlying records.

Relationships: app/base → tables → fields + records; views belong to a table.

---


## Token selection (tenant vs user)

Many OpenAPI endpoints accept **tenant** or **user** access tokens. You can control which token type the CLI uses:

- Per command: `--token-type tenant|user|auto`
- Default preference: `lark config set --default-token-type tenant|user`

Behavior:
- If an API supports **only one** token type, the CLI uses it automatically and errors if you explicitly request the other.
- If an API supports **both**, `--token-type=auto` uses `config.default_token_type` (default: `tenant`).
- When `user` is selected and no user token is available, the CLI guides you to run `lark auth user login` with recommended `--scopes` (derived from the command→service registry).
- When using user tokens, the selected account comes from `--account`, `LARK_ACCOUNT`, or `config.default_user_account`.

---


## Integration tests

Integration tests run only when explicitly enabled (real network + credentials). `LARK_INTEGRATION` must be exactly `1`:

```bash
export LARK_INTEGRATION=1
```

Recommended (all integration tests):

```bash
make it
```

Run the Wiki SpaceMember.Create role-upsert verification test:

```bash
export LARK_TEST_WIKI_SPACE_ID=<space_id>
export LARK_TEST_USER_EMAIL=<member_email>

go test ./cmd/lark -run '^TestWikiMemberRoleUpdateIntegration$' -count=1 -v
# or:
make it-wiki-member-role-update
```

Prereqs: app creds configured (`lark auth login ...` or `LARK_APP_ID/LARK_APP_SECRET`) and cached user token (`lark auth user login`).

---

## Not implemented yet / TODO (from backlog)

This README is written in the style of “what the CLI will look like once the backlog is complete”.
Items not finished yet (high-level):

- **Mail UX:** use configured default mailbox for mail commands + additional mailbox management commands
- **Sheets:** row/col insert/delete commands
- **Base (Bitable):** `bases` top-level command tree (records CRUD, tables/fields/views) (alias: `base`)
- **Wiki:** v2 SDK endpoints (v1 node search is available via `wiki node search`)
- **Integration tests:** `*_integration_test.go` suite gated by `LARK_INTEGRATION=1`

For the full detailed task breakdown, see:
- `<workspace>/lark/BACKLOG.md`
