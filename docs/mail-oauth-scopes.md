# Verify Feishu/Lark Mail user OAuth scopes (send/get/list)

Goal: capture the exact user OAuth scope strings for Mail so we do not guess.

Status: Verified scopes for user mailbox messages:
- list/get: `mail:user_mailbox.message:readonly`
- send: `mail:user_mailbox.message:send`

Registry/tests updated to match these values.

## Step-by-step (Developer Console)
1. Open the Feishu/Lark Developer Console and select the target app.
2. Go to the permissions section (often labeled "Permissions/Scopes", "Permission Management", or "Permissions & Features").
3. Switch to the **user OAuth permissions** view (not app/tenant permissions).
4. Filter/search for **Mail** and locate the user permissions tied to:
   - sending a user mailbox message
   - listing user mailbox messages
   - getting a specific user mailbox message
5. For each permission, copy the **OAuth scope string** exactly as shown (the scope code used in the OAuth URL). Record the exact strings.

## Cross-check in the API Explorer (权限/Permission panel)
Feishu/Lark docs pages can be JS-heavy; the **API Explorer** pages reliably show the required permissions/scopes.

Open these API Explorer links (also available as comments in `oapi-sdk-go`):
- list: <https://open.feishu.cn/api-explorer?from=op_doc_tab&apiName=list&project=mail&resource=user_mailbox.message&version=v1>
- get: <https://open.feishu.cn/api-explorer?from=op_doc_tab&apiName=get&project=mail&resource=user_mailbox.message&version=v1>
- send: <https://open.feishu.cn/api-explorer?from=op_doc_tab&apiName=send&project=mail&resource=user_mailbox.message&version=v1>

Note: for Lark (global) docs, replace `open.feishu.cn` with `open.larksuite.com` and verify scopes there as well.

On each page:
1. Find the **权限/Permission** panel.
2. Confirm:
   - token type: **User access token** (or that user token is supported)
   - required **OAuth scope** string(s)
3. Copy the scope strings exactly and cross-check they match what you see in the Developer Console.

If multiple scopes/alternatives are listed, record the minimal set required for each endpoint.

## Record template
Use this quick table while verifying:

| Endpoint | API Explorer link | Console scope(s) | Explorer 权限/Permission scope(s) | Match? | Notes |
| --- | --- | --- | --- | --- | --- |
| list |  |  |  |  |  |
| get |  |  |  |  |  |
| send |  |  |  |  |  |

## Checklist: update authregistry once confirmed
- Update mail scopes in `internal/authregistry/registry.go`:
  - `RequiredUserScopes` = minimal union for list/get/send (stable-sorted).
  - `UserScopes.Readonly` = read/list-only scopes.
  - `UserScopes.Full` = send scope(s) plus any read scope required for send.
- If any tests assert specific mail scopes, update them to the verified values.
- Run `go test ./internal/authregistry` (or `go test ./...`) after the change.
