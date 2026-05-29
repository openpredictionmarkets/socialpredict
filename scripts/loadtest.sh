#!/usr/bin/env bash

set -euo pipefail

[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

print_loadtest_help() {
  cat <<'EOF'
Usage: ./SocialPredict load COMMAND [OPTIONS]

Commands:
  seed    Create/reset guarded load-test users, moderators, markets, and fixture CSVs
  help    Show this help

Examples:
  LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --hot-markets 1 --reset
  LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 50000 --moderators 100 --markets 1000 --hot-markets 10 --reset
EOF
}

print_loadtest_seed_help() {
  cat <<'EOF'
Usage: ./SocialPredict load seed [OPTIONS]

Create or reset load-test fixtures through the normal database.

Safety:
  This command refuses to run unless LOAD_TEST_ENABLED=true.
  This command refuses APP_ENV=production unless LOAD_TEST_ALLOW_PRODUCTION=true.
  Generated fixture CSVs are written to loadtest/fixtures/ and are gitignored.

Created/updated data:
  loaduser000001...       REGULAR users with must_change_password=false
  loadmod000001...        MODERATOR users with active moderator status
  Load Test Market ...    published markets owned by load-test moderators

Defaults:
  password: loadtest-password
  users: 10
  moderators: 2
  markets: 5
  hot markets: 1
  user prefix: loaduser
  moderator prefix: loadmod
  user balance: 1000000

Options:
  --users N              Number of REGULAR test users, 1-50000
  --moderators N         Number of MODERATOR test users, 1-10000
  --markets N            Number of published fixture markets, 1-50000
  --hot-markets N        Number of fixture markets marked hot in markets.csv
  --password VALUE       Password for every load-test user
  --user-prefix VALUE    Prefix for regular users, e.g. loaduser
  --moderator-prefix VALUE
                          Prefix for moderators, e.g. loadmod
  --user-balance N       Starting balance for each load-test user
  --resolution-days N    Days until generated markets resolve, 1-3650
  --bcrypt-cost N        Password hash cost, 4-31; defaults low for bulk fixtures
  --reset                Remove existing prefixed fixture users, markets, and bets first
  --help                 Show this help

Examples:
  LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 10 --moderators 2 --markets 5 --reset
  LOAD_TEST_ENABLED=true ./SocialPredict load seed --users 50000 --moderators 100 --markets 1000 --hot-markets 10 --reset
EOF
}

require_integer() {
  local label="$1"
  local value="$2"
  if ! [[ "${value}" =~ ^[0-9]+$ ]]; then
    print_error "${label} must be a non-negative integer."
    exit 1
  fi
}

loadtest_compose_file() {
  case "${APP_ENV:-development}" in
    development)
      echo "${SCRIPT_DIR}/scripts/docker-compose-dev.yaml"
      ;;
    localhost)
      echo "${SCRIPT_DIR}/scripts/docker-compose-local.yaml"
      ;;
    production)
      echo "${SCRIPT_DIR}/scripts/docker-compose-prod.yaml"
      ;;
    *)
      return 1
      ;;
  esac
}

resolve_loadtest_backend_container() {
  local fixed_name="${BACKEND_CONTAINER_NAME:-}"
  local compose_file=""
  local container_id=""

  if [ -n "${fixed_name}" ] && docker ps --format '{{.Names}}' | grep -qx "${fixed_name}"; then
    echo "${fixed_name}"
    return 0
  fi

  compose_file="$(loadtest_compose_file || true)"
  if [ -n "${compose_file}" ] && [ -f "${compose_file}" ]; then
    container_id="$(docker compose --env-file "${SCRIPT_DIR}/.env" -f "${compose_file}" ps -q backend 2>/dev/null | head -n 1 || true)"
    if [ -n "${container_id}" ] && [ "$(docker inspect -f '{{.State.Running}}' "${container_id}" 2>/dev/null || true)" = "true" ]; then
      echo "${container_id}"
      return 0
    fi
  fi

  container_id="$(docker ps \
    --filter "label=com.docker.compose.service=backend" \
    --format '{{.ID}}' | head -n 1 || true)"
  if [ -n "${container_id}" ]; then
    echo "${container_id}"
    return 0
  fi

  return 1
}

