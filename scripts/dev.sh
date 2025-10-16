#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

init_env() {
	# Create .env file
	cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV='development'/g" "$SCRIPT_DIR/.env"

	# Add OS-specific POSTGRES_VOLUME
	OS=$(uname -s)
	if [[ "$OS" == "Darwin" ]]; then
		echo "POSTGRES_VOLUME=pgdata" >> "$SCRIPT_DIR/.env"
	else
		echo "POSTGRES_VOLUME=../data/postgres" >> "$SCRIPT_DIR/.env"
	fi

	# Update domain name

  	# Change the Domain setting:
  	sed -i -e "s/DOMAIN=.*/DOMAIN='localhost'/g" .env
  	sed -i -e 's/DOMAIN_URL=.*/DOMAIN_URL='\''http:\/\/'"localhost"''\''/g' .env
  	echo

  	# Remove unnecessary lines from .env
  	sed -i '/TRAEFIK_CONTAINER_NAME=.*/d' .env
  	sed -i '/EMAIL=.*/d' .env

  # Update Image Names
  sed -i -e "s/BACKEND_IMAGE_NAME=.*/BACKEND_IMAGE_NAME=socialpredict-dev-backend/g" .env
  sed -i -e "s/FRONTEND_IMAGE_NAME=.*/FRONTEND_IMAGE_NAME=socialpredict-dev-frontend/g" .env
}

if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
	echo "### First time running the script ..."
	echo "Let's initialize the application ..."
	sleep 1
	init_env
	echo "Application initialized successfully."
else
	read -r -p ".env file found. Do you want to re-create it? (y/N) " DECISION
	if [ "$DECISION" != "Y" ] && [ "$DECISION" != "y" ]; then
		:
	else
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

DIRECTORY="${SCRIPT_DIR}/."

BACKEND_DOCKERFILE="${SCRIPT_DIR}/docker/backend/Dockerfile.dev"
FRONTEND_DOCKERFILE="${SCRIPT_DIR}/docker/frontend/Dockerfile.dev"

check_image "$BACKEND_IMAGE_NAME" "${BACKEND_DOCKERFILE}" "${DIRECTORY}"
check_image "$FRONTEND_IMAGE_NAME" "${FRONTEND_DOCKERFILE}" "${DIRECTORY}"

echo
sleep 1;

echo "Images built."
echo "Use "./SocialPredict up" to start the containers"
echo "And "./SocialPredict down" to stop them."
