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

Default SSH user and port:

- user -> `root`
- port -> `22`

Default key path convention:

- `~/.keys/openpredictionmarkets/staging/id_ed25519`
- `~/.keys/openpredictionmarkets/production/id_ed25519`

Default remote repository path convention:

- `/opt/socialpredict`

Override via environment variables:

- `HOSTOPS_STAGING_HOST`, `HOSTOPS_STAGING_HOST_IP`, `HOSTOPS_STAGING_USER`, `HOSTOPS_STAGING_PORT`, `HOSTOPS_STAGING_KEY`, `HOSTOPS_STAGING_REPO_PATH`
- `HOSTOPS_PRODUCTION_HOST`, `HOSTOPS_PRODUCTION_HOST_IP`, `HOSTOPS_PRODUCTION_USER`, `HOSTOPS_PRODUCTION_PORT`, `HOSTOPS_PRODUCTION_KEY`, `HOSTOPS_PRODUCTION_REPO_PATH`

`HOSTOPS_<ENV>_HOST` can be a domain or an IP address. Keep `HOSTOPS_<ENV>_HOST_IP` available as documentation and fallback even when DNS is the normal connection target.

## Per-command setup keys

`./HostOps host ssh <env>`:

- Required: SSH private key at `HOSTOPS_<ENV>_KEY` or `~/.keys/openpredictionmarkets/<env>/id_ed25519`
- Required: host via `HOSTOPS_<ENV>_HOST` or built-in default
- Optional: raw IP fallback via `HOSTOPS_<ENV>_HOST_IP`
- Optional: SSH user via `HOSTOPS_<ENV>_USER`, default `root`
- Optional: SSH port via `HOSTOPS_<ENV>_PORT`, default `22`

`./HostOps host env get <env> <KEY>` (planned):

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Reads from `/opt/socialpredict/.env` by convention

`./HostOps host logs <env> <service>` (planned):

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Expected service names should mirror `./SocialPredict exec` and compose services, such as `backend`, `frontend`, `postgres`, and `nginx`

`./HostOps deploy <env>` (planned):

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Should call remote `./SocialPredict install` and `./SocialPredict up`
- Should not duplicate Docker Compose behavior from `./SocialPredict`

`./HostOps tf <plan|apply|destroy> <env>` (planned):

- Needs Terraform environment directory, likely `infra/terraform/environments/<env>`
- Needs Terraform state/backend configuration, likely `infra/terraform/backend/<env>.hcl`
- Needs DigitalOcean API credentials outside the repo, for example `~/.keys/openpredictionmarkets/<env>/digitalocean.env`
- Should write local plan artifacts under `.context/infra-plans/<env>/`

## Example setup

```bash
mkdir -p ~/.keys/openpredictionmarkets/staging ~/.keys/openpredictionmarkets/production
chmod 700 ~/.keys ~/.keys/openpredictionmarkets ~/.keys/openpredictionmarkets/staging ~/.keys/openpredictionmarkets/production

# Example: copy your private keys into the convention paths
cp ~/Downloads/do-staging-id ~/.keys/openpredictionmarkets/staging/id_ed25519
cp ~/Downloads/do-prod-id ~/.keys/openpredictionmarkets/production/id_ed25519
chmod 600 ~/.keys/openpredictionmarkets/staging/id_ed25519 ~/.keys/openpredictionmarkets/production/id_ed25519
```

Example shell setup:

```bash
export HOSTOPS_STAGING_HOST=kconfs.com
export HOSTOPS_STAGING_HOST_IP=203.0.113.10
export HOSTOPS_STAGING_USER=root
export HOSTOPS_STAGING_PORT=22
export HOSTOPS_STAGING_KEY=~/.keys/openpredictionmarkets/staging/id_ed25519
export HOSTOPS_STAGING_REPO_PATH=/opt/socialpredict
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
  Port 22
  IdentityFile ~/.keys/openpredictionmarkets/staging/id_ed25519
  IdentitiesOnly yes

Host sp-production
  HostName brierfoxforecast.com
  User root
  Port 22
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
