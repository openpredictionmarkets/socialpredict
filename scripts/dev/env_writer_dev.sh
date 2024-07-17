#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "${CALLED_FROM_SOCIALPREDICT}" ] && { echo "Not called from SocialPredict"; exit 42; }

# Function to create and update .env file
init_env() {
	# Create .env file
	cp "${SCRIPT_DIR}/.env.example" "${SCRIPT_DIR}/.env"

	# Update .env file

	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV='development'/g" "${SCRIPT_DIR}/.env"

}

if [[ ! -f "${SCRIPT_DIR}/.env" ]]; then
	echo "### First time running the script ..."
	echo "Let's initialize the application ..."
	sleep 1
	init_env
	echo "Application initialized successfully."
else
	read -p ".env file found. Do you want to re-create it? (y/N) " DECISION
	if [ "${DECISION}" != "Y" ] && [ "${DECISION}" != "y" ]; then
		:
	else
		sleep 1
		echo "Re-creating env file ..."
		sleep 1
		init_env
		echo ".env file re-created successfully."
	fi
fi

echo "Moving on to source_env..."
