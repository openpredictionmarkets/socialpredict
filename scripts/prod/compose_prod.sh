#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

if [ "$1" = "up" ]; then
	# Make sure docker network exists
	docker network inspect socialpredict_external_network > /dev/null 2>&1 || docker network create --driver bridge socialpredict_external_network

	# Make sure acme.json file exists
	if [ ! -f $SCRIPT_DIR/data/traefik/config/acme.json ]; then
		touch $SCRIPT_DIR/data/traefik/config/acme.json
		chmod 600 $SCRIPT_DIR/data/traefik/config/acme.json
	fi

	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" up -d
elif [ "$1" = "down" ]; then
	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" down
else
	echo "Wrong Command."
	exit 1;
fi
