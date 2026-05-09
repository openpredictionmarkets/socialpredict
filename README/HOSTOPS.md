# HostOps (Scaffold)

## Why this exists

`./HostOps` is a sibling tool to `./SocialPredict` for **host/infrastructure orchestration**.

- `./SocialPredict` owns application runtime operations on a host:
  - install / up / down / exec / backup
- `./HostOps` should own orchestration concerns:
  - connect to environments
  - run remote workflows in a controlled way
  - wrap Terraform and cloud-level operations

This keeps app runtime logic and infra control-plane logic separated.

## Boundary contract

`./HostOps` should orchestrate; it should not reimplement runtime behavior.

- Good: `./HostOps` runs `./SocialPredict install` remotely.
- Bad: `./HostOps` directly reproduces docker compose logic from `./SocialPredict`.

If runtime behavior changes, update `./SocialPredict`; `./HostOps` should keep calling stable `./SocialPredict` interfaces.

## Current status

Scaffold only.

- Implemented:
  - `./HostOps host ssh <staging|production>`
- Planned:
  - `./HostOps host env get <env> <KEY>`
  - `./HostOps host logs <env> <service>`
  - `./HostOps deploy <env>`
  - `./HostOps tf <plan|apply|destroy> <env>`

## DigitalOcean host convention

Default environment host mapping:

- `staging` -> `kconfs.com`
- `production` -> `brierfoxforecast.com`

Default key path convention:

- `~/.keys/openpredictionmarkets/staging/id_ed25519`
- `~/.keys/openpredictionmarkets/production/id_ed25519`

Override via environment variables:

- `HOSTOPS_STAGING_HOST`, `HOSTOPS_STAGING_USER`, `HOSTOPS_STAGING_KEY`
- `HOSTOPS_PRODUCTION_HOST`, `HOSTOPS_PRODUCTION_USER`, `HOSTOPS_PRODUCTION_KEY`

## Example setup

```bash
mkdir -p ~/.keys/openpredictionmarkets/staging ~/.keys/openpredictionmarkets/production
chmod 700 ~/.keys ~/.keys/openpredictionmarkets ~/.keys/openpredictionmarkets/staging ~/.keys/openpredictionmarkets/production

# Example: copy your private keys into the convention paths
cp ~/Downloads/do-staging-id ~/.keys/openpredictionmarkets/staging/id_ed25519
cp ~/Downloads/do-prod-id ~/.keys/openpredictionmarkets/production/id_ed25519
chmod 600 ~/.keys/openpredictionmarkets/staging/id_ed25519 ~/.keys/openpredictionmarkets/production/id_ed25519
```

Connect with:

```bash
./HostOps host ssh staging
./HostOps host ssh production
```

## Optional SSH config integration

You can also keep named SSH hosts:

```sshconfig
Host sp-staging
  HostName kconfs.com
  User root
  IdentityFile ~/.keys/openpredictionmarkets/staging/id_ed25519
  IdentitiesOnly yes

Host sp-production
  HostName brierfoxforecast.com
  User root
  IdentityFile ~/.keys/openpredictionmarkets/production/id_ed25519
  IdentitiesOnly yes
```

Then you may either:

- keep using `./HostOps host ssh <env>`, or
- call `ssh sp-staging` / `ssh sp-production` directly.

## Future extension ideas

`deploy`:

- resolve host from environment
- pull/checkout target revision
- run remote `./SocialPredict install ...`
- run remote `./SocialPredict up`
- run health checks and report status

`tf`:

- enforce environment-specific backend/workspace
- run `fmt`/`validate`/`plan` before `apply`
- keep plan artifacts for auditability
