# scripts/backup/lib/env.sh
#
# Guards and environment validation

# --- Guards -------------------------------------------------------------------
[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

# .env is already sourced by the SocialPredict script before calling this file
: "${APP_ENV:?APP_ENV not set}"
: "${POSTGRES_CONTAINER_NAME:?POSTGRES_CONTAINER_NAME not set}"
: "${POSTGRES_USER:?POSTGRES_USER not set}"
: "${POSTGRES_PASSWORD:?POSTGRES_PASSWORD not set}"
: "${POSTGRES_DATABASE:?POSTGRES_DATABASE not set}"
: "${POSTGRES_PORT:=5432}"

# Container helpers
container_running() {
  docker ps --format '{{.Names}}' | grep -qx "$POSTGRES_CONTAINER_NAME"
}

need_container_running() {
  if ! container_running; then
    echo "ERROR: Postgres container '$POSTGRES_CONTAINER_NAME' is not running."
    echo "Start it first: ./SocialPredict up"
    exit 1
  fi
}
