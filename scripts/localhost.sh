#!/usr/bin/env bash
set -euo pipefail

# Resolve paths (don’t rely on caller CWD)
__SP_LOCAL_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__SP_ROOT_DIR="${SOCIALPREDICT_ROOT:-$(cd "${__SP_LOCAL_DIR}/.." && pwd)}"

# Apple Silicon handling (sets FORCE_PLATFORM / DOCKER_DEFAULT_PLATFORM as needed)
# shellcheck source=/dev/null
source "${__SP_ROOT_DIR}/scripts/lib/arch.sh"
echo "== localhost platform: ${FORCE_PLATFORM:-default} =="

# Cross-platform sed -i
if sed --version >/dev/null 2>&1; then
  SED_INPLACE=(sed -i -e)     # GNU
else
  SED_INPLACE=(sed -i '' -e)  # BSD/macOS
fi

# Guard: only run via ./SocialPredict
[ -z "${CALLED_FROM_SOCIALPREDICT:-}" ] && { echo "Not called from SocialPredict"; exit 42; }

init_env() {
  cp "${__SP_ROOT_DIR}/.env.example" "${__SP_ROOT_DIR}/.env"

  # Mode
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV=localhost|" "${__SP_ROOT_DIR}/.env"

  # GHCR images for localhost pulls
  if grep -q '^BACKEND_IMAGE_NAME=' "${__SP_ROOT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^BACKEND_IMAGE_NAME=.*|BACKEND_IMAGE_NAME=ghcr.io/openpredictionmarkets/socialpredict-backend:latest|" "${__SP_ROOT_DIR}/.env"
  else
    printf "\nBACKEND_IMAGE_NAME=ghcr.io/openpredictionmarkets/socialpredict-backend:latest\n" >> "${__SP_ROOT_DIR}/.env"
  fi
  if grep -q '^FRONTEND_IMAGE_NAME=' "${__SP_ROOT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^FRONTEND_IMAGE_NAME=.*|FRONTEND_IMAGE_NAME=ghcr.io/openpredictionmarkets/socialpredict-frontend:latest|" "${__SP_ROOT_DIR}/.env"
  else
    printf "\nFRONTEND_IMAGE_NAME=ghcr.io/openpredictionmarkets/socialpredict-frontend:latest\n" >> "${__SP_ROOT_DIR}/.env"
  fi

  # Postgres (multi-arch tag + named volume)
  if grep -q '^POSTGRES_IMAGE=' "${__SP_ROOT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^POSTGRES_IMAGE=.*|POSTGRES_IMAGE=postgres:16.6-alpine|" "${__SP_ROOT_DIR}/.env"
  else
    printf "\nPOSTGRES_IMAGE=postgres:16.6-alpine\n" >> "${__SP_ROOT_DIR}/.env"
  fi
  if grep -q '^POSTGRES_VOLUME=' "${__SP_ROOT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^POSTGRES_VOLUME=.*|POSTGRES_VOLUME=pgdata|" "${__SP_ROOT_DIR}/.env"
  else
    printf "\nPOSTGRES_VOLUME=pgdata\n" >> "${__SP_ROOT_DIR}/.env"
  fi

  # Localhost URLs
  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN=localhost|" "${__SP_ROOT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL=http://localhost|" "${__SP_ROOT_DIR}/.env"
  if grep -q '^API_URL=' "${__SP_ROOT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^API_URL=.*|API_URL=http://localhost|" "${__SP_ROOT_DIR}/.env"
  else
    printf "\nAPI_URL=http://localhost\n" >> "${__SP_ROOT_DIR}/.env"
  fi

  # Clean prod-only lines
  "${SED_INPLACE[@]}" "/^TRAEFIK_CONTAINER_NAME=.*/d" "${__SP_ROOT_DIR}/.env"
  "${SED_INPLACE[@]}" "/^EMAIL=.*/d" "${__SP_ROOT_DIR}/.env"

  echo "localhost .env prepared for GHCR images."
}

# Pin platform per-service (helpful on Apple Silicon)
cat > "${__SP_ROOT_DIR}/docker-compose.override.yml" <<EOF
services:
  backend:
    platform: ${FORCE_PLATFORM:-linux/amd64}
  frontend:
    platform: ${FORCE_PLATFORM:-linux/amd64}
  db:
    platform: ${FORCE_PLATFORM:-linux/amd64}
EOF
echo "Wrote docker-compose.override.yml to pin platform = ${FORCE_PLATFORM:-linux/amd64}"

# Initialize or refresh .env
if [[ ! -f "${__SP_ROOT_DIR}/.env" ]]; then
  echo "### First time running localhost setup ..."
  init_env
  echo "Application initialized successfully."
else
  read -r -p ".env file found. Re-create for localhost? (y/N) " DECISION
  if [[ "${DECISION}" =~ ^[Yy]$ ]]; then
    init_env
    echo ".env file re-created successfully."
  else
    echo "Keeping existing .env"
  fi
fi

# (Optional) Early pull to surface auth/tag issues; docker-commands.sh will handle 'up'
echo
echo "Pulling images ..."
docker compose \
  -f "${__SP_ROOT_DIR}/scripts/docker-compose-local.yaml" \
  -f "${__SP_ROOT_DIR}/docker-compose.override.yml" \
  --env-file "${__SP_ROOT_DIR}/.env" pull
echo "Images pulled."
