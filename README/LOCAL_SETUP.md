## Repo Purpose

WARNING: As of the time of writing this LOCAL_SETUP document, 2 July 2024, this repo is designed for usage in development version on a local machine. Of course it can be adapted to run in production and is a part of the roadmap, but this is not fully supported yet.

For more information about the project go to: [SocialPredict](https://github.com/openpredictionmarkets/socialpredict)

### Supported Systems

Currently, the project is supported on Ubuntu up to 23.10 LTS and MacOS. Windows users will need to install Ubuntu on WSL2. [Here is the link to the WSL installation guide.](https://learn.microsoft.com/en-us/windows/wsl/install) Docker assumes you will install it on WSL2. **WSL1 support is untested**.

### Setting Up the Project on Your Local Machine

- **Clone the Repository**: Download the repository to your local machine.
-- `git clone https://github.com/openpredictionmarkets/socialpredict.git`
- **Install Docker**: Install Docker on your local machine. [Here is the link to the Docker installation guide.](https://docs.docker.com/get-docker/) We are assuming the latest version of Docker as of the date of this document. Windows users will need to follow the instructions for installing Docker on WSL2. [Here is the link to set up the Docker Desktop backend on WSL2.](https://docs.docker.com/desktop/wsl/)
- **Install docker-compose**: Install docker-compose on your local machine. [Here is the link to the docker-compose installation guide](https://docs.docker.com/compose/install/). We are assuming the latest version of docker-compose as of the date of this document.

### Backend Setup:

- Navigate to the `backend/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

### Frontend Setup:

- Navigate to the `frontend/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

### Nginx Setup:

- Navigate to the `nginx/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

### Database Seeding:

- The database is seeded in the `/backend/main.go` with the lines:

```
	// Seed the users
	seed.SeedUsers(db)
	seed.SeedMarket(db)
```

- If you wish not to seed markets or users, you can coment out those lines.

### Running the Service:

- Navigate back to the root directory `socialpredict/` in your terminal.
- Run `./compose-dev` to spin up the service on Docker and open a port. Access the service at `localhost:8089`.

### Use the following login credentials:

- Username: user1, Password: password
- Username: user2, Password: password
- Access the API at `localhost:8089`. API endpoints can be viewed [here](localhost:8089/api/v0/markets), using the pattern `localhost:8089/api/v0/markets` for example.

### Shutting down the service:

- To shut down the service, navigate to the `socialpredict/` directory in your terminal.
  Run `./compose-dev-down` to destroy the databases and fully shut down the service.

### Development Tools

- In the main directory, there are three scripts:

* `./postgres-exec`
* `./postgres-backend`
* `./postgres-frontend`
