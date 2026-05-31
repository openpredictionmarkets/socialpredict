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

if [ -f "${SCRIPT_DIR}/scripts/lib/jwt_key.sh" ]; then
  source "${SCRIPT_DIR}/scripts/lib/jwt_key.sh"
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
          docker compose --env-file "${SCRIPT_DIR}/.env" -f "${SCRIPT_DIR}/scripts/docker-compose-prod.yaml" down -v
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
  LC_ALL=C tr -dc "$char_set" < /dev/urandom | head -c "$length"
  echo
}

ensure_jwt_signing_key() {
  local jwt_key=""
  if command -v apply_jwt_signing_key >/dev/null 2>&1; then
    jwt_key="$(apply_jwt_signing_key "${SCRIPT_DIR}/.env")"
  fi

  if [[ -z "$jwt_key" ]]; then
    jwt_key="$(generate_password)$(generate_password)$(generate_password)"
    if grep -q '^JWT_SIGNING_KEY=' "${SCRIPT_DIR}/.env"; then
      "${SED_INPLACE[@]}" "s|^JWT_SIGNING_KEY=.*|JWT_SIGNING_KEY='${jwt_key}'|" "${SCRIPT_DIR}/.env"
    else
      printf "\nJWT_SIGNING_KEY='%s'\n" "$jwt_key" >> "${SCRIPT_DIR}/.env"
    fi
  fi

  echo "Setting JWT Signing Key"
}

set_env_value() {
  local key="$1"
  local value="$2"
  if grep -q "^${key}=" "${SCRIPT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^${key}=.*|${key}=${value}|" "${SCRIPT_DIR}/.env"
  else
    printf "\n%s=%s\n" "${key}" "${value}" >> "${SCRIPT_DIR}/.env"
  fi
}

apply_rate_limit_values() {
  local login_rate="$1"
  local login_burst="$2"
  local general_rate="$3"
  local general_burst="$4"
  local cleanup_interval="$5"

  set_env_value "RATE_LIMIT_LOGIN_RATE_PER_SECOND" "${login_rate}"
  set_env_value "RATE_LIMIT_LOGIN_BURST" "${login_burst}"
  set_env_value "RATE_LIMIT_GENERAL_RATE_PER_SECOND" "${general_rate}"
  set_env_value "RATE_LIMIT_GENERAL_BURST" "${general_burst}"
  set_env_value "RATE_LIMIT_CLEANUP_INTERVAL" "${cleanup_interval}"
}

require_rate_limit_env_value() {
  local key="$1"
  local value="${!key:-}"
  if [ -z "${value}" ]; then
    print_error "RATE_LIMIT_PROFILE=env-file requires ${key} to be set by the sourced env overlay."
    exit 1
  fi
}

apply_rate_limit_env_file_values() {
  require_rate_limit_env_value "RATE_LIMIT_LOGIN_RATE_PER_SECOND"
  require_rate_limit_env_value "RATE_LIMIT_LOGIN_BURST"
  require_rate_limit_env_value "RATE_LIMIT_GENERAL_RATE_PER_SECOND"
  require_rate_limit_env_value "RATE_LIMIT_GENERAL_BURST"
  require_rate_limit_env_value "RATE_LIMIT_CLEANUP_INTERVAL"

  apply_rate_limit_values \
    "${RATE_LIMIT_LOGIN_RATE_PER_SECOND}" \
    "${RATE_LIMIT_LOGIN_BURST}" \
    "${RATE_LIMIT_GENERAL_RATE_PER_SECOND}" \
    "${RATE_LIMIT_GENERAL_BURST}" \
    "${RATE_LIMIT_CLEANUP_INTERVAL}"
}

