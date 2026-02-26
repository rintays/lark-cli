#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

MODE="full"
if [[ "${1:-}" == "--cron" ]]; then
  MODE="cron"
fi

now_rfc3339() {
  date -u +"%Y-%m-%dT%H:%M:%SZ"
}

have() {
  command -v "$1" >/dev/null 2>&1
}

section() {
  echo
  echo "## $1"
}

kv() {
  # key, value
  printf -- "- %s: %s\n" "$1" "$2"
}

REPORT_DIR="$HOME/.openclaw/lark-cli"
mkdir -p "$REPORT_DIR"

stamp="$(date -u +"%Y%m%dT%H%M%SZ")"
report="$REPORT_DIR/maintainer-sweep-$stamp.md"
latest="$REPORT_DIR/maintainer-sweep-latest.md"

# Generate report into a buffer file first.
TMP_REPORT="$(mktemp -t lark-cli-sweep.XXXXXX)"
trap 'rm -f "$TMP_REPORT"' EXIT

{
  echo "# lark-cli maintainer sweep"
  kv "time_utc" "$(now_rfc3339)"
  kv "repo" "$REPO_DIR"

section "git"
# Avoid failing if no remotes.
if git rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  git remote -v | sed 's/^/- remote: /' || true
  git fetch --all --prune --tags >/dev/null 2>&1 || true

  branch="$(git rev-parse --abbrev-ref HEAD 2>/dev/null || echo "")"
  kv "branch" "${branch:-unknown}"

  dirty="$(git status --porcelain | wc -l | tr -d ' ')"
  kv "dirty_files" "$dirty"
  if [[ "$dirty" != "0" ]]; then
    echo "- dirty_preview:"
    git status --porcelain | head -n 30 | sed 's/^/  - /'
  fi

  if git show-ref --verify --quiet refs/remotes/origin/main; then
    counts="$(git rev-list --left-right --count origin/main...HEAD 2>/dev/null || echo "")"
    behind="${counts%%\t*}"; ahead="${counts##*\t}"
    kv "ahead_of_origin_main" "${ahead:-?}"
    kv "behind_origin_main" "${behind:-?}"
  else
    kv "origin_main" "not found"
  fi

  latest_tag="$(git tag --sort=-creatordate | head -n 1 || true)"
  kv "latest_tag" "${latest_tag:-none}"
  kv "head" "$(git rev-parse --short HEAD)"
else
  kv "status" "not a git repo"
fi

section "go tests"
if have go; then
  kv "go_version" "$(go version | sed 's/^go version //')"
  # Fast, deterministic check.
  if go test ./... >/tmp/lark_cli_sweep_go_test.log 2>&1; then
    kv "go_test" "ok"
  else
    kv "go_test" "FAIL"
    echo "- go_test_log (tail):"
    tail -n 80 /tmp/lark_cli_sweep_go_test.log | sed 's/^/  /'
  fi
else
  kv "go" "not found"
fi

section "release / brew / local"
if have lark; then
  kv "local_lark_version" "$(lark version 2>/dev/null | tr -d '\n' || true)"
else
  kv "local_lark_version" "lark not found"
fi

if have brew; then
  # Best-effort; homebrew json may not be present on older versions.
  kv "brew" "$(brew --version 2>/dev/null | head -n 1 | tr -d '\n' || true)"
  if brew info rintays/tap/lark --json >/tmp/lark_cli_sweep_brew.json 2>/dev/null; then
    ver="$(jq -r '.[0].installed[0].version // empty' /tmp/lark_cli_sweep_brew.json 2>/dev/null || true)"
    kv "brew_installed_lark" "${ver:-unknown}"
  else
    kv "brew_info" "unavailable"
  fi
else
  kv "brew" "not found"
fi

section "github actions (optional)"
if have gh; then
  # If not logged in, this will error; treat as best-effort.
  if gh run list --repo rintays/lark-cli --limit 5 >/tmp/lark_cli_sweep_gh_runs.txt 2>&1; then
    echo "- recent_runs:"
    sed 's/^/  /' /tmp/lark_cli_sweep_gh_runs.txt
  else
    kv "gh" "available but cannot query runs (not logged in / no access?)"
  fi
else
  kv "gh" "not found"
fi

section "issue watchlist (optional)"
# Keep a tiny watchlist of known-impact issues so the cron job can alert until resolved.
# Checks only; no mutations.
WATCH_ISSUES=(9)
if have gh; then
  for n in "${WATCH_ISSUES[@]}"; do
    if gh issue view "$n" --repo rintays/lark-cli --json number,state,title,url >/tmp/lark_cli_sweep_issue.json 2>/dev/null; then
      state=$(jq -r '.state' </tmp/lark_cli_sweep_issue.json 2>/dev/null || echo "")
      title=$(jq -r '.title' </tmp/lark_cli_sweep_issue.json 2>/dev/null || echo "")
      url=$(jq -r '.url' </tmp/lark_cli_sweep_issue.json 2>/dev/null || echo "")
      echo "- issue_$n: ${state:-unknown} — ${title:-} (${url:-})"
    else
      echo "- issue_$n: unable to query"
    fi
  done
else
  kv "gh" "not found"
fi
} >"$TMP_REPORT"

# Persist full report for AG.
cp -f "$TMP_REPORT" "$report"
cp -f "$report" "$latest"

# If not cron mode, print the full report.
if [[ "$MODE" == "full" ]]; then
  cat "$report"
  exit 0
fi

# Cron mode: be silent unless issues are detected.
dirty=$(grep -E '^- dirty_files:' "$report" | awk '{print $3}' | tr -d ' ' || echo "")
behind_line=$(grep -E '^- behind_origin_main:' "$report" || true)
behind=$(echo "$behind_line" | awk '{print $3}' | tr -d ' ' || echo "")
go_test=$(grep -E '^- go_test:' "$report" | awk '{print $3}' | tr -d ' ' || echo "")

issues=()
if [[ -n "$dirty" && "$dirty" != "0" ]]; then issues+=("dirty_worktree=$dirty"); fi
if [[ -n "$behind" && "$behind" != "0" && "$behind" != "?" ]]; then issues+=("behind_origin_main=$behind"); fi
if [[ -n "$go_test" && "$go_test" != "ok" ]]; then issues+=("go_test=$go_test"); fi

# Issue watchlist alerts (best-effort): if watched issue is still open.
if have gh; then
  for n in 9; do
    if gh issue view "$n" --repo rintays/lark-cli --json state,url >/tmp/lark_cli_sweep_issue_state.json 2>/dev/null; then
      st=$(jq -r '.state' </tmp/lark_cli_sweep_issue_state.json 2>/dev/null || echo "")
      url=$(jq -r '.url' </tmp/lark_cli_sweep_issue_state.json 2>/dev/null || echo "")
      if [[ "$st" == "OPEN" ]]; then
        issues+=("watch_issue_${n}=OPEN ($url)")
      fi
    fi
  done
fi

if [[ ${#issues[@]} -eq 0 ]]; then
  exit 0
fi

echo "lark-cli maintainer sweep: issues detected"
for it in "${issues[@]}"; do
  echo "- $it"
done
echo "- full_report: $latest"
exit 0
