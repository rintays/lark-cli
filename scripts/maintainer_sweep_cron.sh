#!/usr/bin/env bash
set -euo pipefail

REPO_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
REPORT_DIR="$HOME/.openclaw/lark-cli"
mkdir -p "$REPORT_DIR"

stamp="$(date -u +"%Y%m%dT%H%M%SZ")"
report="$REPORT_DIR/maintainer-sweep-$stamp.md"
latest="$REPORT_DIR/maintainer-sweep-latest.md"

# Always generate the full report for AG to read later.
"$REPO_DIR/scripts/maintainer_sweep.sh" >"$report"
cp -f "$report" "$latest"

# Decide whether to alert Master.
# Criteria (tuneable): dirty tree, behind origin/main, go test fail.

dirty=$(grep -E '^- dirty_files:' "$report" | awk '{print $3}' | tr -d ' ' || echo "")
behind_line=$(grep -E '^- behind_origin_main:' "$report" || true)
behind=$(echo "$behind_line" | awk '{print $3}' | tr -d ' ' || echo "")

go_test=$(grep -E '^- go_test:' "$report" | awk '{print $3}' | tr -d ' ' || echo "")

issues=()
if [[ -n "$dirty" && "$dirty" != "0" ]]; then issues+=("dirty_worktree=$dirty"); fi
if [[ -n "$behind" && "$behind" != "0" && "$behind" != "?" ]]; then issues+=("behind_origin_main=$behind"); fi
if [[ -n "$go_test" && "$go_test" != "ok" ]]; then issues+=("go_test=$go_test"); fi

if [[ ${#issues[@]} -eq 0 ]]; then
  # Silent success: no output.
  exit 0
fi

# Alert output (short). Keep it compact; full report is saved to $latest.
echo "lark-cli maintainer sweep: issues detected"
for it in "${issues[@]}"; do
  echo "- $it"
done
echo "- full_report: $latest"
