# gRPChat

## Overview

Simple gRPC chat app with Go

## Prerequisites

- Go 1.17 or higher
- Protocol Buffers compiler (`protoc`)
- gRPC

## Directory Structure

```
.
├── cmd
│   └── server
│       └── main.go
├── internal
│   └── chat
│       └── chat.go
├── proto
│   └── chat.proto
├── go.mod
└── go.sum
```

## Getting Started

### Clone the Repository

```bash
git clone https://github.com/felguerez/grpchat.git
cd grpchat
```

### Generate gRPC Code

Navigate to the `proto` directory and run:

```bash
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./proto/chat.proto
```

### Install Dependencies

Run the following command to download and install the required Go modules:

```bash
go mod tidy
```

### Run the Server

Navigate to the `cmd/server` directory and run:

```bash
go run main.go
```

You should see a log message indicating that the server is running on port 50051.

### Testing with `grpcurl`

You can use `grpcurl` to manually test the gRPC endpoints:

```bash
grpcurl -plaintext -d '{"content": "Hello"}' localhost:50051 chat.ChatService/SendMessage
```