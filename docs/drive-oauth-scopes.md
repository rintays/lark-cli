# Drive OAuth scopes (user OAuth)

This repo uses the **Lark Open Platform** scope codes (open.larksuite.com) for **user OAuth**.

Feishu (open.feishu.cn) uses the **same scope codes** in practice; the main difference is the documentation domain/UI.

Primary sources:
- Scope list (scope codes): https://open.larksuite.com/document/server-docs/getting-started/scope-list
- API reference (endpoints): https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/overview

## lark-cli Drive-related services (internal/authregistry)

Drive commands are mapped to a small set of **bundle-level** services:

| Service | Required user OAuth scope codes |
|---|---|
| `drive-read` | `drive:drive.metadata:readonly`, `space:document:retrieve`, `drive:drive.search:readonly`, `docs:document.comment:read` |
| `drive-download` | `drive:file:download`, `drive:export:readonly` |
| `drive-write` | `drive:file:upload`, `docs:document.comment:create`, `docs:document.comment:update` |
| `drive-admin` | `docs:permission.member:retrieve`, `docs:permission.member:create`, `docs:permission.member:update`, `docs:permission.member:delete`, `docs:permission.setting:write_only` |

> Note: We intentionally avoid the legacy broad scopes `drive:drive` and `drive:drive:readonly`.

## Endpoint → scope-code mapping (evidence)

The table below maps the OpenAPI endpoints used by `lark-cli` Drive subcommands to the minimal scope codes we request.

| lark-cli command(s) | OpenAPI endpoint(s) (method + path) | Scope code(s) | Sources |
|---|---|---|---|
| `drive list`, `docs list`, `drive urls` (folder listing) | `GET /open-apis/drive/v1/files` | `space:document:retrieve` | Scope list + related API; API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file/list |
| `drive info` (metadata) | `GET /open-apis/drive/v1/files/:file_token` | `drive:drive.metadata:readonly` | Scope list (`drive:drive.metadata:readonly`). Note: this specific endpoint is not consistently exposed in the public reference UI; it is a metadata read and is covered by the metadata scope. |
| `drive search` | `POST /open-apis/drive/v1/files/search` | `drive:drive.search:readonly` | Scope list (`drive:drive.search:readonly`). (API Explorer requires login; public reference page for this endpoint is not always available.) |
| `drive download` | `GET /open-apis/drive/v1/files/:file_token/download` | `drive:file:download` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file/download |
| `drive upload` | `POST /open-apis/drive/v1/files/upload_all` | `drive:file:upload` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file/upload_all |
| `drive export`, `docs export` | `POST /open-apis/drive/v1/export_tasks` (create), `GET /open-apis/drive/v1/export_tasks/:ticket` (poll), `GET /open-apis/drive/v1/export_tasks/file/:file_token/download` (download) | `drive:export:readonly` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/export_task/create |
| `drive permissions list` | `GET /open-apis/drive/v1/permissions/:token/members` | `docs:permission.member:retrieve` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/permission-member/list |
| `drive permissions add` | `POST /open-apis/drive/v1/permissions/:token/members` | `docs:permission.member:create` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/permission-member/create |
| `drive permissions update` | `PUT /open-apis/drive/v1/permissions/:token/members/:member_id` | `docs:permission.member:update` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/permission-member/update |
| `drive permissions delete` | `DELETE /open-apis/drive/v1/permissions/:token/members/:member_id` | `docs:permission.member:delete` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/permission-member/delete |
| `drive share` | `PATCH /open-apis/drive/v1/permissions/:token/public` | `docs:permission.setting:write_only` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/permission-public/patch |
| `drive comment list`, `drive comment get` | `GET /open-apis/drive/v1/files/:file_token/comments` (list), `GET /open-apis/drive/v1/files/:file_token/comments/:comment_id` (get) | `docs:document.comment:read` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file-comment/list |
| `drive comment add`, `drive comment reply` | `POST /open-apis/drive/v1/files/:file_token/comments` (create comment), `POST /open-apis/drive/v1/files/:file_token/comments/:comment_id/replies` (create reply) | `docs:document.comment:create` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file-comment/create |
| `drive comment update`, `drive comment reply-update` | `PATCH /open-apis/drive/v1/files/:file_token/comments/:comment_id` (update/solve), `PUT /open-apis/drive/v1/files/:file_token/comments/:comment_id/replies/:reply_id` (update reply) | `docs:document.comment:update` | Scope list + API reference: https://open.larksuite.com/document/uAjLw4CM/ukTMukTMukTM/reference/drive-v1/file-comment/patch |

### Lark vs Feishu

- Scope codes in this document are taken from the **Lark scope list**.
- The corresponding Feishu pages usually exist under the same path on **open.feishu.cn**.

If you find an endpoint that still rejects these granular scopes (and only works with `drive:drive` / `drive:drive:readonly`), treat that as a documentation/API mismatch and document the error + endpoint so we can decide whether to widen scopes for that specific command.
