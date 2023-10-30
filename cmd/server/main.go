package main

import (
	"fmt"
	"github.com/felguerez/grpchat/internal/auth"
	"github.com/felguerez/grpchat/internal/chat"
	"github.com/felguerez/grpchat/internal/handlers"
	"github.com/felguerez/grpchat/internal/wschat"
	chatpb "github.com/felguerez/grpchat/proto"
	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
	"os"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		logger, _ := zap.NewProduction()
		defer logger.Sync()

		logger.Info("Incoming request",
			zap.String("method", r.Method),
			zap.String("url", r.URL.String()),
		)
		next.ServeHTTP(w, r)
	})
}
func main() {
	dir, _ := os.Getwd()
	fmt.Println("Current directory is:", dir)
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	logger.Info(fmt.Sprintf("Callback URL is %s", os.Getenv("SPOTIFY_REDIRECT_CALLBACK_URL")))

	auth.InitializeSpotifyOauthConfig(os.Getenv("SPOTIFY_CLIENT_ID"), os.Getenv("SPOTIFY_CLIENT_SECRET"), os.Getenv("SPOTIFY_REDIRECT_CALLBACK_URL"), []string{"user-read-email"})
	apiMux := http.NewServeMux()
	apiMux.HandleFunc("/messages", handlers.HandleGetAllMessages) // @TODO: implement grpc endpoint for this data

	http.HandleFunc("/", handlers.Root)
	http.Handle("/api/", LoggingMiddleware(http.StripPrefix("/api", handlers.RequireAuthorizationToken(apiMux))))
	http.Handle("/login", LoggingMiddleware(handlers.HandleLogin(logger)))
	http.Handle("/callback", LoggingMiddleware(handlers.HandleCallback(logger)))
	http.HandleFunc("/api/conversations/", wschat.InitializeWebSocket())

	httpPort := "8080"
	go func() {
		logger.Info(fmt.Sprintf("HTTP server is running on port :%s", httpPort))
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
	logger.Info(fmt.Sprintf("gRPC server is running on port :%s", grpcPort))

	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
