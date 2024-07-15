#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

if [ "$1" = "up" ]; then
	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-dev.yaml" up -d && \
	echo "SocialPredict may be found at http://localhost . This may take a few seconds to load initially."
elif [ "$1" = "down" ]; then
	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-dev.yaml" down
else
	echo "Wrong Command."
	exit 1;
fi
