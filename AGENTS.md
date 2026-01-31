# Repository Guidelines

## Project Structure & Module Organization
- `cmd/lark`: CLI entry point and command implementations, organized by domain (e.g., `users`, `drive`, `docs`, `mail`). Add new commands under this tree.
- `internal/`: core packages (`config`, `larksdk`, `output`, `version`). Keep HTTP details and SDK adapters here, not in command files.
- Docs & design: `README.md`, `DESIGN.md` describe usage, constraints, and planned work.
- Tests live alongside code as `*_test.go` under `cmd/lark` and `internal`.
- Dependencies are managed via `go.mod` and `go.sum`.

## Build, Test, and Development Commands
- Build the CLI: `go build -o lark ./cmd/lark`.
- View help: `./lark --help` or `./lark users --help`.
- Run all tests: `go test ./...`.
- Run a single test: `go test ./cmd/lark -run TestChatsList`.
- Inspect package docs: `go doc ./internal/larksdk`.
- Race checks (slower): `go test -race ./...`.

## Coding Style & Naming Conventions
- Language: Go. Run `gofmt` before committing.
- Package names are short and lowercase; exported symbols use CamelCase (e.g., `GetTenantToken`).
- Command files and flags should align with their domain (`users_*`, `drive_*`).
- Preserve dual output modes: human-readable by default, scriptable with `--json`.
- Errors should include context (command, ID type, key parameters).

## CLI Argument Principles
- Required resource identifiers should be positional args (e.g., `lark docs info <doc-id>`), not mandatory `--id` flags.
- Do not keep legacy `--id` style flags for required identifiers.
- Exceptions: `base --app-token` and `wiki --space-id` remain flags.

## Testing Guidelines
- Use Go’s standard `testing` package.
- Keep tests in `*_test.go` beside the package under test.
- Avoid network calls and real credentials in unit tests; add integration tests only with an explicit env gate.
- Cover CLI parsing, JSON output, and error messaging.

## Commit & Pull Request Guidelines
- Commit messages commonly use `scope: action` or a conventional prefix (e.g., `users: migrate list to larksdk`, `feat: add core lark CLI`).
- PRs should include a concise description and the test commands run (e.g., `go test ./...`).
- If CLI behavior changes, update `README.md` and any relevant docs.
- Keep commits small and focused.

## Local Run & Output Expectations
- Validate `--json` output when changes affect CLI output.
- Example: `./lark users list --limit 5 --json`.
- Breaking changes must document migration steps in the PR.

## Security & Configuration Tips
- Default config path: `~/.config/lark/config.json`.
- Common env vars: `LARK_APP_ID`, `LARK_APP_SECRET`, `LARK_USER_ACCESS_TOKEN`.
- Never commit real credentials or tokens.

