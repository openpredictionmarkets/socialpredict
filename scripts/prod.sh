#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Check if script runs for the first time
if [[ ! -f "$SCRIPT_DIR/.first_run" ]]; then
	export CALLED_FROM_SOCIALPREDICT=yes
	source ./scripts/env_writer.sh
	unset CALLED_FROM_SOCIALPREDICT
fi

source_env

echo
sleep 1;

# Check if backend image exists
echo "Searching for Backend Image ..."
sleep 1;

export CALLED_FROM_SOCIALPREDICT=yes
source ./scripts/build.sh
unset CALLED_FROM_SOCIALPREDICT

# Issue SSL Certificate for ${DOMAIN}
echo "Using Domain ${DOMAIN} for the deployment."
sleep 1;
export CALLED_FROM_SOCIALPREDICT=yes
source ./scripts/ssl.sh
unset CALLED_FROM_SOCIALPREDICT

# Build rest of the images
# $COMPOSE $ENV_FILE --file "./scripts/docker-compose-prod.yaml" pull

