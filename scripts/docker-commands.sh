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

compose_cmd() {
  docker compose "${COMPOSE_FILES[@]}" ${ENV_FILE} "$@"
}

compose_services() {
  compose_cmd config --services
}

compose_container_name() {
  local target_service="${1:-}"

  compose_cmd config | awk -v target="${target_service}" '
    /^services:/ { in_services=1; next }
    in_services && /^[^[:space:]]/ { in_services=0 }
    in_services {
      if ($0 ~ /^  [^[:space:]][^:]*:$/) {
        current=$0
        sub(/^  /, "", current)
        sub(/:$/, "", current)
        next
      }

      if (current == target && $0 ~ /^    container_name:/) {
        sub(/^    container_name:[[:space:]]*/, "", $0)
        gsub(/^"/, "", $0)
        gsub(/"$/, "", $0)
        print $0
        exit
      }
    }
  '
}

print_logs_options() {
  local service
  local container_name

  echo "Available log targets:"
  while IFS= read -r service; do
    [ -z "${service}" ] && continue
    container_name="$(compose_container_name "${service}")"
    if [ -n "${container_name}" ]; then
      printf "  %-12s %s\n" "${service}" "${container_name}"
    else
      printf "  %-12s %s\n" "${service}" "(container name managed by docker compose)"
    fi
  done < <(compose_services)
  printf "  %-12s %s\n" "all" "docker compose aggregated logs"
  echo
  echo "Use './SocialPredict logs help' or './SocialPredict logs --help' for examples."
}

resolve_log_service() {
  local target="${1:-}"
  local service
  local container_name

  while IFS= read -r service; do
    [ -z "${service}" ] && continue

    if [ "${target}" = "${service}" ]; then
      echo "${service}"
      return 0
    fi

    container_name="$(compose_container_name "${service}")"
    if [ -n "${container_name}" ] && [ "${target}" = "${container_name}" ]; then
      echo "${service}"
      return 0
    fi
  done < <(compose_services)

  return 1
}

sp_logs() {
  local target="${1:-}"
  local resolved_service
  shift || true

  if [ -z "${target}" ]; then
    print_logs_options
    return 0
  fi

  case "${target}" in
    options|list)
      print_logs_options
      return 0
      ;;
    -h|--help|help)
      cat <<EOF
Usage: ./SocialPredict logs <service> [-f|--follow]
       ./SocialPredict logs all [-f|--follow]

Examples:
  ./SocialPredict logs
  ./SocialPredict logs options
  ./SocialPredict logs help
  ./SocialPredict logs <service>
  ./SocialPredict logs <service> -f
  ./SocialPredict logs all

Services and current container names from the active compose file:
EOF
      print_logs_options
      return 0
      ;;
  esac

  local follow_args=()
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -f|--follow)
        follow_args+=("-f")
        ;;
      *)
        echo "Unknown logs option '$1'. Use -f or --follow."
        exit 1
        ;;
    esac
    shift
  done

  if [ "${target}" = "all" ]; then
    if [ "${#follow_args[@]}" -gt 0 ]; then
      compose_cmd logs "${follow_args[@]}"
    else
      compose_cmd logs
    fi
    return 0
  fi

  resolved_service="$(resolve_log_service "${target}")" || {
    echo "Unknown log target '${target}'."
    print_logs_options
    exit 1
  }

  if [ "${#follow_args[@]}" -gt 0 ]; then
    compose_cmd logs "${follow_args[@]}" "${resolved_service}"
  else
    compose_cmd logs "${resolved_service}"
  fi
}

# --------------------------------------------------------------------------------------
# Entry points (up / down / exec / logs)
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
  logs)
    # Usage: docker-commands.sh logs <service|container-name|all> [-f|--follow]
    svc="${2:-}"
    shift 2 || true
    sp_logs "${svc}" "$@"
    ;;
  *)
    echo "Usage: $0 {up|down|exec <service> [command]|logs <service|container-name|all> [-f|--follow]}"
    exit 1
    ;;
esac
