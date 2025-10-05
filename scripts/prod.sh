#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Check for .env file
export CALLED_FROM_SOCIALPREDICT=yes
source "$SCRIPT_DIR/scripts/prod/env_writer_prod.sh"
_main
unset CALLED_FROM_SOCIALPREDICT

source_env

sleep 1;

# Build Docker Images
export CALLED_FROM_SOCIALPREDICT=yes
source "$SCRIPT_DIR/scripts/prod/build_prod.sh"
unset CALLED_FROM_SOCIALPREDICT
