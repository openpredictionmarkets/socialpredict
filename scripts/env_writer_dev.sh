#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

echo "### First time running the script ..."
echo "### Let's initialize the appliction ..."

# Create .env file
cp .env.example .env

# Update .env file

# Update APP_ENV
sed -i -e "s/APP_ENV=.*/APP_ENV=development/g" .env

# Update Frontend Public Port
sed -i -e "s/REACT_HOSTPORT=.*/REACT_HOSTPORT=80/g" .env

touch "$SCRIPT_DIR/.first_run"

echo
