#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Get the list of current docker images and convert to array
IMAGES=$( docker images -a | awk '{print $1}' )
IMAGES_ARRAY=($IMAGES// / })

# Function to build frontend image
build_frontend() {
        echo "### Building Frontend Image ..."

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

check_and_build_image() {
        local image_name=$1
        local directory=$2
        local dockerfile=$3

        echo "### Checking for $image_name Image ..."
        if docker image inspect "$image_name" > /dev/null 2>&1; then
                read -p "$image_name Image Found. Do you want to re-build it? (y/N) " decision
                if [[ "$decision" =~ ^[Yy]$ ]]; then
                echo "Deleting Image..."
                docker rmi "$image_name"
                echo "Image Deleted."
                else
                return
                fi
        else
                echo "$image_name Image Not Found."
        fi
        echo "Building $image_name now."
        docker build --no-cache -t "$image_name" -f "$directory/$dockerfile" "$directory"
        echo "$image_name Image Built"
}

# Search for images
echo "### Searching for Docker Images ..."
sleep 1;

# Backend and Frontend Image Names (Defaults can be overridden by environment variables)
BACKEND_IMAGE_NAME="${BACKEND_IMAGE_NAME:-socialpredict-backend:latest}"
FRONTEND_IMAGE_NAME="${FRONTEND_IMAGE_NAME:-socialpredict-frontend:latest}"

# Backend and Frontend Directory Paths
BACKEND_DIR="$( readlink -f "$SCRIPT_DIR/backend" )"
FRONTEND_DIR="$( readlink -f "$SCRIPT_DIR/frontend" )"

# Dockerfile Names
BACKEND_DOCKERFILE="Dockerfile"
FRONTEND_DOCKERFILE="Dockerfile"

# Check and build backend image
check_and_build_image "$BACKEND_IMAGE_NAME" "$BACKEND_DIR" "$BACKEND_DOCKERFILE"

# Check and build frontend image
check_and_build_image "$FRONTEND_IMAGE_NAME" "$FRONTEND_DIR" "$FRONTEND_DOCKERFILE"

echo
sleep 1;

echo "Images built."
echo "Use "./SocialPredict up" to start the containers"
echo "And "./SocialPredict down" to stop them."
