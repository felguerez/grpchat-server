#!/bin/bash

# Check if the service type argument is provided
if [ "$#" -ne 1 ]; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

# Calculate the root directory of the project based on the script location
PROJECT_ROOT_DIR="$(dirname "$0")/.."
cd "$PROJECT_ROOT_DIR" || { echo "Could not find project root directory"; exit 1; }


# Fetch the latest commit hash
GIT_COMMIT_HASH=$(git rev-parse --short HEAD)
if [ -z "$GIT_COMMIT_HASH" ]; then
    echo "Error: Failed to get the latest git commit hash."
    exit 1
fi
echo "Latest commit hash is $GIT_COMMIT_HASH"

# Set variables based on the service type argument
SERVICE_TYPE=$1
TASK_DEF_PATH="$PROJECT_ROOT_DIR/deploy/ecs/ecs-task-definition-${SERVICE_TYPE}.json"
ECR_REPO="413025517373.dkr.ecr.us-east-1.amazonaws.com"
IMAGE_TAG="${ECR_REPO}/grpchat-${SERVICE_TYPE}:${GIT_COMMIT_HASH}"

# Update the image in the ECS task definition file
jq --arg image_tag "$IMAGE_TAG" --arg service_type "grpchat-${SERVICE_TYPE}" \
    '(.containerDefinitions[] | select(.name | startswith($service_type)) | .image) |= $image_tag' \
    "$TASK_DEF_PATH" > temp.json && mv temp.json "$TASK_DEF_PATH"

if [ $? -eq 0 ]; then
    echo "Updated $SERVICE_TYPE service image tag to $IMAGE_TAG in $TASK_DEF_PATH"
else
    echo "Failed to update $SERVICE_TYPE service image tag in $TASK_DEF_PATH"
    exit 1
fi
