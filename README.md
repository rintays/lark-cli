# lark — Feishu/Lark in your terminal

One CLI for **Feishu (飞书)** / **Lark**: IM, Drive, Docs, Sheets, Mail, Calendar, Wiki, Bitable, Tasks — with JSON output and sane defaults.

- **JSON output** (`--json`) for scripting
- **Tenant + user OAuth tokens**, multi-account support
- Consistent command layout (product → subcommand → action)
- **SDK-first** implementation using the official `oapi-sdk-go`

> Status: actively developed.

---

## Install

### Homebrew

```bash
brew tap rintays/tap
brew install rintays/tap/lark
```

### Build from source

```bash
git clone https://github.com/rintays/lark.git
cd lark
go install ./cmd/lark
lark --help
lark chats --help
lark messages --help

# or build a local binary:
go build -o lark ./cmd/lark
./lark --help
./lark chats --help
./lark messages --help
```

### Download from GitHub Releases

Download the archive for your OS from `https://github.com/rintays/lark/releases`, extract it, and move `lark` to your PATH.

macOS/Linux example:

```bash
curl -L https://github.com/rintays/lark/releases/latest/download/lark_<VERSION>_darwin_arm64.tar.gz -o lark.tar.gz
tar -xzf lark.tar.gz
chmod +x lark
sudo mv lark /usr/local/bin/lark
```

Windows (PowerShell) example:

```powershell
Invoke-WebRequest -Uri https://github.com/rintays/lark/releases/latest/download/lark_<VERSION>_windows_amd64.zip -OutFile lark.zip
Expand-Archive lark.zip -DestinationPath .
Move-Item .\lark.exe $env:USERPROFILE\bin\lark.exe
```

---

## Quickstart

### 1) Store app credentials

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET>
```

Store app secret in OS keychain (optional):

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET> --store-secret-in-keyring
```

### 2) Get tokens

Tenant token (app-only APIs):

```bash
lark auth tenant
```

User token (user-scoped APIs like Drive search, Mail send):

```bash
lark auth user login
```

### 3) Run commands

```bash
lark whoami
lark chats list --limit 10
lark users search "Ada" --json | jq
lark messages send <CHAT_ID> --text "hello"
```

---

## Output

- Default: human-friendly tables/text
- `--json`: machine-readable JSON (recommended for scripts)
- Data is printed to stdout; logs/errors go to stderr

Examples:

```bash
lark chats list --json
lark users search "Ada" --json
```

---

## I/O shortcuts

- Many commands accept a Lark/Feishu web URL anywhere a token/id is expected (docx/sheet/file).
- File input flags (e.g., `--content-file`, `--raw-file`) accept `-` to read from stdin.
- Export/download output flags accept `--out -` to stream to stdout (binary).

---

## Command discovery

```bash
lark --help
lark chats --help
lark chats create --help
```

---

## Auth, accounts, secrets

Config path (default): `~/.config/lark/config.json`.

Global selection:

- `--token-type tenant|user|auto`
- `--account <ACCOUNT>` (or `LARK_ACCOUNT`)
- `--profile <name>` (or `LARK_PROFILE`)

Keychain & secrets:

- Store user OAuth tokens in OS keychain via `keyring_backend=keychain` (config) or `LARK_KEYRING_BACKEND=keychain`.
- Store app secrets in the keychain via `--store-secret-in-keyring` (auth login/config set).

Platform/base URL:

```bash
lark auth platform set feishu|lark
lark auth platform info

lark config set --base-url https://open.feishu.cn
lark config unset --base-url
```

Token selection behavior:

- If an API supports only one token type, the CLI uses it automatically.
- If an API supports both, `--token-type=auto` uses `config.default_token_type` (default: `tenant`).
- When `user` is selected and no user token is available, the CLI guides you to run `lark auth user login` with the recommended scopes.

---

## Examples (common workflows)

Send a message:

```bash
lark messages send <CHAT_ID> --text "hello"
```

Search messages (user token required):

```bash
lark messages search "hello" --chat-id <CHAT_ID>
```

Drive search + download:

```bash
lark drive search "budget" --limit 10 --type sheet --type docx
lark drive download <FILE_TOKEN> --out ./downloaded.bin
```

Docs get (Markdown from blocks):

```bash
lark docs get <DOCUMENT_ID> --format md
```

Sheets read + update:

```bash
lark sheets read <SPREADSHEET_TOKEN> "Sheet1!A1:B2"
lark sheets update <SPREADSHEET_TOKEN> "Sheet1!A1:B2" --values '[ ["Name","Amount"], ["Ada",42] ]'
```

Mail send (user token required):

```bash
lark mail send --subject "Hello" --to "user@example.com" --text "Hi there"
```

Calendar create:

```bash
lark calendars create --summary "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z"
```

Wiki node tree:

```bash
lark wiki node tree --space-id <SPACE_ID>
```

Bitable record create:

