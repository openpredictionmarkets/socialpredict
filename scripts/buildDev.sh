#!/bin/bash

#################################################################
# Script by Vasileios Ntoufoudis                                #
# info@ntoufoudis.com                                           #
#################################################################

set -e
set -a

#################################################################
#              List of Functions used in the script             #
#################################################################

# Function to replace API_URI in frontentd/src/config.js
frontend_api_uri() {
        template="$SCRIPT_DIR/../frontend/src/config.js.template"
        file="$SCRIPT_DIR/../frontend/src/config.js"
        if [ ${USE_DOMAIN} == 1 ]; then
                export DOMAIN="'https://${DOMAIN}'"
                envsubst < $template > $file
        else
                export DOMAIN="'http://localhost:8089'"
                envsubst < $template > $file
        fi
}

# Function to check if a command exists
command_exists() {
        type "$1" &> /dev/null
}

# Function to check if Docker && Docker Compose are installed
docker_check() {
        if command_exists docker-compose; then
                echo "Found docker-compose."
		COMPOSE='docker-compose'
        elif command_exists docker && docker compose version &> /dev/null; then
                echo "Found docker compose."
		COMPOSE='docker compose'
        else
                echo "Error: Docker Compose is not installed."
                exit 1
        fi
}

# Function to initialize certbot and issue SSL Certificate
certbot_init() {
        DATA_PATH="$SCRIPT_DIR/../data/certbot"
        # Check if files already exist
        if [ -d $DATA_PATH ]; then
                read -p "Existing data found for ${DOMAIN}. Continue and replace existing certificate? (y/N) " DECISION
                if [ "$DECISION" != "Y" ] && [ "$DECISION" != "y" ]; then
                        exit
                else
                        rm -Rf $DATA_PATH
                        mkdir -p $DATA_PATH
                fi
        else
                mkdir -p $DATA_PATH
        fi

        # Download recommended TLS parameters.
        if [ ! -e "$DATA_PATH/conf/options-ssl.nginx.conf" ] || [ ! -e "$DATA_PATH/conf/ssl-dhparams.pem" ]; then
                echo "### Downloading recommended TLS parameters ..."
                mkdir -p "$DATA_PATH/conf"
                curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot-nginx/certbot_nginx/_internal/tls_configs/options-ssl-nginx.conf > "$DATA_PATH/conf/options-ssl-nginx.conf"
                curl -s https://raw.githubusercontent.com/certbot/certbot/master/certbot/certbot/ssl-dhparams.pem > "$DATA_PATH/conf/ssl-dhparams.pem"
                echo
        fi

        # Create dummy certificate
        echo "### Creating dummy certificate for ${DOMAIN} ..."
        path="/etc/letsencrypt/live/${DOMAIN}"
        mkdir -p "$DATA_PATH/conf/live/${DOMAIN}"
        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" run --rm --entrypoint "\
                openssl req -x509 -nodes -newkey rsa:4096 -days 1 \
                -keyout '$path/privkey.pem' \
                -out '$path/fullchain.pem' \
                -subj '/CN=localhost'" certbot

        echo
        echo "### Starting Webserver ..."
        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml"  up -d webserver
        echo

        echo "### Deleting dummy certificate for ${DOMAIN} ..."
        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" run --rm --entrypoint "\
                rm -Rf /etc/letsencrypt/live/${DOMAIN} && \
                rm -Rf /etc/letsencrypt/archive/${DOMAIN} && \
                rm -Rf /etc/letsencrypt/renewal/${DOMAIN}.conf" certbot
	echo
        echo "### Requesting Let's Encrypt Certificate for ${DOMAIN} ..."

        # Join ${DOMAIN} to -d args
        domain_args="-d ${DOMAIN}"

        # Select appropriate email arg
        case "${EMAIL}" in
                "") email_arg="--register-unsafely-without-email" ;;
                *) email_arg="--email ${EMAIL}" ;;
        esac

        # Enable staging mode if needed
        if [ ${STAGING} != "0" ]; then staging_arg="--staging"; fi

        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" run --rm --entrypoint "\
                certbot certonly --webroot -w /var/www/certbot \
                $staging_arg \
                $email_arg \
                $domain_args \
                --agree-tos \
                --no-eff-email \
                --force-renewal" certbot

        echo
        echo "### Reloading Webserver ..."
        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" exec webserver nginx -s reload


}

