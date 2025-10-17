#!/bin/sh
set -eu

TEMPLATE="/app/public/env-config.js.template"
TARGET="/app/public/env-config.js"

# Defaults if not set
: "${DOMAIN_URL:=http://localhost}"
: "${API_URL:=http://localhost:8080}"

if [ -f "$TEMPLATE" ]; then
  cp "$TEMPLATE" "$TARGET"

  sed -i "s|\${DOMAIN_URL:-http://localhost}|${DOMAIN_URL}|g" "$TARGET"
  sed -i "s|\${API_URL:-http://localhost:8080}|${DOMAIN_URL}:${BACKEND_PORT}|g" "$TARGET"
fi

exec npm run start
