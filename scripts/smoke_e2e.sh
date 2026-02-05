#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

LOG_DIR="$(mktemp -d)"
LOG_FILE="$LOG_DIR/smoke_e2e.log"
touch "$LOG_FILE"

cleanup_notes=()
cleanup_sheets=()
doc_id=""

log() {
  printf '[%s] %s\n' "$(date -u +%Y-%m-%dT%H:%M:%SZ)" "$*" | tee -a "$LOG_FILE"
}

run_cmd() {
  local -a cmd=("$@")
  log "RUN: ${cmd[*]}"
  if ! "${cmd[@]}" >>"$LOG_FILE" 2>&1; then
    log "FAILED: ${cmd[*]}"
    return 1
  fi
}

run_json() {
  local -a cmd=("$@")
  local out
  log "RUN: ${cmd[*]}"
  if ! out="$("${cmd[@]}" 2>>"$LOG_FILE")"; then
    log "FAILED: ${cmd[*]}"
    return 1
  fi
  printf '%s\n' "$out" >>"$LOG_FILE"
  printf '%s' "$out"
}

run_allow_fail() {
  local -a cmd=("$@")
  log "RUN (allowed failure): ${cmd[*]}"
  if ! "${cmd[@]}" >>"$LOG_FILE" 2>&1; then
    log "EXPECTED FAILURE: ${cmd[*]}"
    return 0
  fi
}

if command -v python3 >/dev/null 2>&1; then
  PYTHON_BIN="python3"
elif command -v python >/dev/null 2>&1; then
  PYTHON_BIN="python"
else
  log "python is required for JSON parsing"
  exit 1
fi

json_get() {
  local key="$1"
  "$PYTHON_BIN" - "$key" <<'PY'
import json
import sys

key = sys.argv[1]
data = json.load(sys.stdin)
cur = data
for part in key.split("."):
    if isinstance(cur, dict) and part in cur:
        cur = cur[part]
    elif isinstance(cur, list):
        try:
            idx = int(part)
        except ValueError:
            print("")
            sys.exit(0)
        if idx < 0 or idx >= len(cur):
            print("")
            sys.exit(0)
        cur = cur[idx]
    else:
        print("")
        sys.exit(0)
if isinstance(cur, (dict, list)):
    print(json.dumps(cur))
else:
    print(cur)
PY
}

