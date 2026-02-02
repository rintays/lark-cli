# Proposal: AI Agent Skills Documentation

## Summary

Add a lightweight, agent-focused documentation layer for this repo so AI agents without Feishu/Lark context can become productive quickly. The primary entry is `skills/lark-cli/SKILL.md` (an agent README). One-time or deep setup content is split into `skills/lark-cli/references/`.

## Goals

- Give an agent a 3–5 minute ramp-up path.
- Keep `SKILL.md` short and actionable.
- Separate low-frequency or one-time setup into dedicated files.
- Provide stable, copy-paste-friendly commands with placeholders.

## Non-goals

- Replace `README.md` or product docs.
- Provide exhaustive API coverage (that belongs in SDK/docs).

## Proposed structure

- `skills/lark-cli/SKILL.md` (entry point, quickstart, common patterns)
- `skills/lark-cli/references/INSTALL.md` (install/build methods)
- `skills/lark-cli/references/AUTH.md` (tenant vs user token, scopes)
- `skills/lark-cli/references/CONCEPTS.md` (Lark/Feishu primer, IDs, URLs)
- `skills/lark-cli/references/RECIPES.md` (common agent tasks)
- `skills/lark-cli/references/TROUBLESHOOTING.md` (errors/scopes/IDs)
- `skills/lark-cli/references/COMPLETION.md` (Shell completion)
- `skills/lark-cli/references/DOCS.md` (Docs workflows)
- `skills/lark-cli/references/SHEETS.md` (Sheets workflows)
- `skills/lark-cli/references/BASES.md` (Bitable workflows)
- `skills/lark-cli/references/DRIVE.md` (Drive workflows)
- `skills/lark-cli/references/MINUTES.md` (Minutes workflows)
- `skills/lark-cli/references/CALENDARS.md` (Calendars workflows)
- `skills/lark-cli/references/MEETINGS.md` (Meetings workflows)
- `skills/lark-cli/references/CHATS.md` (Chats workflows)
- `skills/lark-cli/references/MESSAGES.md` (Messages workflows)
- `skills/lark-cli/references/CONTACTS.md` (Contacts workflows)
- `skills/lark-cli/references/MAIL.md` (Mail workflows)
- `skills/lark-cli/references/TASKLISTS.md` (Tasklists workflows)
- `skills/lark-cli/references/TASKS.md` (Tasks workflows)
- `skills/lark-cli/references/USERS.md` (Users workflows)
- `skills/lark-cli/references/CONFIG.md` (Config workflows)
- `skills/lark-cli/references/WHOAMI.md` (Whoami command)
- `skills/lark-cli/references/WIKI.md` (Wiki workflows)

## Draft content guidelines

`SKILL.md` should include:

- What this repo provides (CLI scope)
- 3-step quickstart (install, auth, run)
- Token rules (tenant vs user)
- Command model (product/action, positional IDs)
- `--json` usage and stdout/stderr expectations
- Links to deep docs in `references/`

Separate files should:

- Avoid duplication with README (link when possible)
- Be short and task-focused
- Use placeholders (`<APP_ID>`, `<DOC_TOKEN>`) and safe examples

## Maintenance

- Update `SKILL.md` whenever CLI behavior changes or new products are added.
- Keep `references/RECIPES.md` aligned with supported commands.
- Avoid leaking real IDs or credentials.

## Rollout plan

1) Add the proposed files with a minimal but complete set of examples.
2) Iterate as CLI behavior evolves (update AGENTS.md with notable changes).
3) Optionally add links from `README.md` if discoverability is an issue.
