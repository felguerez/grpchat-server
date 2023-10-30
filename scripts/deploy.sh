#!/bin/bash

# Login to ECR
$(aws ecr get-login --no-include-email --region us-east-1)

# Build, tag, and push for grpchat-http service
docker build -t grpchat-http -f Dockerfile.http .
docker tag grpchat-http:latest 413025517373.dkr.ecr.us-east-1.amazonaws.com/grpchat-http:latest
docker push 413025517373.dkr.ecr.us-east-1.amazonaws.com/grpchat-http:latest

# Build, tag, and push for grpchat-grpc service
docker build -t grpchat-grpc -f Dockerfile.grpc .
docker tag grpchat-grpc:latest 413025517373.dkr.ecr.us-east-1.amazonaws.com/grpchat-grpc:latest
docker push 413025517373.dkr.ecr.us-east-1.amazonaws.com/grpchat-grpc:latest

echo "Deployment complete."
