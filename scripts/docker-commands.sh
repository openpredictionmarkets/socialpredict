#!/usr/bin/env bash

# --------------------------------------------------------------------------------------
# Platform handling (Apple Silicon, etc.) – absolute path to avoid CWD issues
# --------------------------------------------------------------------------------------
if [ -f "${SCRIPT_DIR}/scripts/lib/arch.sh" ]; then
  source "${SCRIPT_DIR}/scripts/lib/arch.sh"
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

COMPOSE_MAIN="${SCRIPT_DIR}/scripts/${COMPOSE_BASENAME}"
if [ ! -f "${COMPOSE_MAIN}" ]; then
  echo "ERROR: compose file not found for APP_ENV='${APP_ENV}': ${COMPOSE_MAIN}"
  exit 1
fi
echo "Using compose file: ${COMPOSE_MAIN}"

# optional override (e.g., Apple Silicon platform pinning)
COMPOSE_FILES=(-f "${COMPOSE_MAIN}")
if [ -f "${SCRIPT_DIR}/scripts/docker-compose.override.yml" ]; then
  COMPOSE_FILES+=(-f "${SCRIPT_DIR}/scripts/docker-compose.override.yml")
fi

# --------------------------------------------------------------------------------------
# Helpers
# --------------------------------------------------------------------------------------
sp_up() {
  if [ "${APP_ENV}" = "production" ]; then
    # Ensure external network exists
    docker network inspect socialpredict_external_network > /dev/null 2>&1 || \
      docker network create --driver bridge socialpredict_external_network

    # Ensure acme.json exists with correct perms
    ACME_FILE="${SCRIPT_DIR}/data/traefik/config/acme.json"
    if [ ! -f "${ACME_FILE}" ]; then
      mkdir -p "$(dirname "${ACME_FILE}")"
      touch "${ACME_FILE}"
      chmod 600 "${ACME_FILE}"
    fi
  fi

  docker compose --env-file "${SCRIPT_DIR}/.env" "${COMPOSE_FILES[@]}" up -d

  if [ "${APP_ENV}" = "development" ]; then
    echo "SocialPredict may be found at http://localhost:${FRONTEND_PORT} ."
    echo "This may take a few seconds to load initially."
    echo "Here are the initial settings. These can be changed in setup.yaml"
    if [ -f "${SCRIPT_DIR}/backend/setup/setup.yaml" ]; then
      cat "${SCRIPT_DIR}/backend/setup/setup.yaml"
    fi
  fi
}

sp_down() {
  docker compose --env-file "${SCRIPT_DIR}/.env" "${COMPOSE_FILES[@]}" down -v
}

sp_exec() {
  local target="${1:-}"

  case "${target}" in
    nginx)
      if [ "$#" -lt 2 ]; then
        docker exec -it "${NGINX_CONTAINER_NAME}" /bin/bash
      else
        if [ "$2" == "/bin/bash" ] || [ "$2" == /bin/sh ]; then
          docker exec -it "${NGINX_CONTAINER_NAME}" /bin/bash
        else
          shift
          docker exec -i "${NGINX_CONTAINER_NAME}" "$@"
        fi
      fi
      ;;
    backend)
      if [ "$#" -lt 2 ]; then
        docker exec -it "${BACKEND_CONTAINER_NAME}" /bin/bash
      else
        if [ "$2" == "/bin/bash" ] || [ "$2" == "/bin/sh" ]; then
          docker exec -it "${BACKEND_CONTAINER_NAME}" /bin/bash
        else
          shift
          docker exec -i "${BACKEND_CONTAINER_NAME}" "$@"
        fi
      fi
      ;;
    frontend)
      if [ "$#" -lt 2 ]; then
        docker exec -it "${FRONTEND_CONTAINER_NAME}" /bin/bash
      else
        if [ "$2" == "/bin/bash" ] || [ "$2" == "/bin/sh" ]; then
          docker exec -it "${FRONTEND_CONTAINER_NAME}" /bin/bash
        else
          shift
          docker exec -i "${FRONTEND_CONTAINER_NAME}" "$@"
        fi
      fi
      ;;
    postgres|db)
      if [ "$#" -lt 2 ]; then
        docker exec -it "${POSTGRES_CONTAINER_NAME}" /bin/bash
      else
        if [ "$2" == "/bin/bash" ] || [ "$2" == "/bin/sh" ]; then
          docker exec -it "${POSTGRES_CONTAINER_NAME}" /bin/bash
        else
          docker exec -i "${POSTGRES_CONTAINER_NAME}" "$@"
        fi
      fi
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
    shift
    sp_exec "$@"
    ;;
  *)
    echo "Usage: $0 {up|down|exec <service> [command]}"
    exit 1
    ;;
esac
