# Concepts & IDs (Primer)

This primer is for agents new to Feishu/Lark.

## Feishu vs Lark

- Feishu (飞书) is the China brand.
- Lark is the global brand.
- API surface and identifiers are the same; UI labels may differ.
- API endpoints differ by brand/region; use `--platform feishu|lark` to select.

## Product map (common)

- IM: chats, messages
- Drive: files, permissions
- Docs/Sheets: docx, sheets
- Mail: mailboxes, messages
- Calendar: calendars, events
- Wiki: spaces, nodes
- Bitable: bases, tables, records
- Tasks: tasklists, tasks

## IDs and tokens

- Most commands take IDs as positional args.
- Many commands accept Lark/Feishu web URLs and extract IDs automatically.

Examples:

```bash
lark docs info <doc-id>
lark docs info https://.../docx/<doc-token>
lark drive info <file-token>
```

## Output modes

- Default: human-readable tables/text.
- `--json`: machine-readable output to stdout; logs/errors to stderr.

## Pagination

- Use `--limit` for list size.
- Use `--pages` for page count when available.
