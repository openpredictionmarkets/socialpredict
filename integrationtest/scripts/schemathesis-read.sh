#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
BASE_URL="${BASE_URL:-http://localhost:8080}"
SCHEMA="${SCHEMA:-$ROOT/backend/docs/openapi.yaml}"
STAMP="$(date -u +%Y%m%dT%H%M%SZ)"
OUT="${OUT:-$ROOT/integrationtest/artifacts/schemathesis-read-$STAMP}"
MAX_EXAMPLES="${MAX_EXAMPLES:-5}"
REPORT="${REPORT:-junit}"
PHASES="${PHASES:-coverage}"
READ_PATHS="${READ_PATHS:-/v0/setup /v0/setup/frontend /v0/stats /v0/market-tags}"

if ! command -v schemathesis >/dev/null 2>&1; then
  echo "schemathesis is required. Install it first, e.g. pipx install schemathesis" >&2
  exit 127
fi

mkdir -p "$OUT"

args=(
  run "$SCHEMA"
  --url "$BASE_URL"
  --checks not_a_server_error,status_code_conformance,content_type_conformance,response_schema_conformance
  --phases "$PHASES"
  --max-examples "$MAX_EXAMPLES"
  --mode positive
  --generation-database none
  --request-timeout 5
  --rate-limit 5/s
  --report "$REPORT"
  --report-dir "$OUT"
  --report-junit-path "$OUT/junit.xml"
  --no-color
)

if [[ -n "${AUTH_TOKEN:-}" ]]; then
  args+=(--header "Authorization:Bearer $AUTH_TOKEN")
fi

for path in $READ_PATHS; do
  args+=(--include-path "$path")
done

echo "Schemathesis read-only API run"
echo "Base URL: $BASE_URL"
echo "Schema: $SCHEMA"
echo "Read paths: $READ_PATHS"
echo "Phases: $PHASES"
echo "Artifacts: $OUT"
schemathesis "${args[@]}"
