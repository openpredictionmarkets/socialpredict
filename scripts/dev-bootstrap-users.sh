#!/usr/bin/env bash

set -euo pipefail

[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

print_dev_bootstrap_help() {
  cat <<'EOF'
Usage: ./SocialPredict dev-bootstrap-users [OPTIONS]

Create or reset development-only login fixtures.

This command refuses to run unless APP_ENV=development in .env.
It writes to the development database through the dev Docker network.

Created/updated users:
  admin       ADMIN user
  testuser01  active MODERATOR user and fixture market owner
  testuser02  REGULAR user
  ...

Created/updated fixture markets:
  Market A    published, owned/stewarded by testuser01, tagged Category A
  Market B    published, owned/stewarded by testuser01, tagged Category B
  Market C    published, owned/stewarded by testuser01, tagged Category C

Defaults:
  password: password
  user count: 10
  username prefix: testuser
  must_change_password: false

Options:
  --password VALUE   Password for every bootstrapped user
  --count N          Number of REGULAR test users to create, 1-100
  --prefix VALUE     Prefix for regular users, e.g. player -> player01
  --help             Show this help

Example:
  ./SocialPredict dev-bootstrap-users
  ./SocialPredict dev-bootstrap-users --count 20 --prefix player
EOF
}

password="password"
count="10"
prefix="testuser"

while [ "$#" -gt 0 ]; do
  case "$1" in
    --password)
      shift
      password="${1:-}"
      ;;
    --count)
      shift
      count="${1:-}"
      ;;
    --prefix)
      shift
      prefix="${1:-}"
      ;;
    --help|-h|help)
      print_dev_bootstrap_help
      return 0
      ;;
    *)
      echo "Unknown dev-bootstrap-users option: $1"
      print_dev_bootstrap_help
      exit 1
      ;;
  esac
  shift
done

if [ "${APP_ENV:-}" != "development" ]; then
  print_error "Refusing to run dev bootstrap unless APP_ENV=development. Current APP_ENV='${APP_ENV:-unset}'."
  exit 1
fi

if ! [[ "${count}" =~ ^[0-9]+$ ]] || [ "${count}" -lt 1 ] || [ "${count}" -gt 100 ]; then
  print_error "--count must be an integer between 1 and 100."
  exit 1
fi

if ! [[ "${prefix}" =~ ^[a-z][a-z0-9]*$ ]]; then
  print_error "--prefix must start with a lowercase letter and contain only lowercase letters and numbers."
  exit 1
fi

if ! docker ps --format '{{.Names}}' | grep -qx "${BACKEND_CONTAINER_NAME}"; then
  print_error "Backend container '${BACKEND_CONTAINER_NAME}' is not running. Start dev first with './SocialPredict up'."
  exit 1
fi

print_status "Bootstrapping development users in ${BACKEND_CONTAINER_NAME} ..."

run_dev_bootstrap_env=(
  -e "DEV_BOOTSTRAP_PASSWORD=${password}"
  -e "DEV_BOOTSTRAP_USER_COUNT=${count}"
  -e "DEV_BOOTSTRAP_USER_PREFIX=${prefix}"
)

go_bootstrap_command='export PATH=/usr/local/go/bin:/go/bin:$PATH; cd /src && go run ./cmd/devbootstrap'

if docker exec "${BACKEND_CONTAINER_NAME}" sh -lc 'export PATH=/usr/local/go/bin:/go/bin:$PATH; command -v go >/dev/null 2>&1'; then
  docker exec "${run_dev_bootstrap_env[@]}" "${BACKEND_CONTAINER_NAME}" \
    sh -lc "${go_bootstrap_command}"
  return 0
fi

print_warning "Backend container does not have the Go toolchain; using a temporary Go runner container."

network_name="$(docker inspect -f '{{range $name, $_ := .NetworkSettings.Networks}}{{println $name}}{{end}}' "${BACKEND_CONTAINER_NAME}" | head -n 1)"
if [ -z "${network_name}" ]; then
  print_error "Could not determine Docker network for ${BACKEND_CONTAINER_NAME}."
  exit 1
fi

platform_args=()
if [ -n "${FORCE_PLATFORM:-}" ]; then
  platform_args=(--platform "${FORCE_PLATFORM}")
fi

docker run --rm \
  "${platform_args[@]}" \
  --network "${network_name}" \
  -v "${SCRIPT_DIR}/backend:/src" \
  -w /src \
  -e "APP_ENV=${APP_ENV}" \
  -e "DB_HOST=db" \
  -e "DB_PORT=5432" \
  -e "POSTGRES_USER=${POSTGRES_USER}" \
  -e "POSTGRES_PASSWORD=${POSTGRES_PASSWORD}" \
  -e "POSTGRES_DATABASE=${POSTGRES_DATABASE}" \
  -e "DB_REQUIRE_TLS=${DB_REQUIRE_TLS:-false}" \
  -e "DB_SSLMODE=${DB_SSLMODE:-disable}" \
  "${run_dev_bootstrap_env[@]}" \
  golang:1.25.1-alpine3.22 \
  sh -lc "${go_bootstrap_command}"
