---
name: lark-cli
description: Operate Feishu/Lark via the lark CLI (IM, Drive, Docs, Sheets, Mail, Calendar, Wiki, Bitable, Tasks, and more).
---

# Lark CLI

本技能用于通过 `lark` CLI 操作飞书/Lark 的各类产品能力（IM、Drive、Docs、Sheets、Mail、Calendar、Wiki、Bitable、Tasks 等），并提供最短路径的命令示例与参考资料索引。

## What this repo provides

- `lark` CLI: a single binary to access Feishu/Lark products (IM, Drive, Docs, Sheets, Mail, Calendar, Wiki, Bitable, Tasks).
- Two output modes: human tables by default, JSON with `--json` for automation.
- SDK-first implementation via the official `oapi-sdk-go`.

## Quickstart (minimal)

1) Install the CLI (see `references/INSTALL.md`).
2) Authenticate:
   - Tenant token (app-only, bot/app identity): `lark auth tenant`
   - User token (user-scoped, on behalf of a user): `lark auth user login`
   See `references/AUTH.md` for details and scopes.
3) Run a command:

```bash
lark whoami
lark chats list --limit 10
lark users search "Ada" --json
```

## Core concepts (tl;dr)

- Feishu (飞书) = Lark (global brand). Same API surface, different API endpoints.
- Most commands follow: `lark <product> <action> [args] [flags]`.
- Required IDs are positional args (no required `--id` flags).
- Many commands accept a Lark/Feishu web URL in place of IDs.
- `--json` prints machine-readable output to stdout; logs/errors go to stderr.

See `references/CONCEPTS.md` for a longer primer.

## When to use tenant vs user tokens

- Tenant token: app-level operations as your bot/app identity.
- User token: user-scoped operations on behalf of a specific user.

If a command fails with scope errors, check `references/TROUBLESHOOTING.md`.

## Agent-friendly workflow

- Prefer `--json` and parse in tools/scripts.
- Use `--limit` / `--pages` for pagination-heavy commands.
- Reuse `--account` or `LARK_ACCOUNT` for multi-user scenarios.

## Common recipes

See `references/RECIPES.md` for common tasks (send message, search users, read docs, etc.).

## Deep references

- Install: `references/INSTALL.md`
- Auth & scopes: `references/AUTH.md`
- Concepts & IDs: `references/CONCEPTS.md`
- Recipes: `references/RECIPES.md`
- Troubleshooting: `references/TROUBLESHOOTING.md`
- Completion: `references/COMPLETION.md`
- Docs: `references/DOCS.md`
- Sheets: `references/SHEETS.md`
- Bitable bases: `references/BASES.md`
- Drive: `references/DRIVE.md`
- Minutes: `references/MINUTES.md`
- Calendars: `references/CALENDARS.md`
- Meetings: `references/MEETINGS.md`
- Chats: `references/CHATS.md`
- Messages: `references/MESSAGES.md`
- Contacts: `references/CONTACTS.md`
- Mail: `references/MAIL.md`
- Tasklists: `references/TASKLISTS.md`
- Tasks: `references/TASKS.md`
- Users: `references/USERS.md`
- Config: `references/CONFIG.md`
- Whoami: `references/WHOAMI.md`
- Wiki: `references/WIKI.md`

## Repo hints (for agents working in this codebase)

- Build: `go build -o lark ./cmd/lark`
- Tests: `go test ./...`
- SDK docs: `go doc ./internal/larksdk`
