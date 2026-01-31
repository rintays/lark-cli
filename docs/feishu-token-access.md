# Feishu/Lark OpenAPI Token 支持矩阵（当前仓库使用）

本文针对仓库当前代码实际调用的 OpenAPI 进行调研与梳理，给出每个 API 的 Token 支持情况，并提出调用时的 token 选择策略，供后续重构使用。

## 适用范围
- 覆盖 `internal/larksdk` 与相关集成测试/fixtures 中出现的所有 OpenAPI 调用。
- 以当前依赖的官方 Go SDK `oapi-sdk-go` v3.5.3 的元信息为主，辅以官方文档方法列表补全未覆盖接口。

## Token 类型说明（简版）
- Tenant Access Token：应用在租户维度的访问凭证，适合管理员或应用级权限。
- User Access Token：用户授权后的访问凭证，适合以用户身份访问用户资源。
- App Access Token：用于 OAuth 相关换取/刷新 User Access Token。
- None：仅用 app_id + app_secret 获取 token 本身。

## 调用时 Token 选择策略（完善版）

### 目标
- 当 API 仅支持一种 token 时自动使用该类型。
- 当 API 支持两种 token 时，遵循“用户默认偏好”。
- 用户显式指定与 API 不兼容时，给出明确错误提示与指引。

### 规则（建议落地为统一 resolver）
1. **确定 API 允许的 token 类型**（见下方矩阵）。
2. **读取用户显式选择**：命令行 `--token-type`（可选值 `tenant|user|auto`）。
3. **读取默认偏好**：配置 `default_token_type`（可选 `tenant|user`，默认 `tenant`）。
4. **选择逻辑**：
   - 仅支持 `Tenant`：
     - `--token-type=user` → 报错并提示需要 tenant token。
     - 其他 → 使用 tenant token。
   - 仅支持 `User`：
     - `--token-type=tenant` → 报错并提示需要 user token。
     - 其他 → 使用 user token。
   - 支持 `Tenant / User`：
     - `--token-type=tenant|user` → 直接使用指定类型。
     - `--token-type=auto` → 使用 `default_token_type`。
   - `App`：仅用于 OAuth 刷新等流程，不参与常规命令选择。
   - `None`：仅用于获取 token 本身，不参与常规命令选择。
5. **缺失 token 时的错误提示**：
   - tenant 缺失 → 提示执行 `lark auth` 或配置 `LARK_APP_ID/LARK_APP_SECRET`。
   - user 缺失 → 提示执行 `lark auth user` 或配置 `LARK_USER_ACCESS_TOKEN`。
6. **未知/待确认 API**：
   - 默认按 `default_token_type` 选择，但打印警告提示“需复核 API Explorer”。

## API Token 支持矩阵（当前仓库使用）

> 说明：
> - “待确认”表示未在公开方法列表或 SDK 元信息中命中，需要 API Explorer 复核。
> - 部分路径参数命名差异（如 `mailbox_id` vs `user_mailbox_id`）仅是参数名不同，指向同一接口。

## auth

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/auth/v3/app_access_token/internal` | 无（AppID+AppSecret） | 用于获取 access token |
| POST | `/open-apis/auth/v3/tenant_access_token/internal` | 无（AppID+AppSecret） | 用于获取 access token |

## authen

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/authen/v1/refresh_access_token` | App | 需 app_access_token |

## tenant

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/tenant/v2/tenant/query` | Tenant |  |

## contact

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/contact/v3/users/:user_id` | Tenant / User |  |
| POST | `/open-apis/contact/v3/users/batch_get_id` | Tenant |  |
| GET | `/open-apis/contact/v3/users/find_by_department` | Tenant / User |  |

## im

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/im/v1/chats` | Tenant / User |  |
| POST | `/open-apis/im/v1/chats` | Tenant |  |
| DELETE | `/open-apis/im/v1/chats/:chat_id` | Tenant / User |  |
| POST | `/open-apis/im/v1/messages` | Tenant / User |  |

## drive

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/drive/v1/export_tasks` | Tenant / User |  |
| GET | `/open-apis/drive/v1/export_tasks/:ticket` | Tenant / User |  |
| GET | `/open-apis/drive/v1/export_tasks/file/:file_token/download` | Tenant / User |  |
| GET | `/open-apis/drive/v1/files` | Tenant / User |  |
| DELETE | `/open-apis/drive/v1/files/:file_token` | Tenant / User |  |
| GET | `/open-apis/drive/v1/files/:file_token` | Tenant / User |  |
| GET | `/open-apis/drive/v1/files/:file_token/download` | Tenant / User |  |
| POST | `/open-apis/drive/v1/files/create_folder` | Tenant / User |  |
| POST | `/open-apis/drive/v1/files/search` | 待确认 | 未在公开方法列表/SDK中命中，需 API Explorer 复核 |
| POST | `/open-apis/drive/v1/files/upload_all` | Tenant / User |  |
| PATCH | `/open-apis/drive/v1/permissions/:file_token/public` | Tenant / User |  |

