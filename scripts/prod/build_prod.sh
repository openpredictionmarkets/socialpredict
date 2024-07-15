#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Get the list of current docker images and convert to array
IMAGES=$( docker images -a | awk '{print $1}' )
IMAGES_ARRAY=($IMAGES// / })

# Function to replace API_URI in frontend/src/config.js
frontend_api_uri() {
	template="$SCRIPT_DIR/frontend/src/config.js.template"
	file="$SCRIPT_DIR/frontend/src/config.js"
	export DOMAIN="'https://${DOMAIN}'"
	envsubst < $template > $file
}

# Function to build frontend image
build_frontend() {
        echo "### Building Frontend Image ..."

	# Update API_URI
	frontend_api_uri

        # Get frontend directory
        FRONTEND_DIR="$( readlink -f "$SCRIPT_DIR/frontend" )"

        # Build Image
        docker build --no-cache -t "${FRONTEND_IMAGE_NAME}" -f "$FRONTEND_DIR/Dockerfile" "$FRONTEND_DIR/."
        echo "Frontend Image Built"
}

# Function to build backend image
build_backend() {
        echo "### Building Backend Image ..."

        # Get backend directory
        BACKEND_DIR="$( readlink -f "$SCRIPT_DIR/backend" )"

        # Build Image
        docker build --no-cache -t "${BACKEND_IMAGE_NAME}" -f "$BACKEND_DIR/Dockerfile" "$BACKEND_DIR/."
        echo "Backend Image Built"
}

# Search for images
echo "### Searching for Docker Images ..."
sleep 1;

# Check if backend image exists
echo "Searching for Backend Image ..."
sleep 1;
if [[ ${IMAGES_ARRAY[@]} =~ "${BACKEND_IMAGE_NAME}" ]]; then
        read -p "Backend Image Found. Do you want to re-build it? (y/N) " DECISION

	if [ "$DECISION" = "Y" ] || [ "$DECISION" = "y" ]; then
		echo "Deleting Image..."
		docker rmi "${BACKEND_IMAGE_NAME}"
		echo "Image Deleted."
		echo "Re-building image."
		build_backend
	fi
        sleep 1;
else
        echo "Backend Image Not Found."
        echo "Building now."
        sleep 1;
        build_backend
fi

# Check if frontend image exists
echo
echo "Searching for Frontend Image ..."
if [[ ${IMAGES_ARRAY[@]} =~ "${FRONTEND_IMAGE_NAME}" ]]; then
        read -p "Frontend Image Found. Do you want to re-build it? (y/N) " DECISION

        if [ "$DECISION" = "Y" ] || [ "$DECISION" = "y" ]; then
                echo "Deleting Image..."
                docker rmi "${FRONTEND_IMAGE_NAME}"
                echo "Image Deleted."
                echo "Re-building image."
                build_frontend
        fi
        sleep 1;

else
        echo "Frontend Image Not Found."
        echo "Building now."
        sleep 1;
        build_frontend
fi

echo
sleep 1;

# Pull remaining images
echo "Pulling remaining images ..."
$COMPOSE $ENV_FILE --file "$SCRIPT_DIR/scripts/docker-compose-prod.yaml" pull db webserver certbot
echo

echo "Images built."
echo
