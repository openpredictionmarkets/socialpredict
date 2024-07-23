
#!/bin/bash

set -e # Stop script on error
set -a # Automatically export all variables

# Determine script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Calculate the absolute path to the .env file
ENV_PATH="$( readlink -f "$SCRIPT_DIR/.env" )"
ENV_FILE="--env-file $ENV_PATH"
source "$ENV_PATH"

if [ "$1" = "start" ]; then
    docker-compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" up -d
elif [ "$1" = "stop" ]; then
    docker-compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" stop
else
    echo "use ./start_service start or ./start_service stop!"
fi

