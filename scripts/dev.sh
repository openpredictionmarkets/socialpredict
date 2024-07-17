#!/bin/bash

echo "Checking if called from SocialPredict..."
[ -z "${CALLED_FROM_SOCIALPREDICT}" ] && { echo "Not called from SocialPredict"; exit 42; }

echo "Defining frontend_api_uri function..."
frontend_api_uri() {
	template="${SCRIPT_DIR}/frontend/src/config.js.template"
	file="${SCRIPT_DIR}/frontend/src/config.js"
	export API_DOMAIN="'http://localhost'"
	envsubst < $template > $file
}

echo "Checking for .env file..."
export CALLED_FROM_SOCIALPREDICT=yes
echo "Setting CALLED_FROM_SOCIALPREDICT: ${CALLED_FROM_SOCIALPREDICT}"
source "${SCRIPT_DIR}/scripts/dev/env_writer_dev.sh"
unset CALLED_FROM_SOCIALPREDICT
echo "Unsetting... CALLED_FROM_SOCIALPREDICT: ${CALLED_FROM_SOCIALPREDICT}"

echo "Calling source_env..."
source_env

echo "Calling frontend_api_uri..."
frontend_api_uri

echo "Sleeping..."
sleep 1;

echo "Building Docker Images..."
export CALLED_FROM_SOCIALPREDICT=yes
echo "Setting CALLED_FROM_SOCIALPREDICT: ${CALLED_FROM_SOCIALPREDICT}"
source "${SCRIPT_DIR}/scripts/dev/build_dev.sh"
unset CALLED_FROM_SOCIALPREDICT
echo "Unsetting... CALLED_FROM_SOCIALPREDICT: ${CALLED_FROM_SOCIALPREDICT}"

echo "Script completed."