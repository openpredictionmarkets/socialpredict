# syntax=docker/dockerfile:1.3-labs
FROM golang:1.21-bookworm

SHELL ["/bin/bash", "-euo", "pipefail", "-c"]

# Install Dependencies
RUN go install github.com/cespare/reflex@v0.3.1

RUN apt-get update && apt-get install -y \
    inotify-tools=3.22.6.0-4

WORKDIR /

# Copy the source code from the current directory to the Working Directory inside the container
COPY . /backend

# Switch to /backend dir and build the Go app - output the binary as 'main'
WORKDIR /backend

# Expose port 8080 to the outside world
EXPOSE 8080

# Command to run the executable
ENTRYPOINT [ "./supervisor.sh"]