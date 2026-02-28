# Docx (Docs) OAuth scopes (user OAuth)

This repo uses the **Lark Open Platform** (open.larksuite.com) scope codes for user OAuth.

**Source of truth:** Lark Developer documentation “Scope list”
- https://open.larksuite.com/document/server-docs/getting-started/scope-list

## docx:document:readonly
**Scope name (from scope list):** “View upgraded Docs”

**Description (excerpt):**
> This scope allows an app to do the following within the permitted access:
> View upgraded Docs

Used by lark-cli for read-only Docx operations (query document, raw content, blocks).

## docx:document:create
**Scope name (from scope list):** “Create upgraded Docs”

**Description (excerpt):**
> With this scope added, an app can create Docs in My Space.

Used by lark-cli for doc creation operations.

## docx:document:write_only
**Scope name (from scope list):** “Edit upgraded Docs”

**Description (excerpt):**
> With this scope added, an app can add, edit, or delete content within upgraded Docs it can access.

Used by lark-cli for write/mutation Docx operations (creating/updating/deleting blocks, overwriting content).

## docx:document.block:convert
**Scope name (from scope list):** “Convert text content to cloud document blocks”

Used by lark-cli for text/markdown → blocks conversion flows.

## Why lark-cli avoids docx:document
The scope list includes a broader scope:

- **docx:document** — “Create and edit upgraded Docs”

However, per the scope list description, it is a superset that overlaps with the more granular scopes above (create + readonly + write).

To minimize consent prompts and permissions, lark-cli prefers requesting the granular scopes (`docx:document:create`, `docx:document:readonly`, `docx:document:write_only`, plus `docx:document.block:convert` when needed) instead of requesting `docx:document`.
