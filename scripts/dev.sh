#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Function to replace API_URI in frontend/src/config.js
frontend_api_uri() {
	template="$SCRIPT_DIR/frontend/src/config.js.template"
	file="$SCRIPT_DIR/frontend/src/config.js"
	export DOMAIN="'http://localhost'"
	envsubst < $template > $file
}

# Check for .env file
export CALLED_FROM_SOCIALPREDICT=yes
source "$SCRIPT_DIR/scripts/dev/env_writer_dev.sh"
unset CALLED_FROM_SOCIALPREDICT

source_env

frontend_api_uri

sleep 1;

# Build Docker Images
export CALLED_FROM_SOCIALPREDICT=yes
source "$SCRIPT_DIR/scripts/dev/build_dev.sh"
unset CALLED_FROM_SOCIALPREDICT
