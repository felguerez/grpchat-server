package main

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/chat"
	"github.com/felguerez/grpchat/internal/handlers"
	chatpb "github.com/felguerez/grpchat/proto"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
)

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	auth.InitializeSpotifyOauthConfig(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"), os.Getenv("SPOTIFY_REDIRECT_CALLBACK_URL"), []string{"user-read-email"})

	http.HandleFunc("/login", handlers.HandleLogin)
	http.HandleFunc("/callback", handlers.HandleCallback)

	httpPort := "8080"
	go func() {
		log.Println(fmt.Sprintf("HTTP server is running on port :%s", httpPort))
		if err := http.ListenAndServe(":"+httpPort, nil); err != nil {
			log.Fatalf("Failed to start HTTP server: %v", err)
		}
	}()

	grpcServer := grpc.NewServer()
	chatpb.RegisterChatServiceServer(grpcServer, &chat.Server{})
	reflection.Register(grpcServer)

	grpcPort := "50051"
	lis, err := net.Listen("tcp", ":"+grpcPort)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	log.Println(fmt.Sprintf("gRPC server is running on port :%s", grpcPort))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