# Function to build frontend image
build_frontend() {
        if [ "$APP_ENV" == 'Development' ] || [ "$APP_ENV" == 'development' ]; then
                echo "### Building Frontend Image ..."

                # Get frontend directory
                FRONTEND_DIR="$( readlink -f "${SCRIPT_DIR}/../frontend" )"

                # Build Image
                docker build --progress=plain --no-cache -t "${FRONTEND_IMAGE_NAME}" -f "${FRONTEND_DIR}/Dockerfile" "${FRONTEND_DIR}/."
                echo "Frontend Image Built"
                echo
        elif [ "$APP_ENV" == 'Production' ] || [ "$APP_ENV" == 'production' ]; then
                echo "Currently we don't support Production Builds"
        else
                echo "Wrong APP_ENV in .env file"
                exit 1
        fi
}

# Function to build backend image
build_backend() {
        if [ "$APP_ENV" == 'Development' ] || [ "$APP_ENV" == 'development' ]; then
                echo "### Building Backend Image ..."

                # Get backend directory
                BACKEND_DIR="$( readlink -f "${SCRIPT_DIR}/../backend" )"

                # Build Image
                docker build --progress=plain --no-cache -t "${BACKEND_IMAGE_NAME}" -f "${BACKEND_DIR}/Dockerfile" "${BACKEND_DIR}/."
                echo "Backend Image Built"
                echo
        elif [ "$APP_ENV" == 'Production' ] || [ "$APP_ENV" == 'production' ]; then
                echo "Currently we don't support Production Builds"
        else
                echo "Wrong APP_ENV in .env file"
                exit 1
        fi
}

#################################################################
#                     End List of Functions                     #
#################################################################

#################################################################
#                    List Of Variables Used                     #
#################################################################

# Determine script's directory
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

# Calculate the absolute path to the .env file
ENV_PATH="$( readlink -f "${SCRIPT_DIR}/../.env" )"

# Get the list of current docker images and convert to array
IMAGES=$(docker images -a | awk '{print $1}')
IMAGES_ARRAY=(${IMAGES// / })

#################################################################
#                   End List Of Variables                       #
#################################################################

#################################################################
#                          Begin Logic                          #
#################################################################

# Echo initial message
echo "### Bulding and Deploying SocialPredict..."
echo
sleep 1;

# Check that docker is installed
echo "### Checking that docker compose is installed ..."
docker_check
echo
sleep 1;

# Source the .env file
echo "### Searching for .env file ..."
if [ -f "$ENV_PATH" ]; then
        source "$ENV_PATH"
        echo ".env file found."
else
        echo "Error: .env file not found."
        exit 1
fi

echo
sleep 1;

# Check that APP_ENV is Development
# Production is not currently supported
echo "### Checking APP_ENV value ..."
if [[ "${APP_ENV}" == "development" ]] || [[ "${APP_ENV}" == "Development" ]]; then
	echo "Environment is set to development. Continuing..."
else
	echo "Production Build is not currently supported."
	echo "Please switch to development."
	exit 1;
fi

echo
sleep 1;

# Search for images
echo "### Searching for Docker Images ..."
sleep 1;

# Check if backend image exists
echo "Searching for Backend Image ..."
sleep 1;
if [[ ${IMAGES_ARRAY[@]} =~ "${BACKEND_IMAGE_NAME}" ]]; then
        echo "Backend Image Found."
        sleep 1;
else
        echo "Backend Image Not Found."
        echo "Building now."
        sleep 1;
        build_backend
fi

# Check if frontend image exists
echo "Searching for Frontend Image ..."
sleep 1;
if [[ ${IMAGES_ARRAY[@]} =~ "${FRONTEND_IMAGE_NAME}" ]]; then
        echo "Frontend Image Found."
else
        echo "Frontend Image Not Found."
        echo "Building now."
        sleep 1;
        build_frontend
fi

echo
sleep 1;

# Check if a domain will be used for the deployment
echo "### Checking if a domain or localhost will be used ..."
sleep 1;

if [ ${USE_DOMAIN} == 1 ]; then
        echo "Using Domain ${DOMAIN} for the deployment."
	sleep 1;
	certbot_init
	$COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" down
	frontend_api_uri
	$COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-domain.yml" build
else
        echo "Using Localhost for the deployment."
	sleep 1;
	frontend_api_uri
        $COMPOSE --env-file "$ENV_PATH" --file "$SCRIPT_DIR/docker-compose-dev-localhost.yml" build
fi

#################################################################
#                           End Logic                           #
#################################################################
