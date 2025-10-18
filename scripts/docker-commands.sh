#!/bin/bash

# --- Platform Compatibility, Linux vs. Apple Silicon ---

source "$(dirname "$0")/lib/arch.sh"

COMPOSE_FILES=(-f scripts/docker-compose-${APP_ENV}.yaml)
if [ -f "docker-compose.override.yml" ]; then
  COMPOSE_FILES+=(-f docker-compose.override.yml)
fi

# --- Main SocialPredict Functionality ---

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

if [ "$1" = "up" ]; then
  if [ "${APP_ENV}" = "development" ]; then
    docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-dev.yaml" up -d && \
    	echo "SocialPredict may be found at http://localhost:${FRONTEND_PORT} . This may take a few seconds to load initially."
    echo "Here are the initial settings. These can be changed in setup.yaml"
    cat "$SCRIPT_DIR"/backend/setup/setup.yaml
  elif [ "${APP_ENV}" = "localhost" ]; then
	  docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-local.yaml" up -d
  elif [ "${APP_ENV}" = "production" ]; then
    # Make sure docker network exists
    docker network inspect socialpredict_external_network > /dev/null 2>&1 || docker network create --driver bridge socialpredict_external_network

    # Make sure acme.json file exists
    if [ ! -f "$SCRIPT_DIR"/data/traefik/config/acme.json ]; then
    	touch "$SCRIPT_DIR"/data/traefik/config/acme.json
    	chmod 600 "$SCRIPT_DIR"/data/traefik/config/acme.json
    fi

    docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" up -d
  else
    echo "Wrong Application Environment in .env"
    exit 1
  fi
elif [ "$1" = "down" ]; then
  if [ "${APP_ENV}" = "development" ]; then
	  docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-dev.yaml" down -v
  elif [ "${APP_ENV}" = "localhost" ]; then
  	docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-local.yaml" down -v
  elif [ "${APP_ENV}" = "production" ]; then
	  docker compose --env-file "$SCRIPT_DIR"/.env --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" down -v
	else
    echo "Wrong Application Environment in .env"
    exit 1
  fi
elif [ "$1" = "exec" ]; then
  if [ "$2" = "nginx" ]; then
  	if [ -z "$3" ]; then
  		docker exec -it "${NGINX_CONTAINER_NAME}" /bin/bash
  	else
  		docker exec "${NGINX_CONTAINER_NAME}" "$3"
  	fi
  elif [ "$2" = "backend" ]; then
    if [ -z "$3" ]; then
      docker exec -it "$BACKEND_CONTAINER_NAME}" /bin/bash
    else
      docker exec "${BACKEND_CONTAINER_NAME}" "$3"
    fi
  elif [ "$2" = "frontend" ]; then
    if [ -z "$3" ]; then
      docker exec -it "${FRONTEND_CONTAINER_NAME}" /bin/bash
    else
      docker exec "${FRONTEND_CONTAINER_NAME}" "$3"
    fi
  elif [ "$2" = "postgres" ]; then
    if [ -z "$3" ]; then
      docker exec -it "${POSTGRES_CONTAINER_NAME}" /bin/bash
    else
      docker exec "${POSTGRES_CONTAINER_NAME}" "$3"
    fi
  else
  	echo "Wrong Container Name."
  fi
fi
