#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$REPO_DIR"

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
