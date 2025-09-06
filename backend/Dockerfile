# syntax=docker/dockerfile:1.3-labs
FROM golang:1.25-bookworm

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# Install Dependencies
RUN apt-get update && apt-get install -y --allow-downgrades \
    inotify-tools=3.22.6.0-4             \
    openssl=3.0.14-1~deb12u2

# Set the Working Directory inside the container
WORKDIR /backend

# Copy the go.mod and go.sum files
COPY go.mod go.sum ./

# Download all dependencies.
RUN go mod download

# Copy the source code from the current directory to the Working Directory inside the container
COPY . .

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT [ "./supervisor.sh"]