apply_rate_limit_profile() {
  local profile="${1:-secure-default}"

  case "${profile}" in
    secure-default)
      apply_rate_limit_values "0.1" "3" "1" "10" "5m"
      ;;
    small-droplet-staging)
      apply_rate_limit_values "5" "20" "25" "50" "5m"
      ;;
    loadtest)
      apply_rate_limit_values "100" "200" "1000" "2000" "5m"
      ;;
    env-file)
      apply_rate_limit_env_file_values
      ;;
    custom)
      local login_rate
      local login_burst
      local general_rate
      local general_burst
      local cleanup_interval
      read -r -p "Login rate limit, requests per second? " login_rate
      read -r -p "Login rate limit burst? " login_burst
      read -r -p "General API rate limit, requests per second? " general_rate
      read -r -p "General API rate limit burst? " general_burst
      read -r -p "Rate limiter cleanup interval, e.g. 5m? " cleanup_interval
      apply_rate_limit_values "${login_rate}" "${login_burst}" "${general_rate}" "${general_burst}" "${cleanup_interval}"
      ;;
    *)
      print_error "Unknown rate-limit profile '${profile}'. Use secure-default, small-droplet-staging, loadtest, env-file, or custom."
      exit 1
      ;;
  esac

  echo "Setting Rate Limit Profile: ${profile}"
}

select_rate_limit_profile() {
  local profile="${RATE_LIMIT_PROFILE:-}"
  if [ -n "${profile}" ]; then
    apply_rate_limit_profile "${profile}"
    return
  fi

  echo "### Select Rate Limit Profile:"
  echo "1) secure-default - conservative public defaults"
  echo "2) small-droplet-staging - initial load-test profile for small DigitalOcean staging"
  echo "3) loadtest - permissive profile for temporary single-source k6 hosts"
  echo "4) env-file - use RATE_LIMIT_* values from the current shell environment"
  echo "5) custom - enter explicit values"
  read -r -p "Please enter your choice [1]: " profile_choice

  case "${profile_choice:-1}" in
    1) apply_rate_limit_profile "secure-default" ;;
    2) apply_rate_limit_profile "small-droplet-staging" ;;
    3) apply_rate_limit_profile "loadtest" ;;
    4) apply_rate_limit_profile "env-file" ;;
    5) apply_rate_limit_profile "custom" ;;
    *)
      print_error "Invalid rate-limit profile choice '${profile_choice}'."
      exit 1
      ;;
  esac
}

validate_tls_mode() {
  local tls_mode="${1:-https}"
  case "${tls_mode}" in
    https|http)
      ;;
    *)
      print_error "Unknown TLS mode '${tls_mode}'. Use https or http."
      exit 1
      ;;
  esac
}

configure_public_urls() {
  local domain="$1"
  local tls_mode="${2:-https}"
  local scheme="https"

  validate_tls_mode "${tls_mode}"
  if [ "${tls_mode}" = "http" ]; then
    scheme="http"
  fi

  "${SED_INPLACE[@]}" "s|^DOMAIN=.*|DOMAIN='${domain}'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^DOMAIN_URL=.*|DOMAIN_URL='${scheme}://${domain}'|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^API_URL=.*|API_URL=${scheme}://${domain}/api|" "${SCRIPT_DIR}/.env"
  set_env_value "PUBLIC_BASE_URL" "'${scheme}://${domain}'"
  set_env_value "TLS_MODE" "${tls_mode}"

  echo "Setting DOMAIN to: ${domain}"
  echo "Setting TLS Mode to: ${tls_mode}"
}

render_traefik_config() {
  local email="$1"
  local tls_mode="${2:-https}"
  local template
  local file="${SCRIPT_DIR}/data/traefik/config/traefik.yaml"

  validate_tls_mode "${tls_mode}"
  if [ "${tls_mode}" = "http" ]; then
    template="${SCRIPT_DIR}/data/traefik/config/traefik-http.template"
    envsubst < "${template}" > "${file}"
    echo "Setting Traefik edge to HTTP-only mode"
    return
  fi

  if [ -z "${email}" ]; then
    print_error "TLS mode https requires an email address for Let's Encrypt."
    exit 1
  fi

  template="${SCRIPT_DIR}/data/traefik/config/traefik.template"
  export EMAIL="${email}"
  envsubst < "${template}" > "${file}"
  echo "Setting EMAIL to: ${email}"
}

