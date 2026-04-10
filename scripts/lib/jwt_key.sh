#!/usr/bin/env bash

# Helpers for keeping JWT_SIGNING_KEY entries in sync.
apply_jwt_signing_key() {
  local env_file="$1"
  local override="${2:-}"
  local python_cmd
  python_cmd=$(command -v python3 || command -v python || true)
  if [[ -z "$python_cmd" ]]; then
    printf 'Warning: python3 or python is required to ensure JWT_SIGNING_KEY in %s; please set the variable manually.\n' "$env_file" >&2
    return 0
  fi

  local key
  key=$(ENV_FILE="$env_file" KEY="$override" "$python_cmd" - <<'PY'
import json
import os
import pathlib
import re
import secrets
import sys

path = pathlib.Path(os.environ['ENV_FILE'])
override_raw = os.environ.get('KEY')
override = override_raw.strip() if override_raw is not None else ''

text = path.read_text()
match = re.search(r'^JWT_SIGNING_KEY=(.*)$', text, re.MULTILINE)
existing_value = None
if match:
    raw_value = match.group(1).strip()
    if raw_value:
        try:
            existing_value = json.loads(raw_value)
        except json.JSONDecodeError:
            if raw_value[0] in ('"', "'") and raw_value[-1] == raw_value[0]:
                existing_value = raw_value[1:-1]
            else:
                existing_value = raw_value

if override:
    new_key = override
elif existing_value:
    sys.stdout.write(existing_value)
    sys.exit(0)
else:
    new_key = secrets.token_hex(32)

line = f"JWT_SIGNING_KEY={json.dumps(new_key)}"
if match:
    text = re.sub(r'^JWT_SIGNING_KEY=.*$', line, text, flags=re.MULTILINE)
else:
    text = text.rstrip() + '\n' + line + '\n'

path.write_text(text)
sys.stdout.write(new_key)
PY
  )
  printf '%s' "$key"
}