```bash
lark bases record create <TABLE_ID> --app-token <APP_TOKEN> --field "Name=Ada" --field "Score=42"
```

---

## Features

- **Auth/Config**: tenant token + user OAuth, profiles, keychain support, platform/base URL
- **Users/Contacts**: search users, basic user lookup
- **Chats/Messages (IM)**: list/create/get/update chats, announcements, send/reply/search/list messages, reactions, pin/unpin
- **Drive**: list/search/info/urls/download/upload, permissions add/list/update/delete
- **Docs (docx)**: create/info/export/get, blocks list/get/update/batch/children/descendant, convert/overwrite
- **Sheets**: create/read/update/append/clear/info/delete/list/search, rows/cols insert/delete
- **Calendar**: list/search/get/create/update/delete events
- **Mail**: list/info/get/send, folders/mailbox management, public mailboxes
- **Meetings/Minutes**: list/get + reservation create/update/delete, minutes update/delete
- **Tasks**: task lists + tasks CRUD
- **Wiki**: space create/update-setting, node create/move/update-title/attach/tree/search
- **Bitable (Base)**: apps/tables/fields/views/records

---

## User OAuth scopes (prefer services)

Service-style authorization (recommended):

```bash
lark auth user services
lark auth user login --services drive --drive-scope readonly --force-consent
```

Log in with explicit scopes (when you need fine-grained control):

```bash
lark auth user login --scopes "offline_access drive:drive:readonly" --force-consent
```

Manage default user OAuth scopes:

```bash
lark auth user scopes list
lark auth user scopes set offline_access drive:drive:readonly
lark auth user scopes add drive:drive
lark auth user scopes remove drive:drive:readonly
```

Explain auth requirements for a command:

```bash
lark auth explain drive search
lark auth explain --readonly drive search
lark auth explain mail send
```

---

## Mail notes (user OAuth token)

Some Mail actions (notably **`mail send`**) require a **user access token** (OAuth), not a tenant token.

- Run `lark auth user login` to launch OAuth and store tokens locally.
- Provide a token via `--user-access-token <token>` or env `LARK_USER_ACCESS_TOKEN`.
- Mail commands default `--mailbox-id` to `config.default_mailbox_id` or `me`.

Set a default mailbox:

```bash
lark config set --default-mailbox-id <id|me>
# or
lark mail mailbox set <MAILBOX_ID>
```

---

## Bitable (Base) concepts

Bitable is Lark/Feishu's database product. In the API, a **base** is also called an **app**.

- **App/Base:** the top-level container; identified by an app token.
- **Table:** a grid inside the base; defines fields (columns) and stores records (rows).
- **Field:** a column definition (type + name) used by every record in the table.
- **Record:** a single row of data for the table's fields.
- **View:** a saved presentation of a table (filters/sorts/grouping/hidden columns).

---

## Docs / Drive / Sheets concepts

- **Drive file:** generic file entity; identified by a **file token**. Folder is identified by **folder token**.
- **Docs (docx):** document is composed of **blocks** (list/get/update). `DOCUMENT_ID` is a Drive file token.
- **Sheets:** spreadsheet token identifies the file; **sheet_id** identifies a tab; ranges use A1 notation.

---

## Wiki concepts

- **Space:** top-level wiki container; identified by **space_id**.
- **Node:** wiki entry; identified by **node_token**, with `obj_type` describing the underlying content.
- Many wiki nodes point to Drive files; use Drive permissions for file-level access.

---

## Mail concepts

- **Mailbox:** identified by **mailbox_id** (or `me`).
- **Message:** identified by **message_id**; folders are identified by **folder_id** (Inbox, Sent, etc).

---

## Calendar / Meetings / Minutes concepts

- **Calendar event:** identified by **event_id**.
- **Meeting:** identified by **meeting_id** (different from join **meeting_no**).
- **Minutes:** meeting transcript/recording stored as a Drive file; identified by **minute_token**.

---

## Tasks concepts

- **Task list:** identified by **tasklist_guid**.
- **Task:** identified by **task_guid**; due/start support timestamps or date-only values.

---

## IM (Chats / Messages) concepts

- **Chat:** identified by **chat_id**.
- **Message:** identified by **message_id**; sending uses **receive_id** + `receive-id-type`.
- Message search and some IM operations require **user tokens**.

---

## Shell completion

```bash
lark completion bash
lark completion zsh
lark completion fish
lark completion powershell
```

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

Wiki SpaceMember.Create role-upsert verification test:

```bash
export LARK_TEST_WIKI_SPACE_ID=<space_id>
export LARK_TEST_USER_EMAIL=<member_email>

go test ./cmd/lark -run '^TestWikiMemberRoleUpdateIntegration$' -count=1 -v
# or:
make it-wiki-member-role-update
```

Prereqs: app creds configured (`lark auth login ...` or `LARK_APP_ID/LARK_APP_SECRET`) and cached user token (`lark auth user login`).

---

## Backlog / roadmap

See `BACKLOG.md` and `TODO.md` for planned work and open items.
