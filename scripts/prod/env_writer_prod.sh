#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "${CALLED_FROM_SOCIALPREDICT}" ] && { echo "Not called from SocialPredict"; exit 42; }

# Function to check if the password is strong
check_password_strength() {
    local password=$1
    if [[ ${#password} -ge 8 ]] && [[ "${password}" =~ [A-Za-z] ]] && [[ "${password}" =~ [0-9] ]] && [[ "${password}" =~ [^A-Za-z0-9] ]]; then
        return 0  # Strong password
    else
        echo "Password must be at least 8 characters long, include letters, numbers, and special characters."
        return 1  # Not strong enough
    fi
}

check_username_strength() {
    local username=$1
    if [[ ${#username} -ge 5 ]] && [[ "${username}" =~ ^[A-Za-z0-9]+$ ]]; then
        return 0  # Strong username
    else
        echo "Username must be at least 5 characters."
        return 1  # Not strong enough
    fi
}

# Function to create and update .env file
init_env() {
	# Create .env file
	cp "${SCRIPT_DIR}/.env.example" "${SCRIPT_DIR}/.env"

	# Update .env file

	# Update APP_ENV
	sed -i -e "s/APP_ENV=.*/APP_ENV=production/g" .env

	# Update domain name
	read -r -p "What domain do you wish to use for the application? " domain_answer
	while [ -z "${domain_answer}" ]
	do
		echo "You need to specify a domain."
		read -r -p "What domain do you wish to use for the application? " domain_answer
	done

	# Change the Domain setting:
	sed -i -e "s/DOMAIN=.*/DOMAIN=${domain_answer}/g" .env
	echo "Setting DOMAIN to: ${domain_answer}"

	echo

	# Update email address
	read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
	while [ -z "${email_answer}" ]
	do
		echo "You need to specify an email address."
		read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
	done

	# Change the Email setting:
	sed -i -e "s/EMAIL=.*/EMAIL=${email_answer}/g" .env
	echo "Setting EMAIL to: ${email_answer}"

	echo

	# Update Database User:
	while true; do
		read -r -p "Specify a username for the Database User (default is insecure and not allowed): " db_user_answer
		if check_username_strength "${db_user_answer}"; then
			sed -i -e "s/POSTGRES_USER=.*/POSTGRES_USER=${db_user_answer}'/g" .env
			echo "Setting Database User to: ${db_user_answer}"
			break
		fi
	done

	echo

	# Update Database User Password:
	while true; do
		read -r -p "Specify password for the Database User (default is insecure and not allowed): " db_pass_answer
		if check_password_strength "${db_pass_answer}"; then
			sed -i -e "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD=${db_pass_answer}/g" .env
			echo "Setting Database Password to: ${db_pass_answer}"
			break
		fi
	done

	echo

	# Update Database Name:
	read -r -p "Specify the name for the Database (default: socialpredict_db): " db_name_answer
	if [ -z "${db_name_answer}" ]; then
		db_name_answer="socialpredict_db"  # Default value if no input
	fi
	sed -i -e "s/POSTGRES_DATABASE=.*/POSTGRES_DATABASE=${db_name_answer}/g" .env
	echo "Setting Database Name to: ${db_name_answer}"

	echo

	# Update Admin Password:
	while true; do
		read -r -p "Specify the password for the Admin User (default is insecure and not allowed): " admin_pass_answer
		if check_password_strength "${admin_pass_answer}"; then
			sed -i -e "s/ADMIN_PASSWORD=.*/ADMIN_PASSWORD=${admin_pass_answer}/g" .env
			echo "Setting Admin Password to: ${admin_pass_answer}"
			break
		fi
	done

}

if [[ ! -f "${SCRIPT_DIR}/.env" ]]; then
	echo "### First time running the script ..."
	echo "Let us initialize the application ..."
	sleep 1
	init_env
	echo "Application initialized successfully."
else
	read -p ".env file found. Do you want to re-create it? (y/N) " DECISION
	if [ "${DECISION}" != "Y" ] && [ "${DECISION}" != "y" ]; then
		:
	else
		sleep 1
		echo "Recreating .env file ..."
		sleep 1
		init_env
		echo ".env file re-created successfully."
	fi
fi

echo
