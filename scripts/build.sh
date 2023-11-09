#!/bin/bash

# Assign the first argument to a variable
SERVICE_TYPE=$1

# Check if the service type argument is provided and is valid
if [[ -z "$SERVICE_TYPE" ]] || { [ "$SERVICE_TYPE" != "http" ] && [ "$SERVICE_TYPE" != "grpc" ]; }; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

# Set variables
ECR_REPO="413025517373.dkr.ecr.us-east-1.amazonaws.com"
PROJECT_ROOT_DIR="$(dirname "$0")/.."

echo "Project directory is $PROJECT_ROOT_DIR"

# login
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPO

cd "$PROJECT_ROOT_DIR" || exit

# Build, tag, and push for specified service
if [ "$SERVICE_TYPE" = "http" ]; then
    SERVICE="grpchat-http"
    DOCKERFILE="Dockerfile.http"
else
    SERVICE="grpchat-grpc"
    DOCKERFILE="Dockerfile.grpc"
fi

docker build -t $SERVICE -f $DOCKERFILE .
docker tag $SERVICE:latest $ECR_REPO/$SERVICE:latest
docker push $ECR_REPO/$SERVICE:latest

echo "Build and push for $SERVICE_TYPE service complete."
