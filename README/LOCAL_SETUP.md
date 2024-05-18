### Setting Up the Project on Your Local Machine

- **Clone the Repository**: Download the repository to your local machine. (`git clone https://github.com/openpredictionmarkets/socialpredict.git`)
- **Install Docker**: Install Docker on your local machine. [Here is the link to the Docker installation guide.](https://docs.docker.com/get-docker/)
- **Currently supported platforms**: Currently, the project is supported on Ubuntu and MacOS. Windows support is not guaranteed.

# Backend Setup:

- Navigate to the `backend/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

# Frontend Setup:

- Navigate to the `frontend/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

# Nginx Setup:

- Navigate to the `nginx/` directory in your terminal.
- Run `./builddockerfile` to build and tag the Dockerfile.

# Database Seeding:

- Uncomment the lines here to seed the database. [Here is the reference file.](https://github.com/openpredictionmarkets/socialpredict/blob/c52ad85ee20cc2ab347598db543fb29ad05c45d9/backend/main.go#L32) `line number 33,34,35`

# Running the Service:

- Navigate back to the root directory `socialpredict/` in your terminal.
- Run `docker-compose up` to spin up the service on Docker and open a port. Access the service at `localhost:8089`.

# Use the following login credentials:

- Username: user1, Password: password
- Username: user2, Password: password
- Access the API at `localhost:8089`. API endpoints can be viewed [here](localhost:8089/api/v0/markets), using the pattern `localhost:8089/api/v0/markets` for example.

# Shutting down the service:

- To shut down the service, navigate to the `socialpredict/` directory in your terminal.
- Run `docker-compose down` to destroy the databases and fully shut down the service.

# Note on Backend Changes:

- Any changes to the backend will trigger new seeds to the database unless the lines here are commented out. Failure to do so may result in odd graph behavior due to new bets being entered.

Shutting Down:

- To shut down the service, navigate to the `socialpredict/` directory in your terminal.
  Run `docker-compose down` to destroy the databases and fully shut down the service.
