## Repo Purpose

For more information about the project go to: [SocialPredict](https://github.com/openpredictionmarkets/socialpredict)

### Supported Systems

Currently, the project is supported on Ubuntu up to 23.10 LTS and MacOS. Windows users will need to install Ubuntu on WSL2. [Here is the link to the WSL installation guide.](https://learn.microsoft.com/en-us/windows/wsl/install) Docker assumes you will install it on WSL2. **WSL1 support is untested**.

### Setting Up the Project on Your Local Machine

- **Clone the Repository**: Download the repository to your local machine.
-- `git clone https://github.com/openpredictionmarkets/socialpredict.git`
- **Install Docker**: Install Docker on your local machine. [Here is the link to the Docker installation guide.](https://docs.docker.com/get-docker/) We are assuming the latest version of Docker as of the date of this document. Windows users will need to follow the instructions for installing Docker on WSL2. [Here is the link to set up the Docker Desktop backend on WSL2.](https://docs.docker.com/desktop/wsl/)
- **Install docker compose**: Install `docker compose` on your local machine. [Here is the link to the docker compose installation guide](https://docs.docker.com/compose/install/). We are assuming the latest version of docker compose as of the date of this document. NOTE: `docker-compose` is deprecated and the command to use should be `docker compose`.
- Within the root of the `./socialpredict` repo, run the command, `./SocialPredict install`.
- Select (2) Development, then select to rebuild the `.env` file and rebuild the imgages.
- When this has completed, run `./SocialPredict up`...this will use `docker compose` to spin up the images into containers on your machine.
- Once this is fully up and running, you can then visit `localhost` and log in to use the app.

#### Use the following login credentials:

- Access the API at `localhost` in your browser
- Sign in with the following Username/Password:
- Username: `admin`
- Password: `password`

### Shutting down the service:

- To shut down the service, navigate to the `./socialpredict` root directory in your terminal.
- Run `./SocialPredict down`, which will use `docker compose down` to stop services.
- The data in the database will be maintained for the next run.

### Development Tools

#### Entering a Container

##### Database

* Log into the psql service directly:

```
docker exec -it -e PGPASSWORD=password socialpredict-postgres-container psql -U user -d socialpredict_db
```

### Setting Up the Project on Production

See [./README/PROD_SETUP.md](./README/PROD_SETUP.md)