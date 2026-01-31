# lark CLI — BACKLOG (maintained by AG)

**Owner:** AG  
**Where:** `/Users/fredliang/clawd/lark/BACKLOG.md` (single source of truth)  
**Repo:** `/Users/fredliang/clawd/lark`  
**Branch policy:** default development branch is **`main`** and changes must land on **`origin/main`** (unless Master explicitly requests a branch/PR flow).  
**Last updated:** 2026-01-31 (Asia/Shanghai)

---

## 0) Background / Why this backlog exists

We’re building **`lark`**, a Golang CLI for Feishu/Lark.

Key constraints / goals:
- **Good CLI UX**: command tree is consistent, flags predictable, output supports both human and JSON.
- **SDK-first**: prefer the official **`oapi-sdk-go`**. If SDK doesn’t cover an endpoint, still stay “SDK-native” by using **`core.ApiReq`** wrappers in `internal/larksdk`.
- **Full E2E integration tests**: user confirmed real write operations are OK; we should validate flags and behavior against the real APIs.
- **Feishu vs Lark platform**: base domain differs; we need a convenient switch.

This file is the operational “what to do next” plan so work doesn’t drift across chat logs.

---

## 1) Working agreements / definitions

### 1.1 Tokens / auth model
- **Tenant access token**: used for most app-level APIs.
- **User access token** (OAuth): required for certain features (notably **Mail send**, Wiki search v1, Wiki create space, etc.).

Principles:
- If a command *requires* a user token, it should:
  1) try to load from config (auto-managed) via `ensureUserToken()`
  2) allow override by flag/env for power users
  3) error clearly when not available

### 1.2 Platforms / base domains
Feishu (CN) and Lark (overseas) use different base domains.

Decision:
- Keep `--base-url` (expert override).
- Add a convenient **`--platform feishu|lark`** switch (user requested).
  - `feishu` → `https://open.feishu.cn`
  - `lark` → `https://open.larksuite.com`

Precedence:
1) `--base-url` (highest)
2) `--platform`
3) config value
4) default (feishu)

### 1.3 “Done” definition for each backlog item
For any item to be considered DONE:
- Includes docs (README help snippet) for user-facing changes
- Has tests:
  - unit tests (httptest) when possible
  - integration tests for end-to-end behavior when relevant
- `go test ./...` is green
- committed as a small, reviewable step and pushed to `origin/main`

---

## 2) Current repo state (so we don’t lose context)

- `origin/main` now contains the merged baseline feature set (previous dev/codex work has been merged into main).
- Auto-advance cronjob exists to keep pushing forward:
  - It must read this backlog file and pick the highest-priority small deliverable.

---

## 3) Top priorities (in order)

### A) Auth (user OAuth / user_access_token) — highest priority

> Note: Goal is first-class user-token flows so that features requiring user scope (Mail send, Wiki search, etc.) work without manual token pasting.
>
> Risk/concern (Master): “权限/凭证触发与管理”后面可能会踩坑。需要参考 gog 的成熟做法做一次系统性设计。

**Why:** unlocks Mail send + Wiki user-only endpoints; makes CLI viable.

Deliverables:
- [x] **Command:** `lark auth user login`
  - [x] Runs OAuth authorization code flow on localhost
  - [x] default port **17653**
  - [x] callback path: `/oauth/callback`
  - [x] opens browser automatically
  - [x] stores: `user_access_token`, `refresh_token`, expiry, and any user identifiers we can capture
  - [x] prints a short success message and current user identity (if possible)
  - [x] clear error if redirect URL not registered

**Notes / setup:** register redirect URL in console:
`http://localhost:17653/oauth/callback`

- [x] **Library:** `ensureUserToken()`
  - [x] loads token from config
  - [x] refreshes via `refresh_token` when near expiry
  - [x] handles refresh failures (clear tokens + instruct re-login)

- [x] **Mail send uses user token by default**
  - [x] default path: auto-loaded user token
  - [x] still allow `--user-access-token` and env `LARK_USER_ACCESS_TOKEN` override
  - [x] tests cover:
    - missing token error
    - token override precedence

