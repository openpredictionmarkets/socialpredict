#!/bin/bash

# Check if .env file is present.
echo "### Searching for .env file ..."
if [ ! -f ".env" ]; then
        echo "Error: .env file not found."
	echo "Copy .env.example file and modify it according to your needs."
	echo "cp .env.example .env"
	exit 1
fi

source ./scripts/buildDev.sh
