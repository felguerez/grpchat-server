#!/bin/bash

# Function to print messages in color
print_message() {
  COLOR=$1
  TEXT=$2
  echo -e "\033[${COLOR}m${TEXT}\033[0m"
}

# Check if the service type argument is provided and is valid
if [[ "$#" -ne 1 ]] || { [ "$1" != "http" ] && [ "$1" != "grpc" ]; }; then
  print_message "31" "❌ Usage: $0 <http|grpc>"
  exit 1
fi

SERVICE_TYPE=$1
PROJECT_ROOT_DIR="$(dirname "$0")/.." # Assumes the script is in a subdirectory of the project
cd $PROJECT_ROOT_DIR
SERVICE_LABEL="grpchat-${SERVICE_TYPE}"

# Step 0: Check if the image with the same tag already exists in ECR
if aws ecr describe-images --repository-name "grpchat-${SERVICE_TYPE}" --image-ids imageTag="$GIT_COMMIT_HASH" >/dev/null 2>&1; then
  print_message "33" "🚀 Image with tag $GIT_COMMIT_HASH already exists in ECR for $SERVICE_LABEL."
else
  print_message "34" "🚀 Image with tag $GIT_COMMIT_HASH does not exist in ECR for $SERVICE_LABEL. Starting build and push..."
fi

# Step 1: Build and Push Docker Image
./scripts/build.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
  print_message "31" "❌ Build and push for $SERVICE_SERVICE_LABEL failed."
  exit 1
fi
print_message "32" "✅ Build and push for $SERVICE_SERVICE_LABEL succeeded."

# Step 2: Update ECS Task Definition with the new image tag
print_message "34" "🚀 Updating ECS task definition for $SERVICE_LABEL..."
./scripts/update_ecs_task.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
  print_message "31" "❌ Update ECS task definition for $SERVICE_TYPE failed."
  exit 1
fi
print_message "32" "✅ ECS task definition for $SERVICE_LABEL updated."

# Step 3: Deploy the new version to ECS
print_message "34" "🚀 Initiating deployment for $SERVICE_LABEL..."
./scripts/deploy.sh $SERVICE_TYPE
if [ $? -ne 0 ]; then
  print_message "31" "❌ Deployment for $SERVICE_LABEL failed."
  exit 1
fi
print_message "32" "✅ Deployment for $SERVICE_LABEL initiated."

print_message "35" "🎉 $SERVICE_LABEL service has been built, task definition updated, and deployment initiated."
print_message "35" "🎉 Check the ECS console for the status of the deployment."
