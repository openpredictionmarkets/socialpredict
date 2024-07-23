#!/bin/sh

JWT_SIGNING_KEY=$(openssl rand -base64 32)
export JWT_SIGNING_KEY

# Function to start the server
start_server() {
    go build -o /usr/local/bin/main && /usr/local/bin/main &
    SERVER_PID=$!
    echo "Server started with PID ${SERVER_PID}"
}

# Function to stop the server
stop_server() {
    if [ -n "${SERVER_PID}" ]; then
        echo "Stopping server with PID ${SERVER_PID}"
        kill "${SERVER_PID}" || true
    fi
}

# Trap SIGTERM and SIGINT to perform a clean shutdown
trap 'stop_server; exit 0' TERM INT

# Start the server for the first time
start_server

# Keep the script running and wait for file changes
while true; do
    #inotifywait -e modify -r .
    #stop_server
    #start_server
    sleep 1
done
