#!/usr/bin/env bash

set -euo pipefail

[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

print_dev_bootstrap_help() {
  cat <<'EOF'
Usage: ./SocialPredict dev-bootstrap-users [OPTIONS]

Create or reset development-only login fixtures.

This command refuses to run unless APP_ENV=development in .env.
It runs inside the backend container and writes to the development database.

Created/updated users:
  admin       ADMIN user
  testuser01  REGULAR user
  testuser02  REGULAR user
  ...

Defaults:
  password: Password1
  user count: 10
  username prefix: testuser
  must_change_password: true

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

password="Password1"
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
docker exec \
  -e DEV_BOOTSTRAP_PASSWORD="${password}" \
  -e DEV_BOOTSTRAP_USER_COUNT="${count}" \
  -e DEV_BOOTSTRAP_USER_PREFIX="${prefix}" \
  "${BACKEND_CONTAINER_NAME}" \
  sh -lc 'cd /src && go run ./cmd/devbootstrap'
