package main

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/chat"
	chatpb "github.com/felguerez/grpchat/proto"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	port := "50051"
	lis, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	s := grpc.NewServer()
	chatpb.RegisterChatServiceServer(s, &chat.Server{})
	reflection.Register(s)

	log.Println(fmt.Sprintf("Server is running on port :%s", port))

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
