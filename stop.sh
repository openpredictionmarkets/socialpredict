#!/bin/bash
source .env

if [ ${USE_DOMAIN} == 1 ]; then
	docker compose --env-file .env --file "./scripts/docker-compose-dev-domain.yml" down
else
	docker compose --env-file .env --file "./scripts/docker-compose-dev-localhost.yml" down
fi
