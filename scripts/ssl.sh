#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }


DOCKERFILE="--file $DOCKER_SSL"

# Function to replace API_URI in frontend/src/config.js
frontend_api_uri() {
        template="$SCRIPT_DIR/frontend/src/config.js.template"
        file="$SCRIPT_DIR/frontend/src/config.js"
        export DOMAIN="'https://${DOMAIN}'"
        envsubst < $template > $file
}

DATA_PATH="$SCRIPT_DIR/data/certbot"

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
$COMPOSE $ENV_FILE $DOCKER_SSL run --rm --entrypoint "\
	openssl req -x509 -nodes -newkey rsa:4096 -days 1 \
	-keyout '$path/privkey.pem' \
	-out '$path/fullchain.pem' \
	-subj '/CN=localhost'" certbot
echo
echo "### Starting Webserver ..."
$COMPOSE $ENV_FILE $DOCKER_SSL up -d webserver
echo

echo "### Deleting dummy certificate for ${DOMAIN} ..."
$COMPOSE $ENV_FILE $DOCKER_SSL run --rm --entrypoint "\
	rm -Rf /etc/letsencrypt/live/${DOMAIN} && \
	rm -Rf /etc/letsencrypt/archive/${DOMAIN} && \
	rm -Rf /etc/letsencrypt/renewal/${DOMAIN}.conf" certbot

echo
echo "### Requesting Let's Encrypt Certificate for ${DOMAIN} ..."

# Select appropriate email arg
case "${EMAIL}" in
	"") email_arg="--register-unsafely-without-email" ;;
	*) email_arg="--email ${EMAIL}" ;;
esac

# Enable staging mode if needed
if [ ${STAGING} != "0" ]; then staging_arg="--staging"; fi

$COMPOSE $ENV_FILE $DOCKER_SSL run --rm --entrypoint "\
	certbot certonly --webroot -w /var/www/certbot \
	$staging_arg \
	$email_arg \
	-d ${DOMAIN} \
	--agree-tos \
	--no-eff-email \
	--force-renewal" certbot

echo
echo "### Shutting down Webserver ..."
$COMPOSE $ENV_FILE $DOCKER_SSL down

# Update API_URI in Frontend
frontend_api_uri