Acceptance criteria:
- A fresh machine with `LARK_APP_ID/LARK_APP_SECRET` set can complete `auth user login` and then run `mail send` without manually pasting a token.

Credential/scopes management (research + design, gog-inspired):

- ✅ **Design doc (living)**: `/Users/fredliang/clawd/design/lark-auth-credentials-design.md`
  - covers token taxonomy (tenant/app/user/refresh)
  - services→scopes model (gog-like)
  - storage backend policy (config vs keychain)
  - multi-profile selection + precedence rules
  - re-auth triggers (force-consent analogue)
  - phased implementation plan + risks

Work items (must follow the design, not ad-hoc patches):
- [ ] **Service registry (gog-style)**
  - [x] Define fixed service set (im/drive/docx/sheets/calendar/mail/wiki/base/…)
  - [x] Each service declares: token type(s) (tenant/user), user scopes, offline requirement
  - [x] Compute **stable sorted union** of required scopes (deterministic + testable)
  - [x] Declare RequiredUserScopes for wiki + mail (mail scope is best-effort; TODO verify against official Feishu/Lark docs; see `docs/mail-oauth-scopes.md`)
  - [x] Map commands → services (so runtime can explain “why you need this token/scope”)
    - [x] Initial command→service mapping scaffold in `internal/authregistry` (unit-tested; wired into runtime remediation hints: tailored `auth user login --scopes ...` suggestions)
    - [x] Longest-prefix matching for command paths (e.g. "drive list" maps via "drive")
    - [x] Metadata aggregation helper: command → (services, token types, offline requirement, required user scopes)
    - [x] Detect TokenUser services missing declared RequiredUserScopes (so we don’t pretend we know scopes yet)
- [ ] **Scope variants as first-class knobs**
  - [x] `lark auth explain --readonly` (use per-service UserScopes variants when available; fallback to RequiredUserScopes)
  - [ ] `--readonly` mode (where feasible)
  - [ ] per-service scope variants (if Feishu/Lark has meaningful levels; drive is the likely one)
- [ ] **Incremental authorization** (gog `include_granted_scopes=true` analogue)
  - [ ] Default to incremental grant when adding services/scopes (avoid re-consenting everything)
- [ ] **Explicit re-auth triggers** (gog `--force-consent` analogue)
  - [x] Add `lark auth user login --force-consent` (or equivalent) to force prompt/consent
  - [x] Trigger guidance when: scopes changed, refresh_token missing, insufficient_scope/permission errors
    - [x] refresh_token missing: ensureUserToken/expireUserToken suggests `lark auth user login --scope offline_access --force-consent`
    - [x] scopes changed (auth user login stores canonical scope string + warns when changed)
    - [x] insufficient_scope/permission errors (best-effort: wiki node search adds re-login hint)
  - [x] Make remediation messages print the exact command to run next
- [ ] **Token storage backend policy + implementation**
  - [ ] Backend selection: `auto|keychain|file` (keyring)
  - [ ] Env > config precedence (e.g., `LARK_KEYRING_BACKEND`, `LARK_KEYRING_PASSWORD` for headless)
  - [ ] Store refresh token as JSON payload including metadata (`services`, `scopes`, `created_at`) to power `auth status`
- [ ] **Multi-profile / multi-account / multi-app isolation**
  - [ ] `--profile` / `LARK_PROFILE` selection + default
  - [ ] “client bucket” analogue (gog `--client`): isolate refresh tokens by app_id/base_url/profile to avoid mixing credentials
- [ ] **Auth status & remediation UX**
  - [x] `lark auth user status` shows: offline/refresh availability, expiry, and stored scope (minimal v1)
  - [x] Standardized remediation messages:
    - [x] missing refresh_token → tell user to rerun with `--force-consent` / correct scopes
    - [x] revoked refresh_token → clear + rerun login
    - [x] insufficient scope → suggest adding service/scope and re-login (best-effort: wiki node search adds re-login hint)

---

### B) SDK-first migration + delete legacy HTTP client

> Note: “SDK-first” means: use typed SDK services whenever possible; if SDK has a gap, implement a small wrapper using `core.ApiReq` inside `internal/larksdk` (still within SDK ecosystem).

