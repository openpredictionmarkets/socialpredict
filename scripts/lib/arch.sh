#!/usr/bin/env bash
# scripts/lib/arch.sh
set -euo pipefail

# Detect host architecture
HOST_ARCH="$(uname -m || true)"

is_arm() {
  case "${HOST_ARCH}" in
    arm64|aarch64) return 0 ;;
    *) return 1 ;;
  esac
}

# Normalize env defaults (in case .env didn’t define them)
: "${BUILD_NATIVE_ARM64:=false}"
: "${FORCE_PLATFORM:=}"

if is_arm; then
  # On Apple Silicon:
  if [ "${APP_ENV:-}" = "development" ]; then
    if [ "${BUILD_NATIVE_ARM64}" = "true" ]; then
      # We’ll build native images; do not force emulation.
      export FORCE_PLATFORM="linux/arm64"
    else
      # Default: run amd64 images under emulation for max compatibility
      export FORCE_PLATFORM="linux/amd64"
      export DOCKER_DEFAULT_PLATFORM="linux/amd64"
    fi
  else
    # localhost/prod pull from registry; prefer emulation unless registry is multi-arch
    if [ -z "${FORCE_PLATFORM}" ]; then
      export FORCE_PLATFORM="linux/amd64"
      export DOCKER_DEFAULT_PLATFORM="linux/amd64"
    fi
  fi
fi
