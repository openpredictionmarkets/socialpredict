#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Function to create and update .env file
init_env() {
	# Create .env file
	cp "$SCRIPT_DIR/.env.example" "$SCRIPT_DIR/.env"

	# Update .env file
	
	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV=production/g" .env

	# Update domain name
	read -r -p "What domain do you wish to use for the application? " domain_answer
	while [ -z "$domain_answer" ]
	do
		echo "You need to specify a domain."
		read -r -p "What domain do you wish to use for the application? " domain_answer
	done

	# Change the Domain setting:
	sed -i -e "s/DOMAIN=.*/DOMAIN='$domain_answer'/g" .env
	echo "Setting DOMAIN to: $domain_answer"

	echo

	# Update email address
	read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
	while [ -z "$email_answer" ]
	do
		echo "You need to specify an email address."
		read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
	done

	# Change the Email setting:
	sed -i -e "s/EMAIL=.*/EMAIL='$email_answer'/g" .env
	echo "Setting EMAIL to: $email_answer"

	echo

	# Update Database User:
	read -r -p "Specify username for the Database User. (default: user) " db_user_answer
	if [ ! -z "$db_user_answer" ]; then
		# Change DB User:
		sed -i -e "s/POSTGRES_USER=.*/POSTGRES_USER='$db_user_answer'/g" .env
		echo "Setting Database User to: $db_user_answer"
	fi

	echo

	# Update Database User Password:
	read -r -p "Specify password for the Database User. (default: password) " db_pass_answer
	if [ ! -z "$db_pass_answer" ]; then
		# Change DB Password:
		sed -i -e "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD='$db_pass_answer'/g" .env
		echo "Setting Database Password to: $db_pass_answer"
	fi

	echo

	# Update Database Name:
	read -r -p "Specify the name for the Database. (default: socialpredict_db) " db_name_answer
	if [ ! -z "$db_name_answer" ]; then
		# Change DB Name:
		sed -i -e "s/POSTGRES_DATABASE=.*/POSTGRES_DATABASE='$db_name_answer'/g" .env
		echo "Setting Database Password to: $db_name_answer"
	fi

	echo

	# Update Admin Password:
	read -r -p "Specify the password for the Admin User. (default: adminpass) " admin_pass_answer
	if [ ! -z "$admin_pass_answer" ]; then
		# Change Admin Password:
		sed -i -e "s/ADMIN_PASSWORD=.*/ADMIN_PASSWORD='$admin_pass_answer'/g" .env
		echo "Setting Admin Password to: $admin_pass_answer"
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
