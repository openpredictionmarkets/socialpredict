#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Function to create and update .env file
# Updated to be compatible with MacOS Sonoma.
# Uses POSTGRES_VOLUME to deal with MacOS xattrs provenence.
init_env() {
	# Create .env file
	cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV='development'/g" "$SCRIPT_DIR/.env"
	echo "ENV_PATH=$SCRIPT_DIR/.env" >> "$SCRIPT_DIR/.env"

	# Add OS-specific POSTGRES_VOLUME
	OS=$(uname -s)
	if [[ "$OS" == "Darwin" ]]; then
		echo "POSTGRES_VOLUME=pgdata:/var/lib/postgresql/data" >> "$SCRIPT_DIR/.env"
	else
		echo "POSTGRES_VOLUME=../data/postgres:/var/lib/postgresql/data" >> "$SCRIPT_DIR/.env"
	fi
}

if [[ ! -f "$SCRIPT_DIR/.env" ]]; then
	echo "### First time running the script ..."
	echo "Let's initialize the application ..."
	sleep 1
	init_env
	echo "Application initialized successfully."
else
	read -p ".env file found. Do you want to re-create it? (y/N) " DECISION
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