**Why:** reduce maintenance & inconsistency; SDK gives typed requests and consistent auth.

Deliverables:
- [ ] Migrate commands to SDK services in this order:
  - [x] `users` — already SDK-backed; tests/validation tightened
  - [x] `chats` (list now uses oapi-sdk-go im/v1 typed service)
  - [x] `msg` — msg send uses oapi-sdk-go im/v1
  - [ ] `drive`
    - [x] `drive urls` → SDK (`GetDriveFileMetadata`)
    - [x] `drive share` (public permission update) → SDK (`UpdateDrivePermissionPublic`)
    - [x] `drive search` → SDK (`/drive/v1/files/search`)
    - [x] `drive list` → SDK types (no legacy `larkapi` types)
    - [x] `drive download` → SDK typed service (`Drive.V1.File.Download`)
    - [x] `drive info` → SDK-native `core.ApiReq` wrapper (`GET /drive/v1/files/:file_token`)
  - [x] `drive upload` → SDK typed service (`Drive.V1.File.UploadAll`)
  - [x] `drive export` → SDK typed service (`Drive.V1.ExportTask.*`)
  - [x] remaining drive subcommands: none (drive fully migrated)
  - [ ] `docs`
    - [x] `docs info` → SDK typed service (`Docx.V1.Document.Get`)
    - [x] `docs create` → SDK typed service (`Docx.V1.Document.Create`)
    - [x] `docs export/cat` → Export task + download (supports `txt` and `md` output)
    - [ ] **Docs Markdown ⇄ Docx bidirectional sync** (requested)
      - Goal: support keeping a local `.md` file and a Docx document in sync.
      - [ ] Export: already supported via `docs cat --format md` (ensure docs + integration test)
      - [ ] Import/overwrite: add a command (TBD):
        - candidate: `lark docs import --doc-id <ID> --file <path.md> [--mode overwrite|append]`
        - or: `lark docs sync --doc-id <ID> --file <path.md> [--direction push|pull|both]`
      - [ ] Decide implementation path:
        - prefer official Docx content APIs (blocks) if stable
        - otherwise evaluate any official “import markdown” endpoint/workflow
      - [ ] Integration tests: roundtrip `md → docx → md` for a small fixture
  - [ ] `sheets`
    - [x] `sheets info` → SDK typed service (`Sheets.V3.Spreadsheet.Get`)
  - [ ] `mail`
    - [x] `mail info` → SDK typed service (`Mail.V1.UserMailboxMessage.Get`)
    - [x] `mail public-mailboxes list` → SDK typed service (`Mail.V1.PublicMailbox.List`)
    - [x] `mail send` → SDK typed service (`Mail.V1.UserMailboxMessage.Send`)
    - [x] `mail list` → SDK-first: List IDs (`Mail.V1.UserMailboxMessage.List`) + Get details (`Mail.V1.UserMailboxMessage.Get`)
- [ ] Remove legacy HTTP client code paths for endpoints covered by SDK.
  - [x] Drop `coreConfig` availability gating for SDK-only `ListMailMessages` in `internal/larksdk`.
- [ ] Delete transitional “fallback” tests after migration.
- [ ] Maintain a **Coverage Matrix** section in this backlog (or a sibling doc later):
  - desired API → SDK support? (yes/no)
  - token type: tenant/user
  - version: v1/v2
  - if not supported: wrapper name in `internal/larksdk` using `core.ApiReq`

Acceptance criteria:
- Commands compile with minimal/no dependency on `internal/larkapi` where SDK covers.

---

### C) Mail UX improvements

**Why:** reduce friction; mailbox selection shouldn’t be required for common flows.

Updated facts (Feishu Mail OpenAPI):
- For user-token operations, `user_mailbox_id` can be the literal **`me`** (current authenticated user).
- Listing user mailboxes via `GET /mail/v1/user_mailboxes` appears unavailable (404 on `open.feishu.cn`), but **public mailboxes** can be listed.

Deliverables:
- [x] Add discovery command for public mailboxes:
  - [x] `mail public-mailboxes list` (returns `mailbox_id`)
