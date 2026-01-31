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
