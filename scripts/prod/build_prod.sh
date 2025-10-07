#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Pull images
echo "Pulling images ..."
$COMPOSE $ENV_FILE --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" pull
echo

echo "Images pulled."
echo
echo "Your admin credentials are:"
echo "Username: admin"
echo "Password: $ADMIN_PASS"
