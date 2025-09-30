#!/bin/sh
set -eu

TEMPLATE="/usr/share/nginx/html/env-config.js.template"
TARGET="/usr/share/nginx/html/env-config.js"

# Defaults if not set
: "${DOMAIN_URL:=http://localhost}"
: "${API_URL:=http://localhost:8080}"

if [ -f "$TEMPLATE" ]; then
  cp "$TEMPLATE" "$TARGET"
  if command -v ep >/dev/null 2>&1; then
    ep -v "$TARGET"
  else
    # Fallback simple substitution
    sed -i "s|\${DOMAIN_URL:-http://localhost}|${DOMAIN_URL}|g" "$TARGET"
    sed -i "s|\${API_URL:-http://localhost}|${API_URL}|g" "$TARGET"
  fi
fi

exec nginx -g "daemon off;"