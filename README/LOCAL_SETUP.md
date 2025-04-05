## Repo Purpose

For more information about the project go to: [SocialPredict](https://github.com/openpredictionmarkets/socialpredict)

### Supported Systems

Currently, the project is supported on Ubuntu up to 23.10 LTS and MacOS. Windows users will need to install Ubuntu on WSL2. [Here is the link to the WSL installation guide.](https://learn.microsoft.com/en-us/windows/wsl/install) Docker assumes you will install it on WSL2. **We cannot guarantee that everything here will work with WSL2. WSL1 support is untested**.

### Setting Up the Project on Your Local Machine

#### Prerequisites

##### MacOS

* Beyond docker as discussed below, you may also need to install `gettext` with: `brew install gettext` to get access to `envsubst` to run the install script.
* You may also need to create a data file, which is in the `.gitignore`, as MacOS docker may not create this automatically upon installation. You might have to run the following commands from the main `./socialpredict` directory:

```
mkdir -p data/postgres data/certbot
chown -R $(whoami):staff data
```

Next, for newer versions of MacOS, you may need to allow Docker within system settings, as well as terminal, because otherwise Apple adds some xattrs which maintain provenance over files, not allowing Docker to work with them properly. [This issue here explains a fix](https://github.com/docker/for-mac/issues/7636#issuecomment-2755395642).

#### Instructions

- **Clone the Repository**: Download the repository to your local machine.
-- `git clone https://github.com/openpredictionmarkets/socialpredict.git`
- **Install Docker**: Install Docker on your local machine. [Here is the link to the Docker installation guide.](https://docs.docker.com/get-docker/) We are assuming the latest version of Docker as of the date of this document. Windows users will need to follow the instructions for installing Docker on WSL2. [Here is the link to set up the Docker Desktop backend on WSL2.](https://docs.docker.com/desktop/wsl/)
- **Install docker compose**: Install `docker compose` on your local machine. [Here is the link to the docker compose installation guide](https://docs.docker.com/compose/install/). We are assuming the latest version of docker compose as of the date of this document. NOTE: `docker-compose` is deprecated and the command to use should be `docker compose`.
- If you have not run this previously, within the root of the `./socialpredict` directory, create a `.env` file which will be ignored in the gitignore.
- Within the root of the `./socialpredict` repo, run the command, `./SocialPredict install`.
- Select (1) Development, then select to rebuild the `.env` file and rebuild the images.
- When this has completed, run `./SocialPredict up`...this will use `docker compose` to spin up the images into containers on your machine.
- Once this is fully up and running, you can then visit `localhost` and log in to use the app.

#### Use the following login credentials:

- Access the API at `localhost` in your browser
- Sign in with the following Username/Password:
- Username: `admin`
- Password: `password`
- Navigate to "Dashboard" and create a new user
- Log in as the new user
- Change the default password
- Log back out
- Log back in with your new password
- The new user will have an initial account balance and maximum debt as configured in `setup.yaml`. **You can configure `setup.yaml` yourself, but must do so prior to installing SocialPredict**.

### Shutting down the service:

- To shut down the service, navigate to the `./socialpredict` root directory in your terminal.
- Run `./SocialPredict down`, which will use `docker compose down` to stop services.
- The data in the database will be maintained for the next run.

### Development Tools and Troubleshooting

#### Entering a Container

##### Database

* Log into the psql service directly:

```
docker exec -it -e PGPASSWORD=password socialpredict-postgres-container psql -U user -d socialpredict_db
```

* The postgres data is stored locally at `./data/postgres` and is mounted by `./scripts/docker-compose-dev.yaml` with:

```
    volumes:
      - ../data/postgres:/var/lib/postgresql/data
```

* The database on dev can be completely deleted and rebooted with the following, from `./` (root):

1. `./SocialPredict down`
2. `rm -rf ./data/postgres/*`
3. `./SocialPredict up`

##### Getting Logs from Different Containers

* There are different containers which serve different purposes in our app, and `docker compose` is used to spin them all up and hook them together. If something goes wrong in the app, you can use a `docker compose -p scripts logs` command to view logs from all of the containers.

* However, there are tons and tons of logs, so to be able to look at logs from a specific container, you should add `| grep {whatever}` where `wherever` is the container in question, to filter your logs to see what the problem may have been.

```
docker compose -p scripts logs | grep backend
```

* Errors from the `backend` container can be viewed by adding the `|` (pipe) and then `grep backend`, which is a way of filtering any lines which include, "backend," in them.

* Likewise, frontend, nginx, database, certbot errors can be filtered out similarly with:

```
docker compose -p scripts logs | grep backend
docker compose -p scripts logs | grep frontend
docker compose -p scripts logs | grep nginx
docker compose -p scripts logs | grep postgres
docker compose -p scripts logs | grep certbot
```

### Setting Up a Staging Instance On the Web

While we don't like to say we have an official, "prod," version yet per se, we do give the ability for you to run an staging instance online which for all intents and purposes is a functioning web app. However we don't endorse this app as being completely secure and recoverable without extra work. We're working on getting SocialPredict in a more reliable, secure state.

See [Stage Setup](/README/STAGE_SETUP.md)
