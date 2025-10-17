#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

init_env() {
	# Create .env file
	cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV='localhost'/g" "$SCRIPT_DIR/.env"

	# Add OS-specific POSTGRES_VOLUME
	OS=$(uname -s)
	if [[ "$OS" == "Darwin" ]]; then
		echo "POSTGRES_VOLUME=pgdata" >> "$SCRIPT_DIR/.env"
	else
		echo "POSTGRES_VOLUME=../data/postgres" >> "$SCRIPT_DIR/.env"
	fi

	# Update domain name

  	# Change the Domain setting:
  	sed -i -e "s/DOMAIN=.*/DOMAIN='localhost'/g" .env
  	sed -i -e 's/DOMAIN_URL=.*/DOMAIN_URL='\''http:\/\/'"localhost"''\''/g' .env
  	echo

  	# Remove unnecessary lines from .env
  	sed -i '/TRAEFIK_CONTAINER_NAME=.*/d' .env
  	sed -i '/EMAIL=.*/d' .env
}

if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
	echo "### First time running the script ..."
	echo "Let's initialize the application ..."
	sleep 1
	init_env
	echo "Application initialized successfully."
else
	read -r -p ".env file found. Do you want to re-create it? (y/N) " DECISION
	if [ "$DECISION" != "Y" ] && [ "$DECISION" != "y" ]; then
		:
	else
		sleep 1
		echo "Re-creating env file ..."
		sleep 1
		init_env
		echo ".env file re-created successfully."
	fi
fi

echo

sleep 1;

# Pull images
echo "Pulling images ..."
$COMPOSE $ENV_FILE --file "$SCRIPT_DIR/scripts/docker-compose-local.yaml" pull
echo

echo "Images pulled."
echo
echo "Your admin credentials are:"
echo "Username: admin"
echo "Password: password"
