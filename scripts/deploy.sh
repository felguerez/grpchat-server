#!/bin/bash

CLUSTER_NAME="grpchat-grpc-cluster"
HTTP_SERVICE="grpchat-http"
GRPC_SERVICE="grpchat-grpc"
ECR_REPO="413025517373.dkr.ecr.us-east-1.amazonaws.com"
PROJECT_ROOT_DIR="$(dirname "$0")/.."  # This gets the parent directory of the script location

echo "Project directory is $PROJECT_ROOT_DIR"

# login
aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin $ECR_REPO

cd "$PROJECT_ROOT_DIR" || ..

# Build, tag, and push for grpchat-http service
docker build -t $HTTP_SERVICE -f Dockerfile.http .
docker tag $HTTP_SERVICE:latest $ECR_REPO/$HTTP_SERVICE:latest
docker push $ECR_REPO/$HTTP_SERVICE:latest

# Build, tag, and push for grpchat-grpc service
docker build -t ${GRPC_SERVICE} -f Dockerfile.grpc .
docker tag ${GRPC_SERVICE}:latest ${ECR_REPO}/${GRPC_SERVICE}:latest
docker push ${ECR_REPO}/${GRPC_SERVICE}:latest

echo "Deployment complete."

aws ecs update-service --cluster $CLUSTER_NAME --service $HTTP_SERVICE --force-new-deployment
aws ecs update-service --cluster $CLUSTER_NAME --service $GRPC_SERVICE --force-new-deployment
