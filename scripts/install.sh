#!/usr/bin/env bash

# Make sure the script can only be run via SocialPredict Script
[ -z "${CALLED_FROM_SOCIALPREDICT}" ] && { echo "Not called from SocialPredict"; exit 42; }

# Set FORCE to false
FORCE="n"

# Cross-platform "sed -i" (GNU vs BSD/macOS)
if sed --version >/dev/null 2>&1; then
  # GNU sed
  SED_INPLACE=(sed -i -e)
else
  # BSD sed (macOS) requires an empty backup extension: -i ''
  SED_INPLACE=(sed -i '' -e)
fi

# Check if Application is already running
check_if_running() {
  if [ -f "${SCRIPT_DIR}/.env" ]; then
    local app_env
    app_env=$(grep -i 'app_env' "${SCRIPT_DIR}/.env" | awk -F"=" '{print $NF}')
    if [ "${app_env}" == 'development' ]; then
      if [ -z "$(docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" ps -q)" ]; then
        :
      elif [ "$FORCE" == "y" ]; then
        docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" down -v
      else
        print_warning 'Application is already running.'
        print_error "Please stop it with './SocialPredict down' before proceeding with a new installation"
        exit 1
      fi
    elif [ "${app_env}" == 'localhost' ]; then
      if [ -z "$(docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" ps -q)" ]; then
        :
      elif [ "$FORCE" == "y" ]; then
        docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" down -v
      else
        print_warning 'Application is already running.'
        print_error "Please stop it with './SocialPredict down' before proceeding with a new installation"
        exit 1
      fi
    elif [ "${app_env}" == 'production' ]; then
      if [ -z "$(docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-prod.yaml" ps -q)" ]; then
        :
      elif [ "$FORCE" == "y" ]; then
        docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" down -v
      else
        print_warning 'Application is already running.'
        print_error "Please stop it with './SocialPredict down' before proceeding with a new installation"
        exit 1
      fi
    fi
  else
    :
  fi
}

# Check if .env file already exists
check_env() {
  if [ -f "${SCRIPT_DIR}/.env" ]; then
    if [ "$FORCE" == "y" ]; then
      rm "${SCRIPT_DIR}/.env"
      cp "${SCRIPT_DIR}/.env.example" "${SCRIPT_DIR}/.env"
    else
      read -r -p ".env file found. Do you want to re-create it? (y/N) " DECISION
      if [[ "${DECISION}" =~ ^([yY][eE][sS]|[yY])+$ ]]; then
    	  rm "${SCRIPT_DIR}/.env"
        cp "${SCRIPT_DIR}/.env.example" "${SCRIPT_DIR}/.env"
        print_status ".env file re-created successfully."
      else
        print_error "Aborting installation."
        exit 1
      fi
    fi
  else
    cp "${SCRIPT_DIR}/.env.example" "${SCRIPT_DIR}/.env"
  fi
}