- [x] Make `--mailbox-id` optional for user-mailbox operations
  - [x] default precedence: `--mailbox-id` > `config.default_mailbox_id` > **`me`**
  - [x] apply consistently across: `mail folders`, `mail list`, `mail info`, `mail send`
  - Notes: default mailbox resolution applied across folders/list/info/send; tests cover precedence + defaults.
- [x] Config helper commands
  - [x] `mail mailbox set` (persist default mailbox id)

Acceptance criteria:
- `mail send` works with no `--mailbox-id` (defaults to `me`).
- `mail folders/list/info` work with no `--mailbox-id` if API supports `me`; otherwise they fall back to config/default with a clear error.

---

### D) Sheets completion (row/col operations)

**Why:** common spreadsheet operations.

Deliverables:
- [ ] Add row/col operations (Sheets v3 sheet-rowcol):
  - [x] insert rows (v3 insert_dimension, `sheets rows insert`) (endpoint: `/open-apis/sheets/v3/spreadsheets/:spreadsheet_token/sheets/:sheet_id/insert_dimension`)
  - [x] insert cols (v3 insert_dimension, `sheets cols insert`)
  - [x] delete rows (`sheets rows delete`)
  - [x] delete cols (`sheets cols delete`)
  - [ ] follow gog-style command tree
  - [ ] SDK-first; otherwise `core.ApiReq` wrappers

Acceptance criteria:
- Integration tests cover at least one insert and one delete.

---

### E) Base (Bitable) — new top-level command `base`

**Why:** critical for real workflows.

P0 deliverables:
- [x] `base table list`
- [x] `base field list`
- [x] `base view list`
- [x] `base record info`
- [x] `base record search`
- [x] `base record create`
- [x] `base record update`
- [x] `base record delete`

P1:
- [x] `base table create`
- [x] `base table delete`
- [ ] record batch operations
- [ ] schema/view management
  - [ ] `base field create/update/delete`
  - [ ] `base view create/update/delete`
  - [ ] `base record batch-create/batch-update/batch-delete`

P2:
- [ ] `base app create/info/update/copy` (SDK supports; enables CLI-only lifecycle)
- [ ] `base list` / `base app list` (discover app_token via Drive/Wiki)
  - [ ] implement via `drive search --type bitable --query ...` and parse `file.url` to extract app_token
- [ ] attachments workflows across Drive

Acceptance criteria:
- P0 has unit tests + at least one integration test for create/update/delete.
  - Integration env: set `LARK_INTEGRATION=1` (or `INTEGRATION=1`) and:
    - `LARK_APP_ID`, `LARK_APP_SECRET`
    - (optional) `LARK_TEST_APP_TOKEN`, `LARK_TEST_TABLE_ID` (Base/Bitable tests now auto-create/find a dedicated Base app and manage temporary tables unless overridden)
    - optional `LARK_TEST_FIELD_NAME` (otherwise auto-pick first text field)

---

### F) Command tree consistency

Discovery coverage (list/search) gaps to close for “CLI-only” workflows:
- [x] Docs/Docx: add list/search commands so users can find doc tokens without leaving CLI
  - [x] `docs search --query ...` (filters drive search results to docx)
  - [x] `docs list --folder-id ...` (filters drive list results to docx)
  - Fallback (documented): `drive search --type docx --query ...`
- [x] Sheets: add list/search commands so users can find spreadsheet tokens without leaving CLI
  - [x] `sheets search --query ...` (filters drive search results to sheet)
  - [x] `sheets list --folder-id ...` (filters drive list results to sheet)
  - [x] `sheets create --title ... [--folder-id ...]` (creates a spreadsheet; returns spreadsheet_token)
  - Fallback (documented): `drive search --type sheet --query ...`
- [x] Meetings: add `meetings list` (so users can find meeting_id without leaving CLI)
  - [x] Research API availability and required token type
  - [x] Implement `meetings list` + `--limit/--page-token` (or equivalent) + unit tests