select_tls_mode() {
  local tls_mode="${TLS_MODE:-https}"
  echo "### Select Public Edge TLS Mode:" >&2
  echo "1) https - domain-backed HTTPS with Let's Encrypt" >&2
  echo "2) http - raw-IP HTTP mode for temporary load-test hosts" >&2
  read -r -p "Please enter your choice [1]: " tls_choice

  case "${tls_choice:-1}" in
    1) tls_mode="https" ;;
    2) tls_mode="http" ;;
    *)
      print_error "Invalid TLS mode choice '${tls_choice}'."
      exit 1
      ;;
  esac

  echo "${tls_mode}"
}

print_install_help() {
  cat <<'EOF'
Usage: ./SocialPredict install [OPTIONS]

Initialize SocialPredict.

Options:
  -e VALUE              Environment: development, localhost, production
  -d VALUE              Public domain or raw IP for production installs
  -m VALUE              Email for Let's Encrypt when --tls-mode https
  -r VALUE              Rate-limit profile
  --tls-mode VALUE      Public edge mode: https or http
  -h, --help            Show this help

Rate-limit profiles:
  secure-default        Conservative public defaults
  small-droplet-staging Initial staging/load-test profile
  loadtest              Permissive single-source k6 profile for temporary hosts
  env-file              Read RATE_LIMIT_* values from the shell environment
  custom                Prompt for explicit values in interactive production install

Temporary raw-IP load-test example:
  ./SocialPredict install -e production -d 45.55.227.1 -r loadtest --tls-mode http
  ./SocialPredict up

Production/domain example:
  ./SocialPredict install -e production -d example.com -m ops@example.com
  ./SocialPredict up
EOF
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
  set_env_value "PUBLIC_BASE_URL" "'http://localhost'"

  # Remove unnecessary lines from .env
  "${SED_INPLACE[@]}" "/^TRAEFIK_CONTAINER_NAME=.*/d" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "/^EMAIL=.*/d" "${SCRIPT_DIR}/.env"

  # Update Image Names
  "${SED_INPLACE[@]}" "s|^BACKEND_IMAGE_NAME=.*|BACKEND_IMAGE_NAME=${BACKEND_IMAGE_NAME}|" "${SCRIPT_DIR}/.env"
  "${SED_INPLACE[@]}" "s|^FRONTEND_IMAGE_NAME=.*|FRONTEND_IMAGE_NAME=${FRONTEND_IMAGE_NAME}|" "${SCRIPT_DIR}/.env"

  apply_rate_limit_profile "secure-default"

  ensure_jwt_signing_key

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

# Build procedure for Local Environment
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
  set_env_value "PUBLIC_BASE_URL" "'http://localhost'"

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

  apply_rate_limit_profile "secure-default"

  ensure_jwt_signing_key

  docker compose \
    -f "${SCRIPT_DIR}/scripts/docker-compose-local.yaml" \
    -f "${SCRIPT_DIR}/scripts/docker-compose.override.yml" \
    --env-file "${SCRIPT_DIR}/.env" pull
}

