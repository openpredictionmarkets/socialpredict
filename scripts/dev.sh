#!/bin/bash

# --- Platform Compatibility, Linux vs. Apple Silicon ---
__SP_DEV_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
__SP_ROOT_DIR="${SOCIALPREDICT_ROOT:-$(cd "${__SP_DEV_DIR}/.." && pwd)}"

source "${__SP_ROOT_DIR}/scripts/lib/arch.sh"
echo "== dev build platform: ${FORCE_PLATFORM:-default} =="

# Cross-platform "sed -i" (GNU vs BSD/macOS)
if sed --version >/dev/null 2>&1; then
  # GNU sed
  SED_INPLACE=(sed -i -e)
else
  # BSD sed (macOS) requires an empty backup extension: -i ''
  SED_INPLACE=(sed -i '' -e)
fi

# (existing) set local image names for dev
export BACKEND_IMAGE_NAME=${BACKEND_IMAGE_NAME:-socialpredict-dev-backend}
export FRONTEND_IMAGE_NAME=${FRONTEND_IMAGE_NAME:-socialpredict-dev-frontend}

# --- Main SocialPredict Functionality ---

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

init_env() {
  # Create .env file
  cp "${__SP_ROOT_DIR}/.env.example" "${__SP_ROOT_DIR}/.env"

  # Update APP_ENV
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV='development'|" "${__SP_ROOT_DIR}/.env"

  # Add OS-specific POSTGRES_VOLUME
  OS=$(uname -s)
  if [[ "$OS" == "Darwin" ]]; then
    echo "POSTGRES_VOLUME=pgdata" >> "${__SP_ROOT_DIR}/.env"
  else
    echo "POSTGRES_VOLUME=../data/postgres" >> "${__SP_ROOT_DIR}/.env"
  fi

  # Change the Domain setting:
  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='localhost'|" "${__SP_ROOT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='http://localhost'|" "${__SP_ROOT_DIR}/.env"

  # Remove unnecessary lines from .env
  "${SED_INPLACE[@]}" "/^TRAEFIK_CONTAINER_NAME=.*/d" "${__SP_ROOT_DIR}/.env"
  "${SED_INPLACE[@]}" "/^EMAIL=.*/d" "${__SP_ROOT_DIR}/.env"

  # Update Image Names
  "${SED_INPLACE[@]}" "s|^BACKEND_IMAGE_NAME=.*|BACKEND_IMAGE_NAME=${BACKEND_IMAGE_NAME}|" "${__SP_ROOT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^FRONTEND_IMAGE_NAME=.*|FRONTEND_IMAGE_NAME=${FRONTEND_IMAGE_NAME}|" "${__SP_ROOT_DIR}/.env"
}

if [[ ! -f "${__SP_ROOT_DIR}/.env" ]]; then
  echo "### First time running the script ..."
  echo "Let's initialize the application ..."
  sleep 1
  init_env
  echo "Application initialized successfully."
else
  read -r -p ".env file found. Do you want to re-create it? (y/N) " DECISION
  if [[ "$DECISION" = "Y" || "$DECISION" = "y" ]]; then
    sleep 1
    echo "Re-creating env file ..."
    sleep 1
    init_env
    echo ".env file re-created successfully."
  fi
fi

echo

sleep 1;

check_image() {
  local image_name=$1
  local dockerfile=$2
  local directory=$3

  echo "### Checking for $image_name Image ..."
  if docker image inspect "$image_name" > /dev/null 2>&1; then
    read -r -p "$image_name Image Found. Do you want to re-build it? (y/N) " decision
    if [[ "$decision" =~ ^[Yy]$ ]]; then
      echo "Deleting Image ..."
      docker rmi "$image_name"
      echo "Image Deleted."
      build_image
    else
      :
    fi
  else
    echo "$image_name Image Not Found."
    build_image
  fi
}

build_image() {
  echo "Building $image_name now."
  docker build --no-cache -t "$image_name" -f "$dockerfile" "$directory"
  echo "$image_name Image Built."
}

echo "### Searching for Docker Images ..."
sleep 1;

BACKEND_IMAGE_NAME="${BACKEND_IMAGE_NAME:-socialpredict-dev-backend}"
FRONTEND_IMAGE_NAME="${FRONTEND_IMAGE_NAME:-socialpredict-dev-frontend}"

DIRECTORY="${__SP_ROOT_DIR}"
BACKEND_DOCKERFILE="${__SP_ROOT_DIR}/docker/backend/Dockerfile.dev"
FRONTEND_DOCKERFILE="${__SP_ROOT_DIR}/docker/frontend/Dockerfile.dev"

check_image "$BACKEND_IMAGE_NAME" "${BACKEND_DOCKERFILE}" "${DIRECTORY}"
check_image "$FRONTEND_IMAGE_NAME" "${FRONTEND_DOCKERFILE}" "${DIRECTORY}"

echo
sleep 1;

echo "Images built."
echo "Use "./SocialPredict up" to start the containers"
echo "And "./SocialPredict down" to stop them."
