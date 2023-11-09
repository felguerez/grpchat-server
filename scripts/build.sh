#!/bin/bash

# Check if the service type argument is provided and is valid
if [[ "$#" -ne 1 ]] || { [ "$1" != "http" ] && [ "$1" != "grpc" ]; }; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

SERVICE_TYPE=$1
ECR_REPO="413025517373.dkr.ecr.us-east-1.amazonaws.com"
PROJECT_ROOT_DIR="$(dirname "$0")/.."  # Assumes the script is in a subdirectory of the project

echo "Project directory is $PROJECT_ROOT_DIR"

# Fetch the latest git commit hash
GIT_COMMIT_HASH=$(git -C "$PROJECT_ROOT_DIR" rev-parse --short HEAD)
if [ -z "$GIT_COMMIT_HASH" ]; then
    echo "Error: Failed to get the latest git commit hash."
    exit 1
fi

echo "Latest commit hash is $GIT_COMMIT_HASH"

# Login to AWS ECR
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPO

# Build, tag, and push the Docker image
cd "$PROJECT_ROOT_DIR" || exit

DOCKERFILE_PREFIX="Dockerfile"
SERVICE_NAME="grpchat-${SERVICE_TYPE}"
DOCKERFILE="${DOCKERFILE_PREFIX}.${SERVICE_TYPE}"

docker build -t ${SERVICE_NAME}:${GIT_COMMIT_HASH} -f $DOCKERFILE .
docker tag ${SERVICE_NAME}:${GIT_COMMIT_HASH} ${ECR_REPO}/${SERVICE_NAME}:${GIT_COMMIT_HASH}
docker push ${ECR_REPO}/${SERVICE_NAME}:${GIT_COMMIT_HASH}

echo "Build and push for ${SERVICE_TYPE} service with commit ${GIT_COMMIT_HASH} complete."
