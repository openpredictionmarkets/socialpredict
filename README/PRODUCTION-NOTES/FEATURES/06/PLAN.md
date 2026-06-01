# Temporary Load-Test Droplets Plan

Status: implemented baseline
Date: 2026-05-31

## 01. Planning Docs

Checklist:

- [x] Add feature overview.
- [x] Add design document.
- [x] Add implementation checklist.
- [x] Cross-link from production-notes index.

## 02. Installer Rate-Limit Profile

Checklist:

- [x] Add `loadtest` to supported rate-limit profiles.
- [x] Include `loadtest` in interactive install profile choices.
- [x] Update unknown-profile error copy.

## 03. Installer TLS Mode

Checklist:

- [x] Add `--tls-mode https|http` to non-interactive install args.
- [x] Keep `https` as the default.
- [x] In `https` mode, preserve current behavior.
- [x] In `http` mode, write `DOMAIN_URL`, `API_URL`, and `PUBLIC_BASE_URL` with `http://`.
- [x] In `http` mode, render HTTP-only Traefik config without Let's Encrypt.
- [x] Reject unsupported TLS modes.

## 04. Documentation

Checklist:

- [x] Document temporary raw-IP load-test install command.
- [x] Document that `-e production` remains correct for load-test hosts.
- [x] Document that `--tls-mode http` is not for model-office/production domains.

## 05. Verification

Checklist:

- [x] Run shell syntax check on changed scripts.
- [x] Run non-interactive install smoke in a temporary directory with Docker stubbed.
- [x] Confirm generated `.env` uses HTTP URLs for `--tls-mode http`.
- [x] Confirm generated Traefik config has no HTTPS redirect or Let's Encrypt resolver for HTTP mode.
