# Infrastructure: HostOps Scaffold

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
  - `./HostOps env list`
  - `./HostOps host ssh <env>`
  - `./HostOps host env get <env> <KEY>`
- Planned:
  - `./HostOps host logs <env> <service>`
  - `./HostOps deploy <env>`
  - `./HostOps tf <plan|apply|destroy> <env>`

## Local Environment Convention

HostOps treats each directory under this root as a local environment:

- `~/.keys/socialpredict/<env>/`

These are local operator settings on your laptop or workstation. They are not
GitHub Actions secrets, and they are not used by the Ansible deployment
workflows. GitHub/Ansible deploy secrets live in
`openpredictionmarkets/ansible_playbooks`; HostOps files live outside the repo
so a human can connect to or inspect a host after deployment.

This is intentionally directory-based. Your local key/config layout can mirror your cloud operations layout:

- `~/.keys/socialpredict/staging` for `kconfs.com`
- `~/.keys/socialpredict/mo` for `brierfoxforecast.com`
- `~/.keys/socialpredict/dev`, `prod`, `demo`, or any other environment name if your setup needs them

List the environments HostOps can see:

```bash
./HostOps env list
```

Each environment directory can contain:

- `hostops.env` for host/user/port/path settings
- `id_ed25519` for the SSH private key used by HostOps for that environment

HostOps reads per-machine configuration from:

- `~/.keys/socialpredict/<env>/hostops.env`

You can override that path for one command:

```bash
HOSTOPS_CONFIG=/path/to/hostops.env ./HostOps host ssh staging
```

You can also point HostOps at a different environment root:

```bash
HOSTOPS_CONFIG_ROOT=/path/to/socialpredict-keys ./HostOps env list
HOSTOPS_CONFIG_ROOT=/path/to/socialpredict-keys ./HostOps host ssh staging
```

The config file is intentionally outside the repository because it may point at private keys, server IPs, and future cloud credentials.

OpenPredictionMarkets environment conventions:

- `staging` -> `kconfs.com`
- `mo` -> `brierfoxforecast.com`

The `<env>` value is polymorphic: it is just the directory name under `~/.keys/socialpredict`. A different user can choose `prod`, `dev`, `demo`, or a single environment such as `site`.

Default SSH user and port:

- user -> `root`
- port -> `22`

Default key path convention:

- `~/.keys/socialpredict/staging/id_ed25519`
- `~/.keys/socialpredict/mo/id_ed25519`

Default remote repository path convention:

- `/opt/socialpredict`

Override via environment variables:

- Environment config file keys: `HOSTOPS_HOST`, `HOSTOPS_HOST_IP`, `HOSTOPS_USER`, `HOSTOPS_PORT`, `HOSTOPS_KEY`, `HOSTOPS_REPO_PATH`
- Shell override keys: `HOSTOPS_<ENV>_HOST`, `HOSTOPS_<ENV>_HOST_IP`, `HOSTOPS_<ENV>_USER`, `HOSTOPS_<ENV>_PORT`, `HOSTOPS_<ENV>_KEY`, `HOSTOPS_<ENV>_REPO_PATH`

`HOSTOPS_HOST` and `HOSTOPS_<ENV>_HOST` can be a domain or an IP address. Keep `HOSTOPS_HOST_IP` or `HOSTOPS_<ENV>_HOST_IP` available as documentation and fallback even when DNS is the normal connection target.

## DigitalOcean CLI authentication

SSH-only commands such as `./HostOps host ssh <env>` do not require DigitalOcean API access. Any HostOps command that inspects or changes DigitalOcean resources will require `doctl` authentication first.

Authenticate `doctl` locally before running DigitalOcean operations:

```bash
doctl auth init --context socialpredict
```

The command prompts for a DigitalOcean personal access token. Generate the token in the DigitalOcean control panel:

```text
https://cloud.digitalocean.com/account/api/tokens
```

Do not commit or paste the token into chat. `doctl` stores it in your local doctl config.

Verify authentication:

```bash
doctl auth list
doctl account get --context socialpredict
doctl compute droplet list --context socialpredict --format ID,Name,PublicIPv4,PrivateIPv4,Region,Status,Tags
```

If `socialpredict` is not the current context, either switch to it:

```bash
doctl auth switch --context socialpredict
```

or pass `--context socialpredict` on each `doctl` command.

Recommended local DigitalOcean credential convention for future HostOps cloud/Terraform commands:

```bash
~/.keys/socialpredict/<env>/digitalocean.env
```

That file should remain local and may contain non-repo settings such as:

```bash
DIGITALOCEAN_CONTEXT=socialpredict
```

The DigitalOcean context is not necessarily unique per environment. Multiple HostOps environments can point at the same DigitalOcean account/context while still keeping separate host, key, droplet, firewall, and Terraform settings under each `~/.keys/socialpredict/<env>/` directory.

Example shared account layout:

```text
~/.keys/socialpredict/staging/digitalocean.env  # DIGITALOCEAN_CONTEXT=socialpredict
~/.keys/socialpredict/mo/digitalocean.env       # DIGITALOCEAN_CONTEXT=socialpredict
```