cleanup() {
  local exit_code=$?

  # Preserve the original exit code and avoid errexit inside cleanup.
  trap - EXIT
  set +e

  if [[ ${#cleanup_sheets[@]} -gt 0 ]]; then
    for token in "${cleanup_sheets[@]}"; do
      local -a cmd=(./lark --force sheets delete "$token")
      log "CLEANUP: ${cmd[*]}"
      if ! "${cmd[@]}" >>"$LOG_FILE" 2>&1; then
        log "CLEANUP failed for sheet: $token"
        cleanup_notes+=("sheet_token:$token")
      fi
    done
  fi
  if [[ -n "$doc_id" ]]; then
    cleanup_notes+=("docx_document_id:$doc_id")
  fi
  if [[ ${#cleanup_notes[@]} -gt 0 ]]; then
    log "Resources not cleaned up:"
    for note in "${cleanup_notes[@]}"; do
      log "  $note"
    done
  fi

  log "Log file: $LOG_FILE"

  # In CI, the temp file isn't accessible after the job ends, so print a redacted
  # tail when failing to make debugging possible.
  if [[ $exit_code -ne 0 ]]; then
    printf '\n---- smoke-e2e log tail (last 200 lines; redacted) ----\n' >&2
    tail -n 200 "$LOG_FILE" 2>/dev/null | sed -E \
      -e 's/(Bearer )[A-Za-z0-9._-]+/\1[REDACTED]/g' \
      -e 's/(Authorization: )[A-Za-z]+ [A-Za-z0-9._-]+/\1[REDACTED]/g' \
      -e 's/(tenant_access_token"?:")[^"]+"/\1[REDACTED]"/gi' \
      -e 's/(user_access_token"?:")[^"]+"/\1[REDACTED]"/gi' \
      -e 's/(refresh[_-]?token"?:")[^"]+"/\1[REDACTED]"/gi' \
      -e 's/(access[_-]?token"?:")[^"]+"/\1[REDACTED]"/gi' \
      >&2
    printf '---- end ----\n' >&2
  fi

  exit "$exit_code"
}

trap cleanup EXIT

run_cmd go build -o lark ./cmd/lark

auth_out=""
auth_code=0
set +e
auth_out="$(./lark auth user status 2>&1)"
auth_code=$?
set -e
printf '%s\n' "$auth_out" >>"$LOG_FILE"
if [[ $auth_code -ne 0 ]]; then
  log "Skipping smoke E2E: user auth status unavailable (missing credentials?)"
  exit 0
fi
if printf '%s\n' "$auth_out" | grep -qE 'refresh_token_present: false|user_access_token_present: false'; then
  log "Skipping smoke E2E: missing user credentials (refresh/user token not present)"
  exit 0
fi

DOC_TITLE="Smoke Doc $(date -u +%Y%m%dT%H%M%SZ)"
doc_json="$(run_json ./lark --json docs create "$DOC_TITLE")"
doc_id="$(printf '%s' "$doc_json" | json_get "document.document_id")"
if [[ -z "$doc_id" ]]; then
  log "docs create missing document_id"
  exit 1
fi

run_cmd ./lark --json docs info "$doc_id"
run_cmd ./lark --json docs overwrite "$doc_id" --content "# Smoke Doc\n\nHello from smoke tests." --content-type markdown

DOC_EXPORT_PATH="$LOG_DIR/doc_export.pdf"
run_allow_fail ./lark --verbose docs export "$doc_id" --format pdf --out "$DOC_EXPORT_PATH"

run_cmd ./lark --json drive search "$DOC_TITLE" --limit 1 --pages 1

SHEET_TITLE="Smoke Sheet $(date -u +%Y%m%dT%H%M%SZ)"
sheet_json="$(run_json ./lark --json sheets create --title "$SHEET_TITLE")"
sheet_token="$(printf '%s' "$sheet_json" | json_get "spreadsheet_token")"
sheet_id="$(printf '%s' "$sheet_json" | json_get "sheet_id")"
if [[ -z "$sheet_token" ]]; then
  log "sheets create missing spreadsheet_token"
  exit 1
fi
cleanup_sheets+=("$sheet_token")

run_cmd ./lark --json sheets info "$sheet_token"
if [[ -n "$sheet_id" ]]; then
  sheet_range="${sheet_id}!A1:B1"
else
  sheet_range="A1:B1"
fi
run_cmd ./lark --json sheets update "$sheet_token" "$sheet_range" --values '[["smoke","test"]]'
run_cmd ./lark --json sheets read "$sheet_token" "$sheet_range"

cal_summary="Smoke Event $(date -u +%H%M%S)"
cal_json="$(run_json ./lark --json calendar create --summary "$cal_summary" --start "+1h" --end "+2h")"
event_id="$(printf '%s' "$cal_json" | json_get "event.event_id")"
if [[ -z "$event_id" ]]; then
  log "calendar create missing event_id"
  exit 1
fi
run_cmd ./lark --json calendar get "$event_id"
run_cmd ./lark --force --json calendar delete "$event_id"

meetings_json="$(run_json ./lark --json meetings list --limit 1)"
meeting_no="$(printf '%s' "$meetings_json" | json_get "meetings.0.meeting_no")"
meeting_id="$(printf '%s' "$meetings_json" | json_get "meetings.0.id")"
if [[ -n "$meeting_no" ]]; then
  run_cmd ./lark --json meetings info "$meeting_no"
elif [[ -n "$meeting_id" ]]; then
  run_cmd ./lark --json meetings info "$meeting_id"
else
  log "No meetings found; skipping meetings info"
fi

run_cmd ./lark --json mail folders
mail_list_json="$(run_json ./lark --json mail list --limit 1)"
message_id="$(printf '%s' "$mail_list_json" | json_get "messages.0.message_id")"
if [[ -n "$message_id" ]]; then
  run_cmd ./lark --json mail info "$message_id"
else
  log "No mail messages found; skipping mail info"
fi