## docx

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/docx/v1/documents` | Tenant / User |  |
| GET | `/open-apis/docx/v1/documents/:document_id` | Tenant / User |  |

## sheets

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| PUT | `/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values` | Tenant / User |  |
| GET | `/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values/:range` | Tenant / User |  |
| POST | `/open-apis/sheets/v2/spreadsheets/:spreadsheet_token/values_append` | Tenant / User |  |
| POST | `/open-apis/sheets/v3/spreadsheets` | Tenant / User |  |
| GET | `/open-apis/sheets/v3/spreadsheets/:spreadsheet_token` | Tenant / User |  |
| POST | `/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/delete_dimension` | 待确认 | 未在公开方法列表/SDK中命中，需 API Explorer 复核 |
| POST | `/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension` | 待确认 | 未在公开方法列表/SDK中命中，需 API Explorer 复核 |
| GET | `/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/query` | Tenant / User |  |

## bitable

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/bitable/v1/apps` | Tenant / User |  |
| GET | `/open-apis/bitable/v1/apps/:app_token/tables` | Tenant / User |  |
| POST | `/open-apis/bitable/v1/apps/:app_token/tables` | Tenant / User |  |
| DELETE | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id` | Tenant / User |  |
| GET | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/fields` | Tenant / User |  |
| POST | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records` | Tenant / User |  |
| DELETE | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id` | Tenant / User |  |
| GET | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id` | Tenant / User |  |
| PUT | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/:record_id` | Tenant / User |  |
| POST | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/records/search` | Tenant / User |  |
| GET | `/open-apis/bitable/v1/apps/:app_token/tables/:table_id/views` | Tenant / User |  |

## wiki

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| POST | `/open-apis/wiki/v1/nodes/search` | User |  |
| GET | `/open-apis/wiki/v2/spaces` | Tenant / User |  |
| GET | `/open-apis/wiki/v2/spaces/:space_id` | Tenant / User |  |
| GET | `/open-apis/wiki/v2/spaces/:space_id/nodes` | Tenant / User |  |
| GET | `/open-apis/wiki/v2/spaces/get_node` | Tenant / User |  |

## calendar

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/calendar/v4/calendars/:calendar_id/events` | Tenant / User |  |
| POST | `/open-apis/calendar/v4/calendars/:calendar_id/events` | Tenant / User |  |
| POST | `/open-apis/calendar/v4/calendars/:calendar_id/events/:event_id/attendees` | Tenant / User |  |
| POST | `/open-apis/calendar/v4/calendars/primary` | Tenant / User |  |

## mail

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/mail/v1/public_mailboxes` | Tenant / User |  |
| GET | `/open-apis/mail/v1/user_mailboxes/:mailbox_id/folders` | Tenant / User | 参数名与文档略有差异 |
| GET | `/open-apis/mail/v1/user_mailboxes/:user_mailbox_id` | Tenant |  |
| GET | `/open-apis/mail/v1/user_mailboxes/:user_mailbox_id/messages` | Tenant / User |  |
| GET | `/open-apis/mail/v1/user_mailboxes/:user_mailbox_id/messages/:message_id` | Tenant / User |  |
| POST | `/open-apis/mail/v1/user_mailboxes/:user_mailbox_id/messages/send` | User |  |

## minutes

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/minutes/v1/minutes` | 待确认 | 未在公开方法列表/SDK中命中，需 API Explorer 复核 |
| GET | `/open-apis/minutes/v1/minutes/:minute_token` | Tenant / User |  |

## vc

| 方法 | 路径 | 支持 Token | 备注 |
| --- | --- | --- | --- |
| GET | `/open-apis/vc/v1/meeting_list` | Tenant / User |  |
| GET | `/open-apis/vc/v1/meetings/:meeting_id` | Tenant / User |  |

## 待确认清单（优先核对）
- `POST /open-apis/drive/v1/files/search`
- `GET /open-apis/minutes/v1/minutes`
- `POST /open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension`
- `POST /open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/delete_dimension`

## 主要来源（后续可替换为 API Explorer 直出）
- 官方 Go SDK `oapi-sdk-go` v3.5.3（本地依赖）
- 飞书开放平台方法列表/概述页面（Apifox 镜像）
- 飞书开放平台访问凭证说明
