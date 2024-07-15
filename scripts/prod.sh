#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Check if script runs for the first time
if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
	export CALLED_FROM_SOCIALPREDICT=yes
	source ./scripts/env_writer_prod.sh
	unset CALLED_FROM_SOCIALPREDICT
else
	read -p ".env file found. Do you want to re-create it? (y/N) " DECISION
	if [ "$DECISION" != "Y" ] && [ "$DECISION" != "y" ]; then
		:
	else
		export CALLED_FROM_SOCIALPREDICT=yes
		source ./scripts/env_writer_prod.sh
		unset CALLED_FROM_SOCIALPREDICT
	fi
fi

source_env

echo
sleep 1;

# Build Docker Images
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