run_loadtest_seed() {
  local users="10"
  local moderators="2"
  local markets="5"
  local hot_markets="1"
  local password="loadtest-password"
  local user_prefix="loaduser"
  local moderator_prefix="loadmod"
  local user_balance=""
  local resolution_days="30"
  local bcrypt_cost=""
  local reset="false"

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --users)
        shift
        users="${1:-}"
        ;;
      --moderators)
        shift
        moderators="${1:-}"
        ;;
      --markets)
        shift
        markets="${1:-}"
        ;;
      --hot-markets)
        shift
        hot_markets="${1:-}"
        ;;
      --password)
        shift
        password="${1:-}"
        ;;
      --user-prefix)
        shift
        user_prefix="${1:-}"
        ;;
      --moderator-prefix)
        shift
        moderator_prefix="${1:-}"
        ;;
      --user-balance)
        shift
        user_balance="${1:-}"
        ;;
      --resolution-days)
        shift
        resolution_days="${1:-}"
        ;;
      --bcrypt-cost)
        shift
        bcrypt_cost="${1:-}"
        ;;
      --reset)
        reset="true"
        ;;
      --help|-h|help)
        print_loadtest_seed_help
        return 0
        ;;
      *)
        print_error "Unknown load seed option: $1"
        print_loadtest_seed_help
        exit 1
        ;;
    esac
    shift
  done

  require_integer "--users" "${users}"
  require_integer "--moderators" "${moderators}"
  require_integer "--markets" "${markets}"
  require_integer "--hot-markets" "${hot_markets}"
  require_integer "--resolution-days" "${resolution_days}"
  if [ -n "${user_balance}" ]; then
    require_integer "--user-balance" "${user_balance}"
  fi
  if [ -n "${bcrypt_cost}" ]; then
    require_integer "--bcrypt-cost" "${bcrypt_cost}"
  fi

  if [ "${LOAD_TEST_ENABLED:-false}" != "true" ]; then
    print_error "Refusing to run load seed unless LOAD_TEST_ENABLED=true is set."
    exit 1
  fi

  local backend_container
  if ! backend_container="$(resolve_loadtest_backend_container)"; then
    print_error "Could not find a running backend service container. Start the app first with './SocialPredict up'."
    exit 1
  fi

  local network_name
  network_name="$(docker inspect -f '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' "${backend_container}" | head -n 1)"
  if [ -z "${network_name}" ]; then
    print_error "Could not determine Docker network for backend container '${backend_container}'."
    exit 1
  fi

  local platform_args=()
  if [ -n "${FORCE_PLATFORM:-}" ]; then
    platform_args=(--platform "${FORCE_PLATFORM}")
  fi

  local optional_env=()
  if [ -n "${user_balance}" ]; then
    optional_env+=(-e "LOAD_TEST_USER_BALANCE=${user_balance}")
  fi
  if [ -n "${bcrypt_cost}" ]; then
    optional_env+=(-e "LOAD_TEST_BCRYPT_COST=${bcrypt_cost}")
  fi

  print_status "Seeding load-test fixtures through backend container '${backend_container}' on Docker network '${network_name}' ..."

  docker run --rm \
    "${platform_args[@]}" \
    --network "${network_name}" \
    -v "socialpredict_loadtest_go_mod:/go/pkg/mod" \
    -v "socialpredict_loadtest_go_build:/root/.cache/go-build" \
    -v "${SCRIPT_DIR}:/workspace" \
    -w /workspace/backend \
    -e "APP_ENV=${APP_ENV}" \
    -e "DB_HOST=db" \
    -e "DB_PORT=5432" \
    -e "POSTGRES_USER=${POSTGRES_USER}" \
    -e "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
    -e "POSTGRES_DATABASE=${POSTGRES_DATABASE}" \
    -e "DB_REQUIRE_TLS=${DB_REQUIRE_TLS:-false}" \
    -e "DB_SSLMODE=${DB_SSLMODE:-disable}" \
    -e "LOAD_TEST_ENABLED=${LOAD_TEST_ENABLED:-false}" \
    -e "LOAD_TEST_ALLOW_PRODUCTION=${LOAD_TEST_ALLOW_PRODUCTION:-false}" \
    -e "LOAD_TEST_RESET=${reset}" \
    -e "LOAD_TEST_USER_COUNT=${users}" \
    -e "LOAD_TEST_MODERATOR_COUNT=${moderators}" \
    -e "LOAD_TEST_MARKET_COUNT=${markets}" \
    -e "LOAD_TEST_HOT_MARKET_COUNT=${hot_markets}" \
    -e "LOAD_TEST_PASSWORD=${password}" \
    -e "LOAD_TEST_USER_PREFIX=${user_prefix}" \
    -e "LOAD_TEST_MODERATOR_PREFIX=${moderator_prefix}" \
    -e "LOAD_TEST_FIXTURE_DIR=/workspace/loadtest/fixtures" \
    -e "LOAD_TEST_RESOLUTION_DAYS=${resolution_days}" \
    "${optional_env[@]}" \
    golang:1.25-alpine \
    sh -lc "export PATH=/usr/local/go/bin:/go/bin:\$PATH; go run ./cmd/loadtestseed"
}

command="${1:-help}"
case "${command}" in
  seed)
    shift
    run_loadtest_seed "$@"
    ;;
  --help|-h|help)
    print_loadtest_help
    ;;
  *)
    print_error "Unknown load command: ${command}"
    print_loadtest_help
    exit 1
    ;;
esac