- [ ] Drive: `drive list/search` exist, but add better discoverability flags if needed:
  - [x] `drive search --type <docx|sheet|bitable|file|doc>` (implemented; request uses file_types + README example)
  - [x] `drive search --folder-id <token>` (implemented; request uses folder_token)
  - [x] `drive search --pages <N>` caps pagination (prevents unbounded API calls); unit-tested

Mail CLI-only usability gaps:
- [x] Make `--mailbox-id` optional across user-mailbox commands (default to `me`):
  - [x] `mail folders` (defaults mailbox-id via resolveMailboxID; help text updated)
  - [x] `mail list` (already defaults mailbox-id via resolveMailboxID + tests cover)
  - [x] `mail info` (message-id only; mailbox defaults via resolveMailboxID + tests cover)
  - [x] `mail send` (defaults mailbox-id via resolveMailboxID + tests cover)
  - [x] Update help text + README examples
- [ ] Config CRUD to support “CLI-only” setup (no manual editing config.json):
  - [x] `lark config info`
  - [x] `lark config set --base-url ...`
  - [x] `lark config set --platform feishu|lark`
  - [x] `lark config unset --base-url`
  - [x] `lark config unset --default-mailbox-id`
  - [x] `lark config unset --user-tokens`
  - [ ] Fill remaining config knobs for true CLI-only workflows:
    - [x] `lark config set --default-mailbox-id <id|me>` (parity with unset)
    - [x] `lark config set --app-id ... --app-secret ...` (optional; alternative to `lark auth login`)
    - [x] `lark config list-keys` (or document all supported keys in `lark config set --help`)
    - [ ] Multi-profile config selection (after profiles land): `--profile` / `LARK_PROFILE` + per-profile info/set
  - [ ] (Alternative) keep domain-specific where it’s clearer: `auth platform set/info`, `mail mailbox info-default/unset-default`, `auth user status/logout`


**Why:** users build muscle memory; consistency beats features.

Deliverables:
- [ ] Audit naming: singular vs plural
  - [x] `contacts users` renamed to `contacts user` (avoid overlap with top-level `users`)
  - `meeting` → `meetings` already done
  - [x] Policy: top-level resource collections use plural canonical names; abbreviations are aliases; keep backward-compatible aliases when renaming. Rationale: consistent help discovery and stable scripts.
  - [x] `calendar` → `calendars` (keep `calendar` as alias)
  - [x] `msg` → `messages` (keep `msg` as alias)
  - [x] `msg` short help clarified to "Send chat messages"
- [ ] Align help text and examples.
  - [x] `users` top-level Short changed to "Manage users"
  - [x] `mail mailbox info` defaults mailbox-id (flag > config default_mailbox_id > `me`) + unit test
  - [x] `mail folders/list` help now documents mailbox-id defaulting (commit 23c634c)
  - [x] README now points to the correct backlog path (`/Users/fredliang/clawd/lark/BACKLOG.md`)
- [x] Fix: `docs` command no longer registers `list` twice.

Additional consistency work:
- [ ] **Required flags validation**: use Cobra’s `MarkFlagRequired` / `Args` validators consistently across commands.
  - [x] `drive list` now rejects positional args (Args=cobra.NoArgs) + test
  - [x] `calendar list/create` now reject positional args (Args=cobra.NoArgs) + test
  - [x] `drive info` now uses required flag validation for `--file-token` (positional arg sets the flag) + unit test asserts stable required-flag error
  - [x] `drive export` now uses required flag validation for `--file-token` (positional arg sets the flag) + unit test asserts stable required-flag error
  - [x] `drive share` now uses required flag validation for `--file-token` (positional arg sets the flag) + unit test asserts stable required-flag error
  - Goal: missing required flags should fail *before* making API calls.
  - Commands should not rely on scattered `if x == ""` checks.
  - Keep runtime validations for things like file existence, output path not a directory, etc.
  - Add/adjust tests to ensure required-flag errors are stable.

---

### G) Wiki (research → implementation)

**Why:** wiki is a major surface (spaces/nodes/search/permissions) and is needed for discovery workflows.

Known facts from prior research (must be reflected in code decisions):
- Wiki v2: SDK covers most endpoints.
- Wiki create space: user token only.
- Wiki search: **v1** `POST /wiki/v1/nodes/search` and user token only; SDK service/wiki/v2 doesn’t include this.

