#!/usr/bin/env bash
set -euo pipefail

# --------------------------------------------------------------------------------------
# Path resolution (stable even when this script is "sourced" by ./SocialPredict)
# --------------------------------------------------------------------------------------
__SP_CMD_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__SP_ROOT_DIR="${SOCIALPREDICT_ROOT:-$(cd "${__SP_CMD_DIR}/.." && pwd)}"

# --------------------------------------------------------------------------------------
# Platform handling (Apple Silicon, etc.) – absolute path to avoid CWD issues
# --------------------------------------------------------------------------------------
if [ -f "${__SP_ROOT_DIR}/scripts/lib/arch.sh" ]; then
  # shellcheck source=/dev/null
  source "${__SP_ROOT_DIR}/scripts/lib/arch.sh"
fi

# --------------------------------------------------------------------------------------
# Guard: must be run via ./SocialPredict
# --------------------------------------------------------------------------------------
[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

# --------------------------------------------------------------------------------------
# Compose files and env
# --------------------------------------------------------------------------------------
APP_ENV="${APP_ENV:-development}"  # fallback if not set by .env yet

# Map APP_ENV to actual compose filename in scripts/
case "${APP_ENV}" in
  development) COMPOSE_BASENAME="docker-compose-dev.yaml" ;;
  localhost)   COMPOSE_BASENAME="docker-compose-local.yaml" ;;
  production)  COMPOSE_BASENAME="docker-compose-prod.yaml" ;;
  *) echo "ERROR: unknown APP_ENV='${APP_ENV}'. Expected one of: development | localhost | production"; exit 1 ;;
esac

COMPOSE_MAIN="${__SP_ROOT_DIR}/scripts/${COMPOSE_BASENAME}"
if [ ! -f "${COMPOSE_MAIN}" ]; then
  echo "ERROR: compose file not found for APP_ENV='${APP_ENV}': ${COMPOSE_MAIN}"
  exit 1
fi
echo "Using compose file: ${COMPOSE_MAIN}"

# optional override (e.g., Apple Silicon platform pinning)
COMPOSE_FILES=(-f "${COMPOSE_MAIN}")
if [ -f "${__SP_ROOT_DIR}/docker-compose.override.yml" ]; then
  COMPOSE_FILES+=(-f "${__SP_ROOT_DIR}/docker-compose.override.yml")
fi

# absolute .env
ENV_FILE="--env-file ${__SP_ROOT_DIR}/.env"

# --------------------------------------------------------------------------------------
# Helpers
# --------------------------------------------------------------------------------------
sp_up() {
  if [ "${APP_ENV}" = "production" ]; then
    # Ensure external network exists
    docker network inspect socialpredict_external_network > /dev/null 2>&1 || \
      docker network create --driver bridge socialpredict_external_network

    # Ensure acme.json exists with correct perms
    ACME_FILE="${__SP_ROOT_DIR}/data/traefik/config/acme.json"
    if [ ! -f "${ACME_FILE}" ]; then
      mkdir -p "$(dirname "${ACME_FILE}")"
      touch "${ACME_FILE}"
      chmod 600 "${ACME_FILE}"
    fi
  fi

  docker compose "${COMPOSE_FILES[@]}" ${ENV_FILE} up -d

  if [ "${APP_ENV}" = "development" ]; then
    echo "SocialPredict may be found at http://localhost:${FRONTEND_PORT} ."
    echo "This may take a few seconds to load initially."
    echo "Here are the initial settings. These can be changed in setup.yaml"
    if [ -f "${__SP_ROOT_DIR}/backend/setup/setup.yaml" ]; then
      cat "${__SP_ROOT_DIR}/backend/setup/setup.yaml"
    fi
  fi
}

sp_down() {
  docker compose "${COMPOSE_FILES[@]}" ${ENV_FILE} down -v
}

sp_exec() {
  local target="${1:-}"
  local cmd="${2:-/bin/bash}"

  case "${target}" in
    nginx)
      docker exec -it "${NGINX_CONTAINER_NAME}" ${cmd}
      ;;
    backend)
      docker exec -it "${BACKEND_CONTAINER_NAME}" ${cmd}
      ;;
    frontend)
      docker exec -it "${FRONTEND_CONTAINER_NAME}" ${cmd}
      ;;
    postgres|db)
      docker exec -it "${POSTGRES_CONTAINER_NAME}" ${cmd}
      ;;
    *)
      echo "Unknown service '${target}'. Use one of: nginx | backend | frontend | postgres"
      exit 1
      ;;
  esac
}

# --------------------------------------------------------------------------------------
# Entry points (up / down / exec)
# --------------------------------------------------------------------------------------
case "${1:-}" in
  up)
    sp_up
    ;;
  down)
    sp_down
    ;;
  exec)
    # Usage: docker-commands.sh exec <service> [command]
    # Example: docker-commands.sh exec backend "bash -lc 'ls -la'"
    svc="${2:-}"
    shift 2 || true
    if [ -z "${svc}" ]; then
      echo "Usage: $0 exec <nginx|backend|frontend|postgres> [command]"
      exit 1
    fi
    sp_exec "${svc}" "$*"
    ;;
  *)
    echo "Usage: $0 {up|down|exec <service> [command]}"
    exit 1
    ;;
esac
