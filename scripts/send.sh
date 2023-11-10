#!/bin/bash

# Check if the service type argument is provided and is valid
if [[ "$#" -ne 1 ]] || { [ "$1" != "http" ] && [ "$1" != "grpc" ]; }; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

SERVICE_TYPE=$1
PROJECT_ROOT_DIR="$(dirname "$0")/.."  # Assumes the script is in a subdirectory of the project
cd $PROJECT_ROOT_DIR

# Step 1: Build and Push Docker Image
./scripts/build.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
    echo "Build and push for $SERVICE_TYPE failed."
    exit 1
fi

# Step 2: Update ECS Task Definition with the new image tag
./scripts/update_ecs_task.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
    echo "Update ECS task definition for $SERVICE_TYPE failed."
    exit 1
fi

# Step 3: Deploy the new version to ECS
./scripts/deploy.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
    echo "Deployment for $SERVICE_TYPE failed."
    exit 1
fi

echo "$SERVICE_TYPE service has been built, task definition updated, and deployment initiated."