Deliverables:
- [ ] P0 (v2 SDK-backed):
  - [x] `wiki space list/info` implemented (v2)
  - [x] `wiki node info/list`
    - [x] `wiki node info` implemented (v2)
    - [x] `wiki node list` implemented (v2)
  - [x] `wiki member list` implemented (v2)
  - [ ] `wiki member` management (v2)
    - [x] `wiki member delete`
    - [x] `wiki member add` (SpaceMember.Create)
    - [ ] Verify whether SpaceMember.Create is an upsert that can change roles for existing members ("update role")
      - Integration test added: `cmd/lark/wiki_member_role_update_integration_test.go`
      - How to run (single test in a real env):
        - Required env vars: `LARK_INTEGRATION=1`, `LARK_TEST_WIKI_SPACE_ID`, `LARK_TEST_USER_EMAIL`
        - Prereqs: app creds configured (`lark auth login ...` or `LARK_APP_ID/LARK_APP_SECRET`) and cached user token (`lark auth user login`)
        - Command:
          - `LARK_INTEGRATION=1 LARK_TEST_WIKI_SPACE_ID=<space_id> LARK_TEST_USER_EMAIL=<member_email> go test ./cmd/lark -run '^TestWikiMemberRoleUpdateIntegration$' -count=1 -v`
  - [x] `wiki task` query (`GET /open-apis/wiki/v2/tasks/:task_id`)
- [ ] P1 (gap fill):
  - [x] implement `internal/larksdk/wiki_search_v1.go` using `core.ApiReq`
  - [x] expose `wiki node search`

Acceptance criteria:
- Integration test can do a wiki search (requires user token) once auth user login exists.

---

### H) Integration testing (full end-to-end, writable)

**Goal:** prove the CLI is correct from a user’s POV: flags, output, and real API behavior.

User confirmed:
- `LARK_APP_ID` / `LARK_APP_SECRET` already configured
- app has broad permissions
- write operations are OK (no need to avoid “pollution”)
- `auth user login` can use browser automation

#### H1) How to run
- Default `go test ./...` should remain unit-only (httptest / no real network).
- Integration tests run only when explicitly enabled:
  - [x] env gate: `LARK_INTEGRATION=1` (or `INTEGRATION=1`)

#### H2) What to test (rules)
Each integration test must validate:
- [ ] command succeeds (happy path)
- [ ] `--json` output parses and includes expected fields
- [ ] missing required flags produce stable, human-readable errors
- [ ] where applicable: write operations really happened (assert via follow-up GET/list/search)

#### H3) Test targets / fixtures (make it repeatable)
Integration tests should be runnable with minimal env.

Fixture strategy:
- Tests dynamically create required resources (Drive folder, spreadsheet, chat) under a predictable name prefix.
- A sweeper pass runs at the beginning and end of the suite to delete leftovers by prefix (best-effort).
  - Prefixes: `lark-cli-it-` (current) and `lark-it-`/`clawdbot-it` (legacy).
- Tests use `--config <temp>` so cached tokens are isolated from the developer’s real config.

Still-required env vars (fail-fast by default; can opt into skip with `LARK_INTEGRATION_ALLOW_SKIP=1`):
- `LARK_TEST_USER_EMAIL` (needed to resolve a real user and create a chat / send messages)
- `LARK_TEST_MAIL_TO` (recipient for `mail send`)
- (optional) `LARK_TEST_APP_TOKEN` + `LARK_TEST_TABLE_ID` (Base/Bitable integration tests; otherwise tests create/find the `lark-cli-it-base` app and manage temporary tables automatically)

Optional env vars:
- `LARK_TEST_FIELD_NAME` (Base field override; auto-detected when possible)
- `LARK_TEST_DOC_ID` (if/when adding export/cat tests for a stable existing doc)

#### H4) How to invoke CLI from tests
Prefer invoking Cobra commands directly (faster, controllable stdout/stderr), but allow subprocess mode when needed:
- [ ] direct Cobra invocation with in-memory stdout buffer
- [ ] subprocess `go run ./cmd/lark ...` for true “binary-like” behavior tests

