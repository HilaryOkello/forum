#!/bin/bash

# Set the name of the Docker image and container
IMAGE_NAME="forum-app"
CONTAINER_NAME="forum-app"

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "Docker is not running. Please start Docker and try again."
    exit 1
fi

# Stop and remove any existing container with the same name
if docker ps -a --format '{{.Names}}' | grep -q $CONTAINER_NAME; then
    echo "Stopping and removing existing container '$CONTAINER_NAME'..."
    docker stop $CONTAINER_NAME
    docker rm $CONTAINER_NAME
fi

# Build the Docker image
echo "Building Docker image '$IMAGE_NAME'..."
docker build -t $IMAGE_NAME .

# Run the Docker container
echo "Running Docker container '$CONTAINER_NAME'..."
docker run -d -p 8080:8080 --name $CONTAINER_NAME $IMAGE_NAME

# Check if the container is running
if docker ps --format '{{.Names}}' | grep -q $CONTAINER_NAME; then
    echo "Container '$CONTAINER_NAME' is running on port 8080."
else
    echo "Failed to start the container."
    exit 1
fi
