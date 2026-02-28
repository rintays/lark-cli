# Drive OAuth scopes (user OAuth)

This repo uses the **Lark Open Platform** (open.larksuite.com) scope codes for user OAuth. Feishu (open.feishu.cn) uses the same scope codes, but the Feishu docs UI sometimes hides codes behind hover popovers.

**Source of truth:** Lark Developer documentation “Scope list”
- https://open.larksuite.com/document/server-docs/getting-started/scope-list

## drive:drive:readonly
**Scope name (from scope list):** “View, comment, and download all files in My Space”

**Description (excerpt):**
> This scope allows an app to perform the following operations within the access range of Docs:
> Obtain file content
> Comment file content
> Save file content

**Used by lark-cli for read-only Drive operations**, e.g. `drive list`, `drive info`, `drive download`, `drive urls`, `drive search`.

## drive:drive
**Scope name (from scope list):** “View, comment, edit, and manage all files in My Space”

**Description (excerpt):**
> Add, delete, and modify a file
> Add, delete, and modify file content
> Add, delete, and modify permissions related to a file

**Note (excerpt):**
> This scope contains all the permissions of "drive:drive:readonly".

**Used by lark-cli for write/mutation Drive operations**, e.g. `drive upload`, `drive share`, `drive permissions add|update|delete`, `drive comment add|update|reply|reply-update`.

## drive:export:readonly
**Scope name (from scope list):** “Export Docs documents”

**Description (excerpt):**
> This scope allows an app to export Docs documents.

**Used by lark-cli for export operations**, e.g. `drive export` and `docs export`.

## drive:drive.metadata:readonly
**Scope name (from scope list):** “View the metadata of files in My Space”

This scope exists and may be sufficient for some metadata-only workflows, but **lark-cli currently standardizes on** `drive:drive:readonly` for read-only Drive operations (and `drive:drive` for write operations) to avoid surprising API gaps.
