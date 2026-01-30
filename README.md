# lark

A Golang CLI for Feishu/Lark inspired by gog.

## Usage

Configure credentials (writes `~/.config/lark/config.json` by default):

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET>
```

Optionally override the API base URL:

```bash
lark auth login --app-id <APP_ID> --app-secret <APP_SECRET> --base-url https://open.feishu.cn
```

Fetch a tenant access token (cached in config):

```bash
lark auth
```

Get tenant info:

```bash
lark whoami
```

Send a message:

```bash
lark msg send --chat-id <CHAT_ID> --text "hello"
```

Send a message to a user by email:

```bash
lark msg send --receive-id-type email --receive-id user@example.com --text "hello"
```

List recent chats:

```bash
lark chats list --limit 10
```

Search users:

```bash
lark users search --email user@example.com
lark users search --mobile "+1-555-0100"
lark users search --name "Ada"
lark users search --name "Ada" --department-id 0
```

List Drive files in a folder:

```bash
lark drive list --folder-id <FOLDER_TOKEN> --limit 20
```

Search Drive files by text:

```bash
lark drive search --query "budget" --limit 10
```

Get Drive file metadata:

```bash
lark drive get <FILE_TOKEN>
lark drive get --file-token <FILE_TOKEN>
```

Download a Drive file:

```bash
lark drive download --file-token <FILE_TOKEN> --out ./downloaded.bin
```

Get Drive file URLs:

```bash
lark drive urls <FILE_ID> [FILE_ID...]
```

Update Drive share permissions:

```bash
lark drive share <FILE_TOKEN> --type docx --link-share tenant_readable --external-access
```

Upload a file to Drive:

```bash
lark drive upload --file ./report.pdf --folder-token <FOLDER_TOKEN> --name "report.pdf"
```

Create a Docs (docx) document:

```bash
lark docs create --title "Weekly Update" --folder-id <FOLDER_TOKEN>
```

Get Docs (docx) metadata:

```bash
lark docs get --doc-id <DOCUMENT_ID>
```

Export a Docs (docx) document to PDF:

```bash
lark docs export --doc-id <DOCUMENT_ID> --format pdf --out ./document.pdf
```

Print a Docs (docx) document as text or Markdown:

```bash
lark docs cat --doc-id <DOCUMENT_ID> --format txt
```

Read a Sheets range:

```bash
lark sheets read --spreadsheet-id <SPREADSHEET_TOKEN> --range "Sheet1!A1:B2"
```

Fetch spreadsheet metadata:

```bash
lark sheets metadata --spreadsheet-id <SPREADSHEET_TOKEN>
```

List calendar events in a time range:

```bash
lark calendar list --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --limit 20
```

Create a calendar event:

```bash
lark calendar create --summary "Weekly Sync" --start "2026-01-02T03:04:05Z" --end "2026-01-02T04:04:05Z" --attendee dev@example.com
```

Get a contact user:

```bash
lark contacts user get --open-id <OPEN_ID>
lark contacts user get --user-id <USER_ID>
```

### Global flags

- `--config` override the config path.
- `--json` output JSON.