# Build procedure for Production Environment
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

  tls_mode="$(select_tls_mode)"
  configure_public_urls "${domain_answer}" "${tls_mode}"

  # Update Email Address
  email_answer=""
  if [ "${tls_mode}" = "https" ]; then
    read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
    while [ -z "$email_answer" ]
    do
      echo "You need to specify an email address."
      read -r -p "What email address do you wish to use for the SSL Certificate? " email_answer
    done
  fi

  render_traefik_config "${email_answer}" "${tls_mode}"

  select_rate_limit_profile

  # Update Database User Password:
  # The packaged production compose topology uses local Docker Postgres.
  # Keep TLS disabled for that in-container DB connection unless an operator
  # replaces the DB topology and opts into TLS explicitly.
  if grep -q '^DB_REQUIRE_TLS=' "${SCRIPT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^DB_REQUIRE_TLS=.*|DB_REQUIRE_TLS=false|" "${SCRIPT_DIR}/.env"
  else
    printf "\nDB_REQUIRE_TLS=false\n" >> "${SCRIPT_DIR}/.env"
  fi

  local db_pass
  db_pass=$(generate_password)
  "${SED_INPLACE[@]}" "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD='${db_pass}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Database Password"

  # Update Admin Password:
  ADMIN_PASS=$(generate_password)
  "${SED_INPLACE[@]}" "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD='${ADMIN_PASS}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Admin Password"

  ensure_jwt_signing_key

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

  local tls_mode="${4:-${TLS_MODE:-https}}"
  configure_public_urls "$1" "${tls_mode}"

  # Update Email Address
  render_traefik_config "$2" "${tls_mode}"

  apply_rate_limit_profile "${3:-${RATE_LIMIT_PROFILE:-secure-default}}"

  # Update Database User Password:
  # The packaged production compose topology uses local Docker Postgres.
  # Keep TLS disabled for that in-container DB connection unless an operator
  # replaces the DB topology and opts into TLS explicitly.
  if grep -q '^DB_REQUIRE_TLS=' "${SCRIPT_DIR}/.env"; then
    "${SED_INPLACE[@]}" "s|^DB_REQUIRE_TLS=.*|DB_REQUIRE_TLS=false|" "${SCRIPT_DIR}/.env"
  else
    printf "\nDB_REQUIRE_TLS=false\n" >> "${SCRIPT_DIR}/.env"
  fi

  local db_pass
  db_pass=$(generate_password)
  "${SED_INPLACE[@]}" "s|^POSTGRES_PASSWORD=.*|POSTGRES_PASSWORD='${db_pass}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Database Password"

  # Update Admin Password:
  ADMIN_PASS=$(generate_password)
  "${SED_INPLACE[@]}" "s|^ADMIN_PASSWORD=.*|ADMIN_PASSWORD='${ADMIN_PASS}'|" "${SCRIPT_DIR}/.env"
  echo "Setting Admin Password"

  ensure_jwt_signing_key

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
  env=""
  domain=""
  email=""
  rate_limit_profile="${RATE_LIMIT_PROFILE:-secure-default}"
  tls_mode="${TLS_MODE:-https}"
  while [ "$#" -gt 0 ]; do
    case "$1" in
      -e)
        if [ "$#" -lt 2 ]; then
          print_error "Option -e requires an argument."
          exit 1
        fi
        if [ "$2" != "development" ] && [ "$2" != "localhost" ] && [ "$2" != "production" ]; then
          print_error "Wrong environment selection."
          print_status "Acceptable environments: 'development', 'localhost', 'production'"
          exit 1
        fi
        env="$2"
        shift 2
        ;;
      -d)
        if [ "$#" -lt 2 ]; then
          print_error "Option -d requires an argument."
          exit 1
        fi
        domain="$2"
        shift 2
        ;;
      -m)
        if [ "$#" -lt 2 ]; then
          print_error "Option -m requires an argument."
          exit 1
        fi
        email="$2"
        shift 2
        ;;
      -r)
        if [ "$#" -lt 2 ]; then
          print_error "Option -r requires an argument."
          exit 1
        fi
        rate_limit_profile="$2"
        shift 2
        ;;
      --tls-mode|-t)
        if [ "$#" -lt 2 ]; then
          print_error "Option $1 requires an argument."
          exit 1
        fi
        tls_mode="$2"
        validate_tls_mode "${tls_mode}"
        shift 2
        ;;
      --help|-h)
        print_install_help
        exit 0
        ;;
      *)
        print_error "Invalid option: $1"
        print_install_help
        exit 1
        ;;
    esac
  done

  if [ -z "${env}" ]; then
    print_error "Missing required -e environment option."
    print_install_help
    exit 1
  fi

  if [ "$env" == "production" ]; then
    if [ -z "${domain}" ]; then
      print_error "Production install requires -d DOMAIN_OR_IP."
      exit 1
    fi
    if [ "${tls_mode}" = "https" ] && [ -z "${email}" ]; then
      print_error "Production HTTPS install requires -m EMAIL for Let's Encrypt."
      exit 1
    fi
  fi

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
    build_production_args "$domain" "${email:-}" "$rate_limit_profile" "$tls_mode"
  fi
fi
