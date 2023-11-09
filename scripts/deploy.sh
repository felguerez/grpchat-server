#!/bin/bash

# Assign the first argument to a variable
SERVICE_TYPE=$1

# Check if the service type argument is provided and is valid
if [[ -z "$SERVICE_TYPE" ]] || { [ "$SERVICE_TYPE" != "http" ] && [ "$SERVICE_TYPE" != "grpc" ]; }; then
    echo "Usage: $0 <http|grpc>"
    exit 1
fi

# Set variables
CLUSTER_NAME="grpchat-grpc-cluster"

# Deploy the specified service
if [ "$SERVICE_TYPE" = "http" ]; then
    SERVICE="grpchat-http"
else
    SERVICE="grpchat-grpc"
fi

aws ecs update-service --cluster $CLUSTER_NAME --service $SERVICE --force-new-deployment

echo "Deployment for $SERVICE_TYPE service initiated."
