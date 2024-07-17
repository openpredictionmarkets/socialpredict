#!/bin/bash

echo "Ensuring that script can only be run via SocialPredict script"
[ -z "${CALLED_FROM_SOCIALPREDICT}" ] && { echo "Not called from SocialPredict"; exit 42; }

echo "Setting BACKEND and FRONTEND image names."
BACKEND_IMAGE_NAME="socialpredict-backend:latest"
FRONTEND_IMAGE_NAME="socialpredict-frontend:latest"
echo "BACKEND_IMAGE_NAME: ${BACKEND_IMAGE_NAME}"
echo "FRONTEND_IMAGE_NAME: ${FRONTEND_IMAGE_NAME}"

# Function to build frontend image
build_frontend() {
        echo "### Building Frontend Image ..."

        # Get frontend directory
        FRONTEND_DIR="$( readlink -f "${SCRIPT_DIR}/frontend" )"

        # Build Image
        docker build --no-cache -t "${FRONTEND_IMAGE_NAME}" -f "${FRONTEND_DIR}/Dockerfile" "${FRONTEND_DIR}/."
        echo "Frontend Image Built"
}

# Function to build backend image
build_backend() {
        echo "### Building Backend Image ..."

        echo "Get backend directory."
        BACKEND_DIR="$( readlink -f "${SCRIPT_DIR}/backend" )"
        echo "BACKEND_DIR set to ${BACKEND_DIR}"

        echo "Building image with BACKEND_IMAGE_NAME: ${BACKEND_IMAGE_NAME}"
        docker build --no-cache -t "${BACKEND_IMAGE_NAME}" -f "${BACKEND_DIR}/Dockerfile" "${BACKEND_DIR}/."
        echo "Backend Image Built"
}

echo "### Searching for Docker Images ..."
sleep 1;

echo "Checking and building backend image...${BACKEND_IMAGE_NAME} in directory ${BACKEND_DIR}"
build_backend


echo "Checking and building frontend image..."
build_frontend

echo
sleep 1;

echo "Images built."
echo "Use "./SocialPredict up" to start the containers"
echo "And "./SocialPredict down" to stop them."