#### H5) Milestones
Milestone 1 (framework + minimal chain):
- [x] `auth` (tenant token; gated integration test)
- [x] `whoami` (gated integration test)
- [x] `chats list` (gated integration test)
- [x] `users search` (gated integration test; set LARK_TEST_USER_EMAIL)
- [x] `msg send` (gated integration test; set LARK_TEST_CHAT_ID)

Milestone 2 (writes):
- [x] `drive upload` (gated integration test; set LARK_TEST_FOLDER_TOKEN)
- [x] `docs create` (gated integration test; uses LARK_TEST_FOLDER_TOKEN)
- [x] `sheets update/append/clear` (gated integration tests; verify by read TBD)
  - [x] `sheets update` (gated integration test; requires LARK_TEST_SHEET_ID + LARK_TEST_SHEET_RANGE)
  - [x] `sheets append` (gated integration test; requires LARK_TEST_SHEET_ID + LARK_TEST_SHEET_RANGE)
  - [x] `sheets clear` (gated integration test; requires LARK_TEST_SHEET_ID + LARK_TEST_SHEET_RANGE)
- [x] `mail send` (gated integration test; requires user token + LARK_TEST_MAIL_TO)

Milestone 3:
- [ ] meetings/minutes as available
- [x] base record create/update/delete after base lands
- [x] wiki search (gated integration test added; requires user token)

Acceptance criteria:
- `LARK_INTEGRATION=1 go test ./...` covers core commands and proves flag behavior matches expectations.

---

### I) Platform convenience: `--platform`

Goal: Feishu/Lark base domain should be painless to configure and consistent across *all* API calls.

Design principles:
- Separate **runtime override** vs **persistent config** (avoid surprising config mutations).
- Precedence: `--base-url` > `--platform` > `config.base_url` > default feishu.

Deliverables:
1) Persistent configuration
- [x] `lark auth login --platform feishu|lark` persists mapped `base_url` to config
  - [x] keep `--base-url` as highest override
  - [x] tests cover platform mapping + base-url override priority

- [x] Add explicit persistent commands:
  - [x] `lark auth platform set feishu|lark`
  - [x] `lark auth platform info`

2) Runtime override (global)
- [x] Add global flags on root (apply to every command that makes API calls):
  - [x] `lark --platform feishu|lark ...` (runtime only)
  - [x] `lark --base-url <url> ...` (runtime only)
- [x] Implement without auto-writing config.

3) Docs + tests
- [x] Document platform/base-url precedence in README.
- [x] Add unit tests verifying global flag override behavior.

---

## 4) Operational items (must keep working)

- [x] Keep auto-advance cronjob aligned with this backlog.
- [x] Always keep repo clean (avoid untracked leftovers that block pulls).
- [x] Build fix: restored `a1RangeShape` helper used by Sheets clear fallback (commit db5c339)
- [x] Build fix: removed duplicate `ListSpreadsheetSheets` implementation (commit 6891bae)

---

## 5) Change log (backlog edits)

- 2026-01-30: Created detailed backlog as source of truth; expanded Auth/SDK/Mail/Sheets/Base/Wiki/Integration/Platform items.
- 2026-01-30: Drive SDK migration progress: urls/share/search/list moved onto `internal/larksdk` (no legacy client usage).
- 2026-01-31: Added `lark auth user login` (OAuth auth-code flow) + config storage + unit tests.
- 2026-01-31: Added backlog item: Docs Markdown ⇄ Docx bidirectional sync (export md exists; import/sync TBD).
- 2026-01-31: Marked runtime base-url/platform override tests/docs complete; updated Lark base URL to `open.larksuite.com`.
- 2026-01-31: Marked `lark config list-keys` complete.
- 2026-01-31: Started gog-style auth service registry (`internal/authregistry`) + stable-sorted scope union + unit tests.
- 2026-01-31: Added unit tests for drive search pagination capping (`--pages` + `--limit`).
- 2026-01-31: Added `make it` helper target for running all integration tests (gated by `LARK_INTEGRATION=1`).
