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
```

---

## Quick start

### 1) Configure app credentials

Store credentials in config (default: `~/.config/lark/config.json`):

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET>
```

Or set env vars (used only when config is empty; config wins):

```bash
export LARK_APP_ID=<APP_ID>
export LARK_APP_SECRET=<APP_SECRET>
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
lark msg send --chat-id <CHAT_ID> --text "hello"
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

---

## Common recipes (examples)

### Send a message

```bash
lark msg send --chat-id <CHAT_ID> --text "hello"
```

Send to a user by email:

```bash
lark msg send --receive-id-type email --receive-id user@example.com --text "hello"
```

### Drive

List files:

```bash
lark drive list --folder-id <FOLDER_TOKEN> --limit 20
```

Search files:

```bash
lark drive search --query "budget" --limit 10
```

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

Read:

```bash
lark sheets read --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2"
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

## Mail: user OAuth token (important)

Some Mail actions (notably **`mail send`**) require a **user access token** (OAuth), not a tenant token.

Current behavior:
- Run `lark auth user login` to launch OAuth and store tokens locally
- Provide via `--user-access-token <token>`
- or env `LARK_USER_ACCESS_TOKEN`

Example:

```bash
./lark auth user login --help
```

---

## Integration testing (planned)

Goal: full end-to-end integration tests that validate:
- flags behave as expected
- `--json` output parses
- real API requests (including writes) succeed with your test app

Planned switch:
- `LARK_INTEGRATION=1 go test ./...`

---

## Not implemented yet / TODO (from backlog)

This README is written in the style of “what the CLI will look like once the backlog is complete”.
Items not finished yet (high-level):

- **Mail UX:** default mailbox selection + mailbox management commands
- **Sheets:** row/col insert/delete commands
- **Base (Bitable):** `base` top-level command tree (records CRUD, tables/fields/views)
- **Wiki:** `wiki` command tree (v2 SDK endpoints + v1 search via `core.ApiReq`)
- **Platform switching convenience:** `lark auth login --platform feishu|lark` (keep `--base-url` override)
- **Integration tests:** `*_integration_test.go` suite gated by `LARK_INTEGRATION=1`

For the full detailed task breakdown, see:
- `/Users/fredliang/clawd/BACKLOG.md`
