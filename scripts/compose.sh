#!/bin/bash

#!/bin/bash

#################################################################
#                                                               #
# Script By Vasileios Ntoufoudis                                #
# info@ntoufoudis.com                                           #
#                                                               #
#################################################################

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

if [ "$1" = "up" ]; then
	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" up -d
elif [ "$1" = "down" ]; then
	docker compose --env-file $ENV_PATH --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" down
else
	echo "Wrong Command."
	exit 1;
fi