Useful DigitalOcean inspection commands:

```bash
doctl compute droplet list --context socialpredict --format ID,Name,PublicIPv4,PrivateIPv4,Region,Status,Tags
doctl compute firewall list --context socialpredict --format ID,Name,Status,DropletIDs,Tags
```

## Config file format

Create this file locally:

```bash
~/.keys/socialpredict/<env>/hostops.env
```

Staging example:

```bash
HOSTOPS_HOST=kconfs.com
HOSTOPS_HOST_IP=203.0.113.10
HOSTOPS_USER=root
HOSTOPS_PORT=22
HOSTOPS_KEY=~/.keys/socialpredict/staging/id_ed25519
HOSTOPS_REPO_PATH=/opt/socialpredict
```

Model office example:

```bash
HOSTOPS_HOST=brierfoxforecast.com
HOSTOPS_HOST_IP=203.0.113.20
HOSTOPS_USER=root
HOSTOPS_PORT=22
HOSTOPS_KEY=~/.keys/socialpredict/mo/id_ed25519
HOSTOPS_REPO_PATH=/opt/socialpredict
```

Use shell syntax only: `KEY=value`, one setting per line. Do not commit this file.

## Per-command setup keys

`./HostOps env list`:

- Reads local directories under `~/.keys/socialpredict`
- Optional: use `HOSTOPS_CONFIG_ROOT` to list a different root
- Shows the resolved host plus whether each environment has `hostops.env` and an SSH key
- Does not connect to any server

`./HostOps host ssh <env>`:

- Required: SSH private key at `HOSTOPS_KEY`, `HOSTOPS_<ENV>_KEY`, or `~/.keys/socialpredict/<env>/id_ed25519`
- Required: host via `HOSTOPS_HOST`, `HOSTOPS_<ENV>_HOST`, or a built-in default
- Optional: raw IP fallback via `HOSTOPS_HOST_IP` or `HOSTOPS_<ENV>_HOST_IP`
- Optional: SSH user via `HOSTOPS_USER` or `HOSTOPS_<ENV>_USER`, default `root`
- Optional: SSH port via `HOSTOPS_PORT` or `HOSTOPS_<ENV>_PORT`, default `22`

`./HostOps host env get <env> <KEY>`:

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_REPO_PATH` or `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Reads from `/opt/socialpredict/.env` by convention
- Prints the matching `KEY=value` line, so use carefully for secrets

Example:

```bash
./HostOps host env get staging ADMIN_PASSWORD
```

`./HostOps host logs <env> <service>` (planned):

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_REPO_PATH` or `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Expected service names should mirror `./SocialPredict exec` and compose services, such as `backend`, `frontend`, `postgres`, and `nginx`

`./HostOps deploy <env>` (planned):

- Needs the same SSH keys as `host ssh`
- Needs remote repo path via `HOSTOPS_REPO_PATH` or `HOSTOPS_<ENV>_REPO_PATH`, default `/opt/socialpredict`
- Should call remote `./SocialPredict install` and `./SocialPredict up`
- Should not duplicate Docker Compose behavior from `./SocialPredict`

`./HostOps tf <plan|apply|destroy> <env>` (planned):

- Needs Terraform environment directory, likely `infra/terraform/environments/<env>`
- Needs Terraform state/backend configuration, likely `infra/terraform/backend/<env>.hcl`
- Needs authenticated `doctl` or DigitalOcean API credentials outside the repo, for example `~/.keys/socialpredict/<env>/digitalocean.env`
- Should write local plan artifacts under `.context/infra-plans/<env>/`

## Example setup

Staging example:

```bash
mkdir -p ~/.keys/socialpredict/staging
chmod 700 ~/.keys ~/.keys/socialpredict ~/.keys/socialpredict/staging

ssh-keygen -t ed25519 \
  -f ~/.keys/socialpredict/staging/id_ed25519 \
  -C "socialpredict-staging-hostops"

chmod 600 ~/.keys/socialpredict/staging/id_ed25519
```

Then add the public key to the staging VPS user's `authorized_keys`.

Local public key:

```bash
~/.keys/socialpredict/staging/id_ed25519.pub
```

Remote destination:

```bash
root@kconfs.com:~/.ssh/authorized_keys
```

For `mo` or another environment, use that environment's directory and host, for example `~/.keys/socialpredict/mo/id_ed25519`.

Create the config file:

```bash
$EDITOR ~/.keys/socialpredict/staging/hostops.env
```

Connect with:

```bash
./HostOps env list
./HostOps host ssh staging
./HostOps host ssh mo
```

## Optional SSH config integration

You can also keep named SSH hosts:

```sshconfig
Host sp-staging
  HostName kconfs.com
  User root
  Port 22
  IdentityFile ~/.keys/socialpredict/staging/id_ed25519
  IdentitiesOnly yes

Host sp-mo
  HostName brierfoxforecast.com
  User root
  Port 22
  IdentityFile ~/.keys/socialpredict/mo/id_ed25519
  IdentitiesOnly yes
```

Then you may either:

- keep using `./HostOps host ssh <env>`, or
- call `ssh sp-staging` / `ssh sp-mo` directly.

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
