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
- Keep legacy flags as optional aliases when possible to avoid breaking users.
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
- 2026-01-31: Added meeting reservation create/update/delete, expanded meeting list filters with required time range validation, and updated README features note.
- 2026-01-31: Unified read-style commands to `info` (docs/sheets/drive/mail/users/wiki/base/config/auth/meetings/minutes) with no legacy aliases; updated tests/docs.
- 2026-01-31: Promoted required identifiers to positional args for docs/drive/sheets/base/wiki/minutes/meetings/auth scopes/mailbox set (keeping base app-token and wiki space-id as flags), updated tests and README examples.
- 2026-01-31: Minutes list now filters Drive file listings by type; added minutes delete/update commands backed by Drive delete and permission update endpoints, plus minutes command tests and design doc updates.
- 2026-01-31: Aligned message list/search/reply/reactions/pin commands with positional args for required identifiers; updated tests and README examples.
