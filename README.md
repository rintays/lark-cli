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
- **Users / Chats / Msg (IM)**
  - search users
  - list chats
  - send messages (supports `--receive-id-type`)
- **Drive**
  - list/search/get/urls/download/upload
  - share permission updates
- **Docs (docx)**
  - create/get/export/cat
- **Sheets**
  - read/update/append/clear/metadata
- **Calendar**
  - list/create events
- **Contacts**
  - basic user lookup
- **Meetings / Minutes**
  - baseline read operations (see command help)

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
./lark users get --help
./lark msg send --help
```

---

## Quick start

### 1) Configure app credentials

Store credentials in config (default: `~/.config/lark/config.json`):

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
lark auth platform get
```

Or set env vars (used only when config is empty; config wins):

```bash
export LARK_APP_ID=<APP_ID>
export LARK_APP_SECRET=<APP_SECRET>
```

View the currently loaded config:

```bash
lark config get
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

### 2) Get tenant token

```bash
lark auth
```

### 3) Try basic commands

```bash
lark whoami
lark chats list --limit 10
lark users search --email user@example.com
lark messages send --chat-id <CHAT_ID> --text "hello"
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
lark messages send --chat-id <CHAT_ID> --text "hello"
```

Send to a user by email:

```bash
lark messages send --receive-id-type email --receive-id user@example.com --text "hello"
```

### Drive

List files:

```bash
lark drive list --folder-id <FOLDER_TOKEN> --limit 20
```

Search files:

```bash
lark drive search --query "budget" --limit 10 --type sheet --type docx
```

Drive search uses a **user access token**. Make sure your app has `drive:drive`, `drive:drive:readonly`, or `search:docs:read` user scopes, then run `lark auth user login` to refresh user authorization.

Download:

```bash
lark drive download --file-token <FILE_TOKEN> --out ./downloaded.bin
```

Upload:

```bash
lark drive upload --file ./report.pdf --folder-token <FOLDER_TOKEN> --name "report.pdf"
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
lark docs create --title "Weekly Update" --folder-id <FOLDER_TOKEN>
```

Export:

```bash
lark docs export --doc-id <DOCUMENT_ID> --format pdf --out ./document.pdf
```

Cat:

```bash
lark docs cat --doc-id <DOCUMENT_ID> --format txt
```

### Sheets

List:

```bash
lark sheets list --folder-id <FOLDER_TOKEN> --limit 50
```

Read:

```bash
lark sheets read --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2"
```

Search:

```bash
lark sheets search --query <TEXT> --limit 50
```

Update:

```bash
lark sheets update --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2" --values '[["Name","Amount"],["Ada",42]]'
```

Append:

```bash
lark sheets append --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2" --values '[["Name","Amount"],["Ada",42]]' --insert-data-option INSERT_ROWS
```

Clear:

```bash
lark sheets clear --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2"
```

Metadata:

```bash
lark sheets metadata --spreadsheet-id <SPREADSHEET_TOKEN>
```

Insert rows:

```bash
lark sheets rows insert --spreadsheet-id <SPREADSHEET_TOKEN> --sheet-id <SHEET_ID> --start-index 1 --count 2
```

Delete rows:

```bash
lark sheets rows delete --spreadsheet-id <SPREADSHEET_TOKEN> --sheet-id <SHEET_ID> --start-index 1 --count 2
```

Insert cols:

```bash
lark sheets cols insert --spreadsheet-id <SPREADSHEET_TOKEN> --sheet-id <SHEET_ID> --start-index 1 --count 2
```

Delete cols:

```bash
lark sheets cols delete --spreadsheet-id <SPREADSHEET_TOKEN> --sheet-id <SHEET_ID> --start-index 1 --count 2
```

### Calendar

List events:

```bash
lark calendar list --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --limit 20
```

Create event:

```bash
lark calendar create --summary "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --attendee dev@example.com
```

---

## User OAuth scopes (important)

Manage default user OAuth scopes:

```bash
lark auth user scopes list
lark auth user scopes set --scopes "offline_access drive:drive:readonly"
lark auth user scopes add --scopes "drive:drive"
lark auth user scopes remove --scopes "drive:drive:readonly"
```

Log in with explicit scopes:

```bash
lark auth user login --scopes "offline_access drive:drive:readonly" --force-consent
```

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
- Mail commands `mail folders/list/get/send` default `--mailbox-id` to `config.default_mailbox_id` or `me`
- Set a default with `lark config set --default-mailbox-id <id|me>` or `lark mail mailbox set --mailbox-id <id>`

Example:

```bash
./lark auth user login --help
./lark mail public-mailboxes list --help
./lark base table list --help
./lark base field list --help
./lark base view list --help
./lark base record create --help
./lark base record get --help
./lark base record search --help
./lark base record update --help
./lark base record delete --help
./lark wiki member list --help
./lark wiki member delete --help
./lark wiki node search --help
./lark wiki task get --help # alias: wiki task list
./lark mail mailbox get --help
./lark mail mailbox set --mailbox-id <MAILBOX_ID>
./lark mail send --subject "Hello" --to "user@example.com" --text "Hi there"
```

---


## Token selection (tenant vs user)

Many OpenAPI endpoints accept **tenant** or **user** access tokens. You can control which token type the CLI uses:

- Per command: `--token-type tenant|user|auto`
- Default preference: `lark config set --default-token-type tenant|user`

Behavior:
- If an API supports **only one** token type, the CLI uses it automatically and errors if you explicitly request the other.
- If an API supports **both**, `--token-type=auto` uses `config.default_token_type` (default: `tenant`).
- When `user` is selected and no user token is available, the CLI guides you to run `lark auth user login` with recommended `--scopes` (derived from the command→service registry).

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
- **Base (Bitable):** `base` top-level command tree (records CRUD, tables/fields/views)
- **Wiki:** v2 SDK endpoints (v1 node search is available via `wiki node search`)
- **Integration tests:** `*_integration_test.go` suite gated by `LARK_INTEGRATION=1`

For the full detailed task breakdown, see:
- `<workspace>/lark/BACKLOG.md`
