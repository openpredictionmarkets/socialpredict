#!/bin/bash

# Make sure the script can only be run via SocialPredict Script
[ -z "$CALLED_FROM_SOCIALPREDICT" ] && { echo "Not called from SocialPredict"; exit 42; }

# Check to see if this file is being run or sources from another script
_is_sourced() {
  # https://unix.stackexchange.com/a/215279
  [ "${#FUNCNAME[@]}" -ge 2 ] \
    && [ "${FUNCNAME[0]}" = '_is_sourced' ] \
    && [ "${FUNCNAME[1]}" = 'source' ]
}

# Function to generate a random password
generate_password() {
        local length=20
        # Define the character set for the password
        local char_set="A-Za-z0-9"

        # Generate a random password
        tr -dc "$char_set" < /dev/urandom | head -c "$length"
        echo
}

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
	sed -i -e 's/DOMAIN_URL=.*/DOMAIN_URL='\''https:\/\/'"$domain_answer"''\''/g' .env
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
	template="$SCRIPT_DIR/data/traefik/config/traefik.template"
        file="$SCRIPT_DIR/data/traefik/config/traefik.yaml"
        export EMAIL="$email_answer"
        envsubst < "$template" > "$file"

	sed -i -e "s/SSLEMAIL/$email_answer/g" ./data/traefik/config/traefik.yaml
	echo "Setting EMAIL to: $email_answer"

	echo

	# Update Database User Password:
	local db_pass
	db_pass=$(generate_password)
	sed -i -e "s/POSTGRES_PASSWORD=.*/POSTGRES_PASSWORD='${db_pass//&/\\&}'/g" .env
	echo "Setting Database Password"

	echo

	# Update Admin Password:
	ADMIN_PASS=$(generate_password)
	sed -i -e "s/ADMIN_PASSWORD=.*/ADMIN_PASSWORD='${ADMIN_PASS}'/g" .env
	echo "Setting Admin Password"
}

_main() {
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
}

if ! _is_sourced; then
  _main
fi
