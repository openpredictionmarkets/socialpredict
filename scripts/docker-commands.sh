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
COMMAND="${1:-}"

if [ "${COMMAND}" != "cleanup" ]; then
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

print_cleanup_help() {
  cat <<'EOF'
Usage: ./SocialPredict cleanup docker [OPTIONS]

Cleanup unused Docker artifacts for the current host.

Default behavior:
  - remove stopped containers
  - remove dangling images
  - remove Docker build cache
  - do not remove volumes

Options:
  --all-images   Remove all unused images, not just dangling images
  --volumes      Also prune unused Docker volumes; never use casually on data hosts
  --help         Show this help

Deploy guidance:
  Use the default command in automation:
    ./SocialPredict cleanup docker

  Avoid --volumes in deployment workflows unless data retention is explicitly understood.
EOF
}

sp_cleanup_docker() {
  local prune_all_images="n"
  local prune_volumes="n"

  while [ "$#" -gt 0 ]; do
    case "$1" in
      --all-images)
        prune_all_images="y"
        ;;
      --volumes)
        prune_volumes="y"
        ;;
      --help|-h)
        print_cleanup_help
        return 0
        ;;
      *)
        echo "Unknown cleanup option: $1"
        print_cleanup_help
        exit 1
        ;;
    esac
    shift
  done

  echo "Docker disk usage before cleanup:"
  docker system df || true
  echo

  echo "Removing stopped containers ..."
  docker container prune -f

  if [ "$prune_all_images" = "y" ]; then
    echo "Removing all unused images ..."
    docker image prune -a -f
  else
    echo "Removing dangling images ..."
    docker image prune -f
  fi

  echo "Removing Docker build cache ..."
  docker builder prune -f

  if [ "$prune_volumes" = "y" ]; then
    print_warning "Pruning unused Docker volumes. This can delete data if a volume is detached."
    docker volume prune -f
  else
    echo "Skipping Docker volume prune."
  fi

  echo
  echo "Docker disk usage after cleanup:"
  docker system df || true
}

sp_cleanup() {
  local target="${1:-docker}"
  if [ "$#" -gt 0 ]; then
    shift
  fi

  case "${target}" in
    docker)
      sp_cleanup_docker "$@"
      ;;
    --help|-h|help)
      print_cleanup_help
      ;;
    *)
      echo "Unknown cleanup target '${target}'. Use: docker"
      exit 1
      ;;
  esac
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
# Entry points (up / down / exec / cleanup)
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
  cleanup)
    shift
    sp_cleanup "$@"
    ;;
  *)
    echo "Usage: $0 {up|down|exec <service> [command]|cleanup docker [options]}"
    exit 1
    ;;
esac