# Check if Postgres data folder exists
check_postgres() {
  if [ -d "${SCRIPT_DIR}/data/postgres" ]; then
    if [ "$FORCE" == "y" ]; then
      sudo rm -rf "${SCRIPT_DIR}/data/postgres"
    else
      echo "Postgres Data Folder found."
      echo "Make sure to backup your data before proceeding."
      read -r -p "Do you want to remove Postgres Data folder? (y/N) " DECISION
      if [[ "${DECISION}" =~ ^([yY][eE][sS]|[yY])+$ ]]; then
        sudo rm -rf "${SCRIPT_DIR}/data/postgres"
        print_status "Postgres Data folder deleted successfully."
      else
        print_error "Aborting installation."
        exit 1
      fi
    fi
  else
    :
  fi
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

# Check if Docker Image exists on the system
check_image() {
  local image_name=$1
  local dockerfile=$2
  local directory=$3

  echo "### Checking for $image_name Image ..."
  if docker image inspect "$image_name" > /dev/null 2>&1; then
    read -r -p "$image_name Image Found. Do you want to re-build it? (y/N) " decision
    if [[ "$decision" =~ ^[Yy]$ ]]; then
      echo "Deleting Image ..."
      docker rmi "$image_name"
      echo "Image Deleted."
      build_image
    else
      :
    fi
  else
    echo "$image_name Image Not Found."
    build_image
  fi
}

# Build Docker Image
build_image() {
  echo "Building $image_name now."
  docker build --no-cache -t "$image_name" -f "$dockerfile" "$directory"
  echo "$image_name Image Built."
}

# Build procedure for Development Environment
build_dev() {
  BACKEND_IMAGE_NAME="${BACKEND_IMAGE_NAME:-socialpredict-dev-backend}"
  FRONTEND_IMAGE_NAME="${FRONTEND_IMAGE_NAME:-socialpredict-dev-frontend}"

  # Update APP_ENV
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV=development|" "${SCRIPT_DIR}/.env"

  # Add OS-specific POSTGRES_VOLUME
  OS=$(uname -s)
  if [[ "$OS" == "Darwin" ]]; then
    "${SED_INPLACE[@]}" "s|^POSTGRES_VOLUME=.*|POSTGRES_VOLUME=pgdata|" "${SCRIPT_DIR}/.env"
  else
    "${SED_INPLACE[@]}" "s|^POSTGRES_VOLUME=.*|POSTGRES_VOLUME=../data/postgres|" "${SCRIPT_DIR}/.env"
  fi

  # Change the Domain settings:
  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='localhost'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='http://localhost'|" "${SCRIPT_DIR}/.env"

  # Remove unnecessary lines from .env
  "${SED_INPLACE[@]}" "/^TRAEFIK_CONTAINER_NAME=.*/d" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "/^EMAIL=.*/d" "${SCRIPT_DIR}/.env"

  # Update Image Names
  "${SED_INPLACE[@]}" "s|^BACKEND_IMAGE_NAME=.*|BACKEND_IMAGE_NAME=${BACKEND_IMAGE_NAME}|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^FRONTEND_IMAGE_NAME=.*|FRONTEND_IMAGE_NAME=${FRONTEND_IMAGE_NAME}|" "${SCRIPT_DIR}/.env"

  print_status "Searching for Docker Images ..."

  DIRECTORY="${SCRIPT_DIR}"
  BACKEND_DOCKERFILE="${SCRIPT_DIR}/docker/backend/Dockerfile.dev"
  FRONTEND_DOCKERFILE="${SCRIPT_DIR}/docker/frontend/Dockerfile.dev"

  check_image "$BACKEND_IMAGE_NAME" "${BACKEND_DOCKERFILE}" "${DIRECTORY}"
  check_image "$FRONTEND_IMAGE_NAME" "${FRONTEND_DOCKERFILE}" "${DIRECTORY}"

  echo
  sleep 1;

  echo "Images built."
  echo "Use "./SocialPredict up" to start the containers"
  echo "And "./SocialPredict down" to stop them."
}

build_local() {
  # Update APP_ENV
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV=localhost|" "${SCRIPT_DIR}/.env"

  # Add OS-specific POSTGRES_VOLUME
  OS=$(uname -s)
  if [[ "$OS" == "Darwin" ]]; then
    "${SED_INPLACE[@]}" "s|^POSTGRES_VOLUME=.*|POSTGRES_VOLUME=pgdata|" "${SCRIPT_DIR}/.env"
  else
    "${SED_INPLACE[@]}" "s|^POSTGRES_VOLUME=.*|POSTGRES_VOLUME=../data/postgres|" "${SCRIPT_DIR}/.env"
  fi

  # Change the Domain settings:
  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='localhost'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='http://localhost'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^API_URL=.*|API_URL=http://localhost/api|" "${SCRIPT_DIR}/.env"

  # Remove unnecessary lines from .env
  "${SED_INPLACE[@]}" "/^TRAEFIK_CONTAINER_NAME=.*/d" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "/^EMAIL=.*/d" "${SCRIPT_DIR}/.env"

  # Pin platform per-service (helpful on Apple Silicon)
  cat > "${SCRIPT_DIR}/scripts/docker-compose.override.yml" <<EOF
services:
  backend:
    platform: ${FORCE_PLATFORM:-linux/amd64}
  frontend:
    platform: ${FORCE_PLATFORM:-linux/amd64}
  db:
    platform: ${FORCE_PLATFORM:-linux/amd64}
EOF
  echo "Wrote docker-compose.override.yml to pin platform = ${FORCE_PLATFORM:-linux/amd64}"

  docker compose \
    -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" \
    -f "${SCRIPT_DIR}/scripts/docker-compose.override.yml" \
    --env-file "${SCRIPT_DIR}/.env" pull
}

build_production() {
  # Update APP_ENV
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV=production|" "${SCRIPT_DIR}/.env"

  # Change the Domain settings:
  read -r -p "What domain do you wish to use for the application? " domain_answer
  while [ -z "$domain_answer" ]
  do
    echo "You need to specify a domain."
    read -r -p "What domain do you wish to use for the application? " domain_answer
  done

  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='$domain_answer'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='https://$domain_answer'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^API_URL=.*|API_URL=https://$domain_answer/api|" "${SCRIPT_DIR}/.env"

  echo "Setting DOMAIN to: $domain_answer"

  # Update Email Address
  read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
  while [ -z "$email_answer" ]
  do
    echo "You need to specify an email address."
    read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
  done

  template="${SCRIPT_DIR}/data/traefik/config/traefik.template"
  file="${SCRIPT_DIR}/data/traefik/config/traefik.yaml"
  export EMAIL="$email_answer"
  envsubst < "$template" > "$file"

  echo "Setting EMAIL to: $email_answer"

  # Update Database User Password:
  local db_pass
  db_pass=$(generate_password)
  "${SED_INPLACE[@]}" "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD='${db_pass}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Database Password"

  # Update Admin Password:
  ADMIN_PASS=$(generate_password)
  "${SED_INPLACE[@]}" "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD='${ADMIN_PASS}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Admin Password"

  # Pull images
  echo "Pulling images ..."
  docker compose --env-file "${SCRIPT_DIR}"/.env --file "${SCRIPT_DIR}/scripts/docker-compose-prod.yaml" pull
  echo

  echo "Images pulled."
  echo
  echo "Your admin credentials are:"
  echo "Username: admin"
  echo "Password: $ADMIN_PASS"
}

build_production_args() {
  # Update APP_ENV
  "${SED_INPLACE[@]}" "s|^APP_ENV=.*|APP_ENV=production|" "${SCRIPT_DIR}/.env"

  # Change the Domain settings:
  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='$1'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='https://$1'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^API_URL=.*|API_URL=https://$1/api|" "${SCRIPT_DIR}/.env"

  echo "Setting DOMAIN to: $1"

  # Update Email Address
  template="${SCRIPT_DIR}/data/traefik/config/traefik.template"
  file="${SCRIPT_DIR}/data/traefik/config/traefik.yaml"
  export EMAIL="$2"
  envsubst < "$template" > "$file"

  echo "Setting EMAIL to: $2"

  # Update Database User Password:
  local db_pass
  db_pass=$(generate_password)
  "${SED_INPLACE[@]}" "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD='${db_pass}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Database Password"

  # Update Admin Password:
  ADMIN_PASS=$(generate_password)
  "${SED_INPLACE[@]}" "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD='${ADMIN_PASS}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Admin Password"

  # Pull images
  echo "Pulling images ..."
  docker compose --env-file "${SCRIPT_DIR}"/.env --file "${SCRIPT_DIR}/scripts/docker-compose-prod.yaml" pull
  echo

  echo "Images pulled."
  echo
  echo "Your admin credentials are:"
  echo "Username: admin"
  echo "Password: $ADMIN_PASS"
}

if [ "$#" -eq 0 ]; then
  # Check if SocialPredict is running
  check_if_running

  # Check if .env file already exists
  check_env

  # Check if Postgres Data folder exists
  check_postgres

  # Echo initial message
  print_status "Building and Deploying SocialPredict ..."
  echo

  # Ask user input for Application Environment
  echo "### Select Application Environment: "
  PS3="Please enter your choice: "
  options=("Development" "Localhost" "Production" "Quit")
  select opt in "${options[@]}"
  do
    case $opt in
      "Development")
        build_dev
        break
        ;;
      "Localhost")
        build_local
        break
        ;;
      "Production")
        build_production
        break
        ;;
      "Quit")
        break
        ;;
      *)
        echo "Invalid option $REPLY"
        ;;
    esac
  done
else
  while getopts ":e:d:m:" opt; do
    case $opt in
      e)
        if [ "$OPTARG" != "development" ] && [ "$OPTARG" != "localhost" ] && [ "$OPTARG" != "production" ]; then
          print_error "Wrong environment selection."
          print_status "Acceptable environments: 'development', 'localhost', 'production'"
          exit 1
        else
          env="$OPTARG"
        fi
        ;;
      d)
        domain="$OPTARG"
        ;;
      m)
        email="$OPTARG"
        ;;
      \?)
        echo "Invalid option: -$OPTARG"
        ;;
      :)
        echo "Option -$OPTARG requires an argument."
        ;;
    esac
  done
  FORCE="y"
  # Check if SocialPredict is running
  check_if_running

  # Check if .env file already exists
  check_env

  # Check if Postgres Data folder exists
  check_postgres

  if [ "$env" == "development" ]; then
    build_dev
  elif [ "$env" == "localhost" ]; then
    build_local
  elif [ "$env" == "production" ]; then
    build_production_args "$domain" "$email"
  fi
fi

