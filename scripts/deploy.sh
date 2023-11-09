#!/bin/bash

# Check if the service type argument is provided and is valid
if [[ "$#" -ne 1 ]] || { [ "$1" != "http" ] && [ "$1" != "grpc" ]; }; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

SERVICE_TYPE=$1
CLUSTER_NAME="grpchat-grpc-cluster"
PROJECT_ROOT_DIR="$(dirname "$0")/.."
cd "$PROJECT_ROOT_DIR" || { echo "Could not find project root directory"; exit 1; }

TASK_DEF_PATH="$PROJECT_ROOT_DIR/deploy/ecs/ecs-task-definition-${SERVICE_TYPE}.json"

# Register the new task definition to create a new revision
NEW_TASK_DEF_ARN=$(aws ecs register-task-definition \
  --cli-input-json file://"$TASK_DEF_PATH" \
  --query 'taskDefinition.taskDefinitionArn' \
  --output text)

# Check if the task definition was registered successfully
if [ -z "$NEW_TASK_DEF_ARN" ]; then
    echo "Failed to register new task definition for $SERVICE_TYPE."
    exit 1
fi

echo "Registered new task definition: $NEW_TASK_DEF_ARN"

# Update the ECS service to use the new task definition revision
aws ecs update-service \
  --cluster "$CLUSTER_NAME" \
  --service "grpchat-${SERVICE_TYPE}" \
  --task-definition "$NEW_TASK_DEF_ARN" \
  --force-new-deployment

if [ $? -eq 0 ]; then
    echo "Deployment for $SERVICE_TYPE service initiated with task definition $NEW_TASK_DEF_ARN."
else
    echo "Failed to initiate deployment for $SERVICE_TYPE service."
    exit 1
fi
