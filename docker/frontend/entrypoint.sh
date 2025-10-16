#!/bin/sh
set -eu

TEMPLATE="/usr/share/nginx/html/env-config.js.template"
TARGET="/usr/share/nginx/html/env-config.js"

# Defaults if not set
: "${DOMAIN_URL:=http://localhost}"
: "${API_URL:=http://localhost}"

if [ -f "$TEMPLATE" ]; then
  cp "$TEMPLATE" "$TARGET"

  sed -i "s|\${DOMAIN_URL:-http://localhost}|${DOMAIN_URL}|g" "$TARGET"
  sed -i "s|\${API_URL:-http://localhost:8080}|${DOMAIN_URL}/api|g" "$TARGET"

  chown nginx:nginx "$TARGET"

fi

exec nginx -g "daemon off;"