## Agent Notes
- 2026-01-30: Unified CLI command naming (messages/contacts users), moved required flag validation to Cobra MarkFlagRequired/Args, added CLI flag validation tests, updated README examples.
- 2026-01-30: Removed legacy HTTP client, migrated whoami/calendar to SDK, added SDK coverage matrix.
- 2026-01-31: Added user OAuth scope management (scopes/services/readonly/drive-scope/force-consent), user-scope hints for missing permissions, and updated drive/docs/sheets search to require user tokens with `search_key`.
- 2026-01-31: Added verbose tracing for drive/docs/sheets search and capped search pagination with a new `--pages` flag; user-scope hints now skip non-permission errors.
- 2026-01-31: Added docs-api search support for drive/docs/sheets search (offset-based pagination, doc type normalization), keeping folder-scoped drive search on the legacy drive endpoint.
- 2026-01-31: Removed drive search folder flags/legacy drive search path; drive search now consistently uses docs-api with offset window guard, README updated.
- 2026-01-31: Added lipgloss-based output formatting for tabular CLI output with auto-tty detection.
- 2026-01-31: Switched docs cat to docx raw_content API to avoid export task validation errors for txt/md.
- 2026-01-31: Mail user mailbox operations now use user access tokens (list/get/folders/mailbox get) to avoid field validation errors; public mailboxes remain tenant-token.
- 2026-01-31: Mail list now resolves folder_id (defaults to Inbox) to satisfy API validation; SDK enforces folder_id requirement.
- 2026-01-31: Mail folder_type now accepts string/number responses; inbox resolution also checks a zh-CN inbox name fallback.
- 2026-01-31: Mail folders now accept `id` in API responses, mail list resolves system folder aliases, and the folder-id hint text is more explicit.
- 2026-01-31: Added `lark sheets create` command plus README/design coverage updates.
- 2026-01-31: Added base app create/copy/get/update commands (tenant-token only), including time zone/advanced/customized config flags, backed by bitable app API endpoints.
- 2026-01-31: Added message list/search/reply, reactions, pin/unpin, chat create/get/update, and chat announcement get/update; expanded message send options and tests.
- 2026-01-31: Added table headers and separator lines for terminal tabular outputs, with shared table formatting helpers.
- 2026-01-31: Added calendar search/get/update/delete commands, improved primary calendar resolution, and expanded calendar auth metadata.
- 2026-01-31: Added meeting reservation create/update/delete, expanded meeting list filters with required time range validation, and updated README features note.
- 2026-01-31: Unified read-style commands to `info` (docs/sheets/drive/mail/users/wiki/base/config/auth/meetings/minutes) with no legacy aliases; updated tests/docs.
- 2026-01-31: Drive info/urls now honor user tokens with auto fallback and added a user-token drive info test.
- 2026-01-31: Promoted required identifiers to positional args for docs/drive/sheets/base/wiki/minutes/meetings/auth scopes/mailbox set (keeping base app-token and wiki space-id as flags), updated tests and README examples.
- 2026-01-31: Minutes list now filters Drive file listings by type; added minutes delete/update commands backed by Drive delete and permission update endpoints, plus minutes command tests and design doc updates.
- 2026-01-31: Added docx chat announcement fallback (docx API + block listing) and updated CLI output/tests for docx announcements.
- 2026-01-31: Aligned message list/search/reply/reactions/pin commands with positional args for required identifiers; updated tests and README examples.
- 2026-01-31: Expanded docs/sheets/mail info output to full SDK fields with key/value tables, enriched docx/sheets/mail SDK mappings, and updated tests.
- 2026-01-31: Docs info now falls back to drive file metadata for URL when the docx info API omits it.
- 2026-01-31: Refined messages list output to show content-first blocks with message metadata (id/type/chat/time) and text extraction from message bodies.
- 2026-01-31: Messages list now renders system templates, formats mentions with user id links, drops chat id from meta, and shows metadata before content.
- 2026-01-31: Chat announcement docx output now renders text content from announcement blocks in CLI output.
- 2026-01-31: Expanded `lark chats get` output with chat metadata plus member previews (new `--members-limit`/`--members-page-size` flags) and updated README/tests.
- 2026-01-31: Added docx block operations (get/list/update/batch/children/descendant) and markdown convert/overwrite commands with docs/tests updates.
- 2026-01-31: Moved docx convert/overwrite to `lark docs convert/overwrite`, replaced docs cat with docs get (default md + blocks format), and updated docs/tests.
- 2026-01-31: Mail send now supports raw EML input via `--raw`/`--raw-file` (base64url), with updated CLI validation and examples.
- 2026-01-31: Added `lark sheets delete` command backed by Drive delete, with README/design coverage and tests.
- 2026-01-31: Message search now resolves message IDs to full message details for output (content/sender/time), and JSON includes both message IDs and message objects.
- 2026-01-31: Adjusted Sheets dimension payloads for better API compatibility, switched columns to COLUMNS, clarified sheet range help/examples, normalized sheets create folder-id root handling, and clarified user token help/error text for Sheets search.
- 2026-01-31: Added user OAuth account management, keychain token storage backend, account selection via --account/LARK_ACCOUNT, and preflight scope checks for user-token commands; removed legacy user-token fallback; updated config/README accordingly.
- 2026-01-31: Messages send now requires `--receive-id`, chats update uses positional chat-id, and message/wiki search added `--pages` pagination with auto page-size; README updated.
- 2026-01-31: Message search scope hints now include `im:message:get_as_user` alongside `search:message` for user OAuth guidance.
- 2026-01-31: Mail list output now includes sender and internal_date columns for quick triage.
- 2026-01-31: Added Bitable base concept explanations to the `bases` command help and README.
- 2026-01-31: Added product concept explanations to top-level help for users/drive/docs/sheets/wiki/mail/messages/chats/calendars/meetings/minutes/contacts.
- 2026-01-31: Sheets create now includes default sheet id/range, sheets read/update/append/clear accept `--sheet-id` with range shorthand, sheets update/append accept JSON/CSV values files (`--values-file` or `--values @file`), and sheets read defaults missing major_dimension to ROWS.
- 2026-01-31: Added `lark mail get` for full message content; `mail info`/`mail list` now return metadata-only message fields.
- 2026-01-31: Fixed Bitable record create payloads, added repeatable `--field` inputs and error detail formatting, added base field create command, and added view-name fallback for base table create.
- 2026-01-31: Corrected Mail user OAuth scopes to include field-level read permissions (`mail:user_mailbox.message.subject:read`, `mail:user_mailbox.message.address:read`, `mail:user_mailbox.message.body:read`) plus `mail:user_mailbox.message:readonly`, added mail-send/public service mappings, and updated auth registry tests/docs.
- 2026-01-31: Removed base table `--view-name`, added base view create command, and added base field types hints aligned to field properties.
- 2026-01-31: Made meetings list and calendars list/search accept optional start/end filters, removed required flags, and added calendar pagination to respect `--limit`.
- 2026-01-31: Added mail message-id normalization retry on invalid params for mail get, with tests.
- 2026-01-31: Meetings list now defaults to the last 6 months when start/end are omitted.
- 2026-01-31: Sheets rows/cols delete now use the v2 dimension_range delete API with 1-based index conversion in payloads, fixing delete failures.
- 2026-01-31: Calendar event outputs now include status in list/search/get/create/update tables and SDK model mapping.
- 2026-01-31: Expanded base record create/update help text with detailed value format examples for record fields.
- 2026-01-31: Clarified wiki help text (space/node/task definitions) and corrected meetings help to reflect the default list time range.
- 2026-01-31: Clarified docs help text to describe document blocks.
- 2026-01-31: Added wiki space create command (v2) with name/visibility/type settings.
