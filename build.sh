#!/bin/bash

set -e # Stop script on error
set -a # Automatically export all variables

#install docker, docker-comp
curl -fsSL https://test.docker.com -o test-docker.sh
sudo sh test-docker.sh
curl -L https://github.com/docker/compose/releases/download/v2.29.0/docker-compose-`uname -s`-`uname -m` > /usr/local/bin/docker-compose
sudo chmod +x /usr/local/bin/docker-compose
echo "docker install success !"

wget https://nodejs.org/dist/v20.15.1/node-v20.15.1-linux-x64.tar.gz    
tar zxvf  node-v20.15.1-linux-x64.tar.gz                         
node-v20.15.1-linux-x64/bin/node -v                         
ln -s /root/node-v20.15.1-linux-x64/bin/npm   /usr/local/bin/ 
ln -s /root/node-v20.15.1-linux-x64/bin/node   /usr/local/bin/ 

# Determine script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
# Calculate the absolute path to the .env file
ENV_PATH="$( readlink -f "$SCRIPT_DIR/.env" )"
ENV_FILE="--env-file $ENV_PATH"
source "$ENV_PATH"
BACKEND_DIR="$( readlink -f "$SCRIPT_DIR/backend" )"

# Build Image
echo "building backend docker image..."
docker build --no-cache -t "${BACKEND_IMAGE_NAME}" -f "$BACKEND_DIR/Dockerfile" "$BACKEND_DIR/."

echo "building frontend static file..."
mkdir -p ${DATA_DIR}/https-portal/${DOMAIN}
cp example.com.ssl.conf.erb  ${DATA_DIR}/https-portal
if [ "example.com" != ${DOMAIN} ]; then
    mv ${DATA_DIR}/https-portal/example.com.ssl.conf.erb  ${DATA_DIR}/https-portal/${DOMAIN}.ssl.conf.erb 
fi
sed -i "s#http://127.0.0.1:8080#http://${BACKEND_IP}:${BACKEND_HOSTPORT}#g" ${DATA_DIR}/https-portal/${DOMAIN}.ssl.conf.erb

cd frontend
npm install
npm run build
cp -R build/* ${DATA_DIR}/https-portal/${DOMAIN}
cd ..
echo "build success!"