#!/usr/bin/env bash
# Fails if the number of public/app files importing `connect` from `react-redux`
# exceeds the committed baseline. The baseline may only decrease (ratchet).
# See docs/refactors/remove-react-redux-connect/plan.md
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
BASELINE_FILE="${ROOT}/scripts/ci/react-redux-connect-baseline"
APP_DIR="${ROOT}/public/app"

if [[ ! -f "${BASELINE_FILE}" ]]; then
  echo "error: missing baseline file at ${BASELINE_FILE}" >&2
  exit 1
fi

baseline="$(tr -d '[:space:]' < "${BASELINE_FILE}")"
if [[ ! "${baseline}" =~ ^[0-9]+$ ]]; then
  echo "error: baseline must be a non-negative integer (got: ${baseline})" >&2
  exit 1
fi

# Portable equivalent of the plan ratchet (ripgrep). Counts files, not lines.
count="$(
  grep -rl --include='*.ts' --include='*.tsx' \
    -E "import[[:space:]]*\{[^}]*\bconnect\b[^}]*\}[[:space:]]*from[[:space:]]*['\"]react-redux['\"]" \
    "${APP_DIR}" \
    | wc -l \
    | tr -d '[:space:]'
)"

echo "react-redux connect imports: ${count} (baseline ${baseline})"

if (( count > baseline )); then
  echo "error: connect-import count increased (${count} > ${baseline})." >&2
  echo "New \`connect\` usage from react-redux is not allowed while the" >&2
  echo "remove-react-redux-connect refactor is in progress." >&2
  echo "Migrate with useSelector/useDispatch from app/types/store instead." >&2
  echo "If you completed a migration PR, lower scripts/ci/react-redux-connect-baseline." >&2
  exit 1
fi

if (( count < baseline )); then
  echo "note: count is below baseline — update ${BASELINE_FILE} to ${count} in this PR."
fi

exit 0
